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

var _ framework.ScorePlugin = &FGDScorePlugin{}



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
	fmt.Printf("DEBUG FRA, plugin.fgd_score.Score() => Scoring a node w.r.t. a pod!\n")
	
	// Step 1 - Check if the considered pod does not request any resource -- in this case we return the maximum score (100) and a success status.
	// "PodRequestsAndLimits()" returns a dictionary of all defined resources summed up for all containers of the pod. 
	// If pod overhead is non-nil, the pod overhead is added to the total container resource requests and to the 
	// total container limits which have a non-zero quantity.
	if podReq, _ := resourcehelper.PodRequestsAndLimits(p); len(podReq) == 0 {
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


	// Step 3 - Retrieve the resources requested by the pod, and check if the node is suitable for the pod, i.e., the node has enough resources to accomodate 
	// the pod, and the GPU type requested by the pod is present on the node.
	podRes := utils.GetPodResource(p)
	if !utils.IsNodeAccessibleToPod(nodeRes, podRes) {
		return framework.MinNodeScore, framework.NewStatus(framework.Error, fmt.Sprintf("Node (%s) %s does not match GPU type request of pod %s\n", nodeName, nodeRes.Repr(), podRes.Repr()))
	}


	// Step 4 - 
	score, _ := calculateGpuShareFragExtendScore(nodeRes, podRes, plugin.typicalPods)
	return score, framework.NewStatus(framework.Success)
}

func (plugin *FGDScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}


// This function computes 
func calculateGpuShareFragExtendScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, typicalPods *simontype.TargetPodList) (score int64, gpuId string) {
	nodeGpuShareFragScore := utils.NodeGpuShareFragAmountScore(nodeRes, *typicalPods)
	
	// Case 1 - the pod requests a fraction of the resources of a single GPU.
	if podRes.GpuNumber == 1 && podRes.MilliGpu < gpushareutils.MILLI {
		
		// Initially set the score to 0 -- this will be the score assigned to nodes that cannot accomodate the pod.
		score, gpuId = 0, ""
		
		// For each node in the cluster, we check how its GPU fragmentation changes by hypotetically assigning the considered pod to it.
		for i := 0; i < len(nodeRes.MilliGpuLeftList); i++ {
			
			// The node has enough GPU-shared resources to accomodate the pod.
			if nodeRes.MilliGpuLeftList[i] >= podRes.MilliGpu {
				// Simulate how the available resources on a node would change by scheduling the pod on it.
				newNodeRes := nodeRes.Copy()
				newNodeRes.MilliCpuLeft -= podRes.MilliCpu
				newNodeRes.MilliGpuLeftList[i] -= podRes.MilliGpu
				
				// Compute the fragmentation score with the updated resource availability.
				newNodeGpuShareFragScore := utils.NodeGpuShareFragAmountScore(newNodeRes, *typicalPods)
				
				
				fragScore := int64(sigmoid((nodeGpuShareFragScore-newNodeGpuShareFragScore)/1000) * float64(framework.MaxNodeScore))
				if gpuId == "" || fragScore > score {
					score = fragScore
					gpuId = strconv.Itoa(i)
				}
			}
		}
		return score, gpuId
		
	// Case 2 - the pod requests one or more entire GPUs.
	} else {
		newNodeRes, _ := nodeRes.Sub(podRes)
		newNodeGpuShareFragScore := utils.NodeGpuShareFragAmountScore(newNodeRes, *typicalPods)
		return int64(sigmoid((nodeGpuShareFragScore-newNodeGpuShareFragScore)/1000) * float64(framework.MaxNodeScore)), simontype.AllocateExclusiveGpuId(nodeRes, podRes)
	}
}


func allocateGpuIdBasedOnFGDScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, _ simontype.GpuPluginCfg, typicalPods *simontype.TargetPodList) (gpuId string) {
	_, gpuId = calculateGpuShareFragExtendScore(nodeRes, podRes, typicalPods)
	return gpuId
}
