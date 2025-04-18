package plugin

import (
	"context"
	"fmt"
	"math"
	"strconv"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	resourcehelper "k8s.io/kubectl/pkg/util/resource"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	simontype "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type"
	gpushareutils "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type/open-gpu-share/utils"
	"github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/utils"
)

type PWRScorePlugin struct {
	handle      framework.Handle
	typicalPods *simontype.TargetPodList
}

// TODO: All the methods and functions should be in place. Now we need to bind the plugin to the scheduler framework, in the right places of the simulator.
//       See FGD.

var _ framework.ScorePlugin = &PWRScorePlugin{} // This assignment is used at compile-time to check if the class implements the plugin interface.

// The function below allows to bind this plugin to the simulator.
// NOTE: typical pods should represent the target workload, i.e., pods passed via YAMLs before workload inflation.
// These are required to compute the cluster fragmentation.
func NewPWRScorePlugin(_ runtime.Object, handle framework.Handle, typicalPods *simontype.TargetPodList) (framework.Plugin, error) {
	log.Infof("DEBUG FRA, plugin.pwr_score.NewPWRScorePlugin() => Instantiating PWR plugin!\n")

	plugin := &PWRScorePlugin{
		handle:      handle,
		typicalPods: typicalPods,
	}

	allocateGpuIdFunc[plugin.Name()] = allocateGpuIdBasedOnPWRScore
	return plugin, nil
}

func (plugin *PWRScorePlugin) Name() string {
	return simontype.PWRScorePluginName
}

func (plugin *PWRScorePlugin) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	// DEBUG: print the gpu type(s) requested by the pod.
	pod_GPU_type := gpushareutils.GetGpuModelFromPodAnnotation(p)
	if pod_GPU_type == "" {
		if gpushareutils.GetGpuMilliFromPodAnnotation(p) > 0 {
			pod_GPU_type = "GENERIC"
		} else {
			pod_GPU_type = "NONE"
		}
	}
	log.Debugf("DEBUG FRA, plugin.pwr_score.Score() => Scoring node %s w.r.t. pod %s (requested GPU: %s)!\n",
		nodeName, p.Name, pod_GPU_type)

	// Step 1 - Check if the considered pod does not request any resource -- in this case we return the maximum score (100) and a success status.
	// "PodRequestsAndLimits()" returns a dictionary of all defined resources summed up for all containers of the pod.
	// If pod overhead is non-nil, the pod overhead is added to the total container resource requests and to the
	// total container limits which have a non-zero quantity.
	if podReq, _ := resourcehelper.PodRequestsAndLimits(p); len(podReq) == 0 {
		log.Debugf("DEBUG FRA, plugin.pwr_score.Score() => the pod does not request any resource!\n")
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

	// Step 3 - Retrieve the resources requested by the pod, and check if the currently considered node is suitable for the pod, i.e.,
	// the node has enough resources to accomodate
	// the pod, and the GPU type requested by the pod is present on the node.
	podRes := utils.GetPodResource(p)
	if !utils.IsNodeAccessibleToPod(nodeRes, podRes) {
		return framework.MinNodeScore, framework.NewStatus(framework.Error, fmt.Sprintf("Node (%s) %s does not match GPU type request of pod %s\n", nodeName, nodeRes.Repr(), podRes.Repr()))
	}

	log.Debugf("DEBUG FRA, plugin.pwr_score.Score() => Resources requested from pod: %+v\n", podRes)
	log.Debugf("DEBUG FRA, plugin.pwr_score.Score() => Resources offered by node: %+v\n", nodeRes)
	// log.Debugf("DEBUG FRA, plugin.pwr_score.Score() => typical pods %+v\n", plugin.typicalPods)

	// Step 4 - compute the score of a node w.r.t. the considered pod.
	//			In this case, the score is calculated based on how much the GPU fragmentation of a node would change IF we hypotetically
	//		    schedule the pod on it -- the more the increase, the worst the score.
	score, _ := calculatePWRShareExtendScore(nodeRes, podRes, plugin.typicalPods)
	return score, framework.NewStatus(framework.Success)
}

// Here we need to return the struct itself in order to use NormalizeScore.
func (plugin *PWRScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return plugin
}

func (p *PWRScorePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	log.Debugf("DEBUG FRA, plugin.pwr_score.NormalizeScore() => Normalizing scores!\n")

	// Find the minimum score, as the maximum score is known to be 0
	minScore := scores[0].Score
	maxScore := minScore
	for _, score := range scores {
		if score.Score < minScore {
			minScore = score.Score
		}
		if score.Score > maxScore {
			maxScore = score.Score
		}
	}

	// Case where all the scores are equal: set them to 100 and return.
	if minScore == maxScore {
		log.Debugf("DEBUG FRA, plugin.pwr_score.NormalizeScore(): all the scores are equal.\n")

		for i, _ := range scores {
			scores[i].Score = framework.MaxNodeScore
			log.Debugf("DEBUG FRA, plugin.pwr_score.NormalizeScore(): normalized score for node %s: %d\n", scores[i].Name, scores[i].Score)
		}

		return framework.NewStatus(framework.Success)
	}

	// Normalize the scores to the range [0, 100].
	for i, _ := range scores {
		// Normalization formula: normalized_score = (score - minScore) / (0 - minScore) * 100
		scores[i].Score = (scores[i].Score - minScore) * framework.MaxNodeScore / (maxScore - minScore)
		log.Debugf("DEBUG FRA, plugin.pwr_score.NormalizeScore(): normalized score for node %s: %d\n", scores[i].Name, scores[i].Score)
	}

	return framework.NewStatus(framework.Success)
}

