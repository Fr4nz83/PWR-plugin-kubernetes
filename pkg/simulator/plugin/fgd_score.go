package plugin

import (
	"context"
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	simontype "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type"
	gpushareutils "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type/open-gpu-share/utils"
	"github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/utils"
)

type FGDScorePlugin struct {
	handle      framework.Handle
	typicalPods *simontype.TargetPodList
}

var _ framework.ScorePlugin = &FGDScorePlugin{} // This assignment is used at compile-time to check if the class implements the plugin interface.

// NOTE: typical pods should represent the target workload, i.e., pods passed via YAMLs before workload inflation.
// These are required to compute the cluster fragmentation.
func NewFGDScorePlugin(_ runtime.Object, handle framework.Handle, typicalPods *simontype.TargetPodList) (framework.Plugin, error) {
	fmt.Printf("DEBUG FRA, plugin.fgd_score.NewFGDScorePlugin() => Instantiating FGD plugin!\n")

	plugin := &FGDScorePlugin{
		handle:      handle,
		typicalPods: typicalPods,
	}

	allocateGpuIdFunc[plugin.Name()] = allocateGpuIdBasedOnFGDScore
	return plugin, nil
}

func (plugin *FGDScorePlugin) Name() string {
	return simontype.FGDScorePluginName
}

func (plugin *FGDScorePlugin) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	fmt.Printf("DEBUG FRA, plugin.fgd_score.Score() => Scoring node %s w.r.t. pod %s!\n", nodeName, p.Name)

	// Step 1 - Check if the considered pod does not request any resource -- in this case we return the maximum score (100) and a success status.
	// "PodRequestsAndLimits()" returns a dictionary of all defined resources summed up for all containers of the pod.
	// If pod overhead is non-nil, the pod overhead is added to the total container resource requests and to the
	// total container limits which have a non-zero quantity.
	if podReq, _ := resourcehelper.PodRequestsAndLimits(p); len(podReq) == 0 {
		fmt.Printf("DEBUG FRA, plugin.fgd_score.Score() => the pod does not request any resource!\n")
		return framework.MaxNodeScore, framework.NewStatus(framework.Success)
	}

	// Step 2 - Retrieves the resources of the node specified by nodeName.
	nodeResPtr := utils.GetNodeResourceViaHandleAndName(plugin.handle, nodeName)
	// Check if "GetNodeResourceViaHandleAndName" failed to retrieve the node's resources, possibly due to the node not being found or some other error.
	// In this case, we return the minimum node score and an error status.
	if nodeResPtr == nil {
		return framework.MinNodeScore, framework.NewStatus(framework.Error, fmt.Sprintf("failed to get nodeRes(%s)\n", nodeName))
	}
	nodeRes := *nodeResPtr

	// Step 3 - Retrieve the resources requested by the pod, and check if the node is suitable for the pod, i.e., the node has the GPU type requested
	//		    by the pod (if that's the case).
	podRes := utils.GetPodResource(p)
	if !utils.IsNodeAccessibleToPod(nodeRes, podRes) {
		return framework.MinNodeScore, framework.NewStatus(framework.Error, fmt.Sprintf("Node (%s) %s does not match GPU/CPU type request of pod %s\n", nodeName, nodeRes.Repr(), podRes.Repr()))
	}

	fmt.Printf("DEBUG FRA, plugin.fgd_score.Score() => Resources requested from pod: %+v\n", podRes)
	fmt.Printf("DEBUG FRA, plugin.fgd_score.Score() => Resources offered by node: %+v\n", nodeRes)
	// fmt.Printf("DEBUG FRA, plugin.fgd_score.Score() => typical pods %+v\n", plugin.typicalPods)

	// Step 4 - compute the score of a node w.r.t. the considered pod.
	//			In this case, the score is calculated based on how much the GPU fragmentation of a node would change IF we hypotetically
	//		    schedule the pod on it -- the more the increase, the worst the score.
	score, _ := calculateGpuShareFragExtendScore(nodeRes, podRes, plugin.typicalPods)
	return score, framework.NewStatus(framework.Success)
}