// This function computes the score of a node w.r.t. an unscheduled pod. This is done by hypotetically scheduling the pod on the node,
// and then measure how much the node's power consumption increases.
func calculatePWRShareExtendScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, _ *simontype.TargetPodList) (score int64, gpuId string) {
	// Compute the node's current power consumption.
	old_CPU_energy, old_GPU_energy := nodeRes.GetEnergyConsumptionNode()
	old_node_energy := old_CPU_energy + old_GPU_energy

	// Case 1 - the pod requests a fraction of the resources of a single GPU.
	if podRes.GpuNumber == 1 && podRes.MilliGpu < gpushareutils.MILLI {

		// For each GPU in the node, check how the node power consumption would change by hypotetically assigning the considered pod to it.
		// NOTE: for now, we are assuming that a GPU consumes max power even if it is minimally used.
		score, gpuId = math.MinInt64, ""
		// minMilliLeft := int64(gpushareutils.MILLI)
		for i := 0; i < len(nodeRes.MilliGpuLeftList); i++ {

			// The considered GPU within the node has enough GPU-shared resources to accomodate the pod.
			if nodeRes.MilliGpuLeftList[i] >= podRes.MilliGpu {
				// Simulate how the available resources on a node would change by scheduling the pod onto a specific node's GPU.
				newNodeRes := nodeRes.Copy()
				newNodeRes.MilliCpuLeft -= podRes.MilliCpu
				newNodeRes.MilliGpuLeftList[i] -= podRes.MilliGpu

				// Compute the node's hypotetical increase in power consumption.
				new_CPU_energy, new_GPU_energy := newNodeRes.GetEnergyConsumptionNode()
				new_node_energy := new_CPU_energy + new_GPU_energy

				// Compute the node's score according to the increase in power consumption that we would have by using the i-th GPU.
				pwrScore := int64(old_node_energy - new_node_energy)
				log.Debugf("DEBUG FRA, plugin.pwr_score.calculatePWRShareExtendScore(): Scoring node %s, GPU %d, with sharing-GPU pod: %d\n",
					nodeRes.NodeName, i, pwrScore)

				// ### Update the node's best score ### //
				// Case 1 - this is the first GPU within the node that can accomodate the pod.
				if gpuId == "" {
					// minMilliLeft = nodeRes.MilliGpuLeftList[i]
					score = pwrScore
					gpuId = strconv.Itoa(i)
				} else {
					// Case 2 - we have found a GPU that is equivalent in terms of power consumption to the best one, but scheduling the pod
					//          on this GPU  ...
					/*if (pwrScore == score) && (minMilliLeft < nodeRes.MilliGpuLeftList[i]) {
						minMilliLeft = nodeRes.MilliGpuLeftList[i]
						gpuId = strconv.Itoa(i)
					}*/

					// Case 3 - we have found a better GPU than the previous one, i.e., by allocating the pod on this GPU,
					//          we consume less energy than the previously found solution.
					if pwrScore > score {
						// minMilliLeft = nodeRes.MilliGpuLeftList[i]
						score = pwrScore
						gpuId = strconv.Itoa(i)
					}
				}
			}
		}

		log.Debugf("DEBUG FRA, plugin.pwr_score.calculatePWRShareExtendScore(): Final score for node %s: selected GPU %s, score %d\n",
			nodeRes.NodeName, gpuId, score)
		return score, gpuId

		// Case 2 - the pod requests no (CPU only), or exactly one, or multiple GPUs.
	} else {
		// Subtract the node's resources that would be taken by the pod once scheduled on it.
		newNodeRes, _ := nodeRes.Sub(podRes)

		// Compute the node's power consumption, with the updated resource availability.
		new_CPU_energy, new_GPU_energy := newNodeRes.GetEnergyConsumptionNode()
		new_node_energy := new_CPU_energy + new_GPU_energy

		pwrScore := int64(old_node_energy - new_node_energy)
		log.Debugf("DEBUG FRA, plugin.pwr_score.calculatePWRShareFragExtendScore(): Scoring node %s with CPU-only or multi-GPU pod: %d\n",
			nodeRes.NodeName, pwrScore)

		return pwrScore, simontype.AllocateExclusiveGpuId(nodeRes, podRes)
	}
}

// This function selects the best GPU(s) found in a given node. It essentially re-executes the allocateGpuIdBasedOnPWRScore function
// executed within Score(), but it considers only the best GPU(s) for a pod found in a node and ignores the computed score.
func allocateGpuIdBasedOnPWRScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, _ simontype.GpuPluginCfg, typicalPods *simontype.TargetPodList) (gpuId string) {
	log.Debugf("DEBUG FRA, plugin.pwr_score.allocateGpuIdBasedOnPWRScore() => Scoring node %s w.r.t. pod!\n", nodeRes.NodeName)
	_, gpuId = calculatePWRShareExtendScore(nodeRes, podRes, typicalPods)
	return gpuId
}