func (plugin *FGDScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// This function computes the score of a node w.r.t. an unscheduled pod. This is done by hypotetically scheduling the pod on the node,
// and then measure how much the node's fragmentation changes w.r.t. the target workload.newNodeGpuShareFragScore
func calculateGpuShareFragExtendScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, typicalPods *simontype.TargetPodList) (score int64, gpuId string) {
	// Compute the node's current fragmentation.
	nodeGpuShareFragScore := utils.NodeGpuShareFragAmountScore(nodeRes, *typicalPods)

	// Case 1 - the pod requests a fraction of the resources of a single GPU.
	if podRes.GpuNumber == 1 && podRes.MilliGpu < gpushareutils.MILLI {

		// Initially set the score to 0 -- this will be the score assigned to nodes that cannot accomodate the pod.
		score, gpuId = 0, ""

		// For each GPU in the node, check how the GPU's fragmentation would change by hypotetically assigning the considered pod to it.
		// The loop below scan the set of GPUs within the node.
		for i := 0; i < len(nodeRes.MilliGpuLeftList); i++ {

			// The considered GPU within the node has enough GPU-shared resources to accomodate the pod.
			if nodeRes.MilliGpuLeftList[i] >= podRes.MilliGpu {
				// Simulate how the available resources on a node would change by scheduling the pod on it.
				newNodeRes := nodeRes.Copy()
				newNodeRes.MilliCpuLeft -= podRes.MilliCpu
				newNodeRes.MilliGpuLeftList[i] -= podRes.MilliGpu

				// Compute the node's fragmentation, with the updated resource availability, considering the target workload.
				newNodeGpuShareFragScore := utils.NodeGpuShareFragAmountScore(newNodeRes, *typicalPods)

				// Compute the difference between the old fragmentation and the new fragmentation -- the higher the result, the better placing the
				// pod on a GPU is. Then, divide the difference by 1000 as the two values are expressed on millisimed resources.
				// Finally, apply the sigmoid to get a value in [0,1], and then multiply this value for the maximum admissible score.
				fragScore := int64(sigmoid((nodeGpuShareFragScore-newNodeGpuShareFragScore)/1000) * float64(framework.MaxNodeScore))
				fmt.Printf("DEBUG FRA, plugin.fgd_score.calculateGpuShareFragExtendScore(): Scoring node %s, GPU %d, with sharing-GPU pod: %d\n", nodeRes.NodeName, i, fragScore)
				if gpuId == "" || fragScore > score {
					score = fragScore
					gpuId = strconv.Itoa(i)
				}
			}
		}
		return score, gpuId

		// Case 2 - the pod requests one or more entire GPUs.
	} else {
		// Subtract the node's resources that would be taken by the pod once scheduled on it.
		newNodeRes, _ := nodeRes.Sub(podRes)
		// Compute the node's fragmentation, with the updated resource availability, considering the target workload.
		newNodeGpuShareFragScore := utils.NodeGpuShareFragAmountScore(newNodeRes, *typicalPods)

		fragScore := int64(sigmoid((nodeGpuShareFragScore-newNodeGpuShareFragScore)/1000) * float64(framework.MaxNodeScore))
		fmt.Printf("DEBUG FRA, plugin.fgd_score.calculateGpuShareFragExtendScore(): Scoring node %s, with no-CPU or multi-GPU pod: %d\n", nodeRes.NodeName, fragScore)

		return fragScore, simontype.AllocateExclusiveGpuId(nodeRes, podRes)
	}
}

func allocateGpuIdBasedOnFGDScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, _ simontype.GpuPluginCfg, typicalPods *simontype.TargetPodList) (gpuId string) {
	_, gpuId = calculateGpuShareFragExtendScore(nodeRes, podRes, typicalPods)
	return gpuId
}
