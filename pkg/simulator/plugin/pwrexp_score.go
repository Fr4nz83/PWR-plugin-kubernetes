package plugin

import (
	"context"
	"fmt"
	"math"
	"strconv"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	simontype "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type"
	gpushareutils "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type/open-gpu-share/utils"
	utils "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/utils"
)

type PWREXPScorePlugin struct {
	handle      framework.Handle
	typicalPods *simontype.TargetPodList
}

// TODO: All the methods and functions should be in place. Now we need to bind the plugin to the scheduler framework, in the right places of the simulator.
//       See FGD.

var _ framework.ScorePlugin = &PWREXPScorePlugin{} // This assignment is used at compile-time to check if the class implements the plugin interface.

// The function below allows to bind this plugin to the simulator.
// NOTE: typical pods should represent the target workload, i.e., pods passed via YAMLs before workload inflation.
// These are required to compute the cluster fragmentation.
func NewPWREXPScorePlugin(_ runtime.Object, handle framework.Handle, typicalPods *simontype.TargetPodList) (framework.Plugin, error) {
	log.Infof("DEBUG FRA, plugin.pwrexp_score.NewPWREXPScorePlugin() => Instantiating PWREXP plugin!\n")

	plugin := &PWREXPScorePlugin{
		handle:      handle,
		typicalPods: typicalPods,
	}

	allocateGpuIdFunc[plugin.Name()] = allocateGpuIdBasedOnPWREXPScore
	return plugin, nil
}

func (plugin *PWREXPScorePlugin) Name() string {
	return simontype.PWREXPScorePluginName
}

func (plugin *PWREXPScorePlugin) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	// DEBUG: print the gpu type(s) requested by the pod.
	pod_GPU_type := gpushareutils.GetGpuModelFromPodAnnotation(p)
	if pod_GPU_type == "" {
		if gpushareutils.GetGpuMilliFromPodAnnotation(p) > 0 {
			pod_GPU_type = "GENERIC"
		} else {
			pod_GPU_type = "NONE"
		}
	}
	log.Debugf("DEBUG FRA, plugin.pwrexp_score.Score() => Scoring node %s w.r.t. pod %s (requested GPU: %s)!\n",
		nodeName, p.Name, pod_GPU_type)

	// Step 1 - Check if the considered pod does not request any resource -- in this case we return the maximum score (100) and a success status.
	// "PodRequestsAndLimits()" returns a dictionary of all defined resources summed up for all containers of the pod.
	// If pod overhead is non-nil, the pod overhead is added to the total container resource requests and to the
	// total container limits which have a non-zero quantity.
	// NOTE: we deactivated this check, and handle this case in the generic code at step 4.
	// if podReq, _ := resourcehelper.PodRequestsAndLimits(p); len(podReq) == 0 {
	// 	log.Debugf("DEBUG FRA, plugin.pwrexp_score.Score() => the pod does not request any resource!\n")
	//	return 0, framework.NewStatus(framework.Success)
	// }

	// Step 2 - Retrieves the resources of the node specified by nodeName.
	nodeResPtr := utils.GetNodeResourceViaHandleAndName(plugin.handle, nodeName)
	// Check if "GetNodeResourceViaHandleAndName" failed to retrieve the node's resources, possibly due to the node not being found or some other error.
	// In this case, we return the minimum node score and an error status.
	// NOTE: in a simulation, we should never enter the if below. In any case, return the largest negative int64, which represents the largest
	// 		 possible increase in expected power consumption (and thus maximally penalize this node).
	if nodeResPtr == nil {
		return int64(math.MinInt64), framework.NewStatus(framework.Error, fmt.Sprintf("failed to get nodeRes(%s)\n", nodeName))
	}
	nodeRes := *nodeResPtr

	// Step 3 - Retrieve the resources requested by the pod, and check if the currently considered node is suitable for the pod, i.e.,
	// the node has enough resources to accomodate the pod, and the GPU type requested by the pod is present on the node.
	// NOTE: in theory we should never enter this if block, as the Filter plugin removes the nodes that fall in this case.
	//       In any case, we return the largest negative int64, which represents the largest possible increase in expected power consumption.
	podRes := utils.GetPodResource(p)
	if !utils.IsNodeAccessibleToPod(nodeRes, podRes) {
		return int64(math.MinInt64), framework.NewStatus(framework.Error, fmt.Sprintf("Node (%s) %s does not match GPU type request of pod %s\n", nodeName, nodeRes.Repr(), podRes.Repr()))
	}

	log.Debugf("DEBUG FRA, plugin.pwrexp_score.Score() => Resources requested from pod: %+v\n", podRes)
	log.Debugf("DEBUG FRA, plugin.pwrexp_score.Score() => Resources offered by node: %+v\n", nodeRes)
	// log.Debugf("DEBUG FRA, plugin.pwrexp_score.Score() => typical pods %+v\n", plugin.typicalPods)

	// Step 4 - compute the score of a node w.r.t. the considered pod.
	//			In this case, the score is calculated based on how much the GPU fragmentation of a node would change IF we hypotetically
	//		    schedule the pod on it -- the more the increase, the worst the score.
	score, _ := calculatePWREXPShareExtendScore(nodeRes, podRes, plugin.typicalPods)
	return score, framework.NewStatus(framework.Success)
}

// Here we need to return the struct itself in order to use NormalizeScore.
func (plugin *PWREXPScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return plugin
}

func (p *PWREXPScorePlugin) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	log.Debugf("DEBUG FRA, plugin.pwrexp_score.NormalizeScore() => Normalizing scores!\n")

	// Find the minimum and maximum scores. A negative score indicates an actual decrease in expected power consumption variation of a node,
	// and are thus better than positive scores.
	minScore, maxScore := scores[0].Score, scores[0].Score
	for i := range scores {
		if scores[i].Score < minScore {
			minScore = scores[i].Score
		}
		if scores[i].Score > maxScore {
			maxScore = scores[i].Score
		}
	}

	// Case where all the scores are equal: set them to 100 and return.
	if minScore != maxScore {
		// Normalize the scores to the range [0, 100].
		for i := range scores {
			// Normalization formula: normalized_score = (score - minScore) / (maxScore - minScore) * 100
			scores[i].Score = (scores[i].Score - minScore) * framework.MaxNodeScore / (maxScore - minScore)
			log.Debugf("DEBUG FRA, plugin.pwrexp_score.NormalizeScore(): normalized score for node %s: %d\n", scores[i].Name, scores[i].Score)
		}
	} else {
		log.Debugf("DEBUG FRA, plugin.pwrexp_score.NormalizeScore(): all the scores are equal, set everything to 100.\n")
		for i := range scores {
			scores[i].Score = framework.MaxNodeScore
			log.Debugf("DEBUG FRA, plugin.pwrexp_score.NormalizeScore(): normalized score for node %s: %d\n", scores[i].Name, scores[i].Score)
		}
	}

	return framework.NewStatus(framework.Success)
}

// This function checks wheter a given node can host a given pod.
func isPodAllocatableToNode(nodeRes simontype.NodeResource, podRes simontype.PodResource) bool {

	test1 := nodeRes.MilliCpuLeft >= podRes.MilliCpu      // Check if the node has enough CPU resources for the POD.
	test2 := utils.IsNodeAccessibleToPod(nodeRes, podRes) // Check if the node has the GPU type requested by the pod (if that's the case).

	// Check if the node has enough GPU resources for the POD.
	// NOTE: 'utils.CanNodeHostPodOnGpuMemory' works correctly only with pods that require GPU resources -- with no-GPU pods it erroneously
	//       returns false. Thus, as a workaround we handle the non-GPU case in the first condition plus the OR.
	test3 := (podRes.GpuNumber == 0) || utils.CanNodeHostPodOnGpuMemory(nodeRes, podRes)

	return test1 && test2 && test3
}

// This function computes the score of a node w.r.t. an unscheduled pod. This is done by hypotetically scheduling the pod on the node,
// and then measure how much the node's power consumption increases.
func calculatePWREXPShareExtendScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, typicalPods *simontype.TargetPodList) (score int64, gpuId string) {

	// Compute the node's expected power consumption increase before hypotetically assigning podRes to the node.
	// To this end, use the typical pods.
	curr_exp_pwr_inc := CalcExpPWRVarNode(nodeRes, typicalPods)

	// Case 1 - the pod is GPU-share, i.e., it requests a fraction of the resources of a single GPU.
	if podRes.IsGpuShare() {
		// For each GPU in the node, check how the node power consumption would change by hypotetically assigning the considered pod to it.
		// NOTE: for now, we are assuming that a GPU consumes max power even if it is minimally used.
		score, gpuId = math.MinInt64, ""
		for i := range nodeRes.MilliGpuLeftList {

			// The considered GPU within the node has enough GPU-shared resources to accomodate the pod.
			if nodeRes.MilliGpuLeftList[i] >= podRes.MilliGpu {
				// Simulate how the available resources on a node would change by scheduling the pod onto a specific node's GPU.
				newNodeRes := nodeRes.Copy()
				newNodeRes.MilliCpuLeft -= podRes.MilliCpu
				newNodeRes.MilliGpuLeftList[i] -= podRes.MilliGpu

				// Now compute the expected variation in power consumption using the typical pods, with the pod hypotetically scheduled to the node.
				hyp_exp_pwr_inc := CalcExpPWRVarNode(newNodeRes, typicalPods)
				pwrScore := int64(curr_exp_pwr_inc - hyp_exp_pwr_inc)
				log.Debugf("DEBUG FRA, plugin.pwrexp_score.calculatePWREXPShareExtendScore(): Scoring node %s, GPU %d, with sharing-GPU pod: %d\n",
					nodeRes.NodeName, i, pwrScore)

				// ### Update the node's best score ### //
				// Case 1 - this is the first GPU within the node that can accomodate the pod.
				if gpuId == "" {
					score = pwrScore
					gpuId = strconv.Itoa(i)
				} else {
					// Case 2 - we have found a better GPU than the previous one, i.e., by allocating the pod on this GPU,
					//          we consume less energy than the previously found solution.
					if pwrScore > score {
						score = pwrScore
						gpuId = strconv.Itoa(i)
					}
				}
			}
		}

		log.Debugf("DEBUG FRA, plugin.pwrexp_score.calculatePWREXPShareExtendScore(): Final score for node %s: selected GPU %s, score %d\n",
			nodeRes.NodeName, gpuId, score)
		return score, gpuId

		// Case 2 - the pod requests no GPU (CPU only), or exactly one GPU, or multiple GPUs.
	} else {
		// Subtract the node's resources that would be taken by the pod if scheduled on it.
		newNodeRes, _ := nodeRes.Sub(podRes)

		// Compute the expected variation in power consumption with the pod hypotetically allocated onto the node.
		hyp_exp_pwr_inc := CalcExpPWRVarNode(newNodeRes, typicalPods)
		// And now compute the difference before vs after hypotetically allocating the pod onto the node.
		// NOTE: the larger the values, the better a node.
		score := int64(curr_exp_pwr_inc - hyp_exp_pwr_inc)
		log.Debugf("DEBUG FRA, plugin.pwrexp_score.calculatePWREXPShareExtendScore(): Scoring node %s with CPU-only or multi-GPU pod: %d\n",
			nodeRes.NodeName, score)

		return score, simontype.AllocateExclusiveGpuId(nodeRes, podRes)
	}
}

// This function computes the expected power variation onto the node represented by 'nodeRes' when considering the pods
// of a typical workload represented by 'typicalPods'.
func CalcExpPWRVarNode(nodeRes simontype.NodeResource, typicalPods *simontype.TargetPodList) (expPwrVariation float64) {

	// Local type.
	type Pair struct {
		pwr_inc float64
		prob    float64
	}

	// Step 1 - Compute the estimated power consumption of the node before hypotetically scheduling the various typical pods on it.
	old_CPU_power, old_GPU_power := nodeRes.GetEnergyConsumptionNode()
	old_node_power := old_CPU_power + old_GPU_power

	// Step 2 - Scan the pods in the target workload, and save the increase in power consumption they would entail.
	skipped := false
	var list_allocatable_pods []Pair
	for _, pod := range *typicalPods {

		podFreq := pod.Percentage       // Retrieve this typical pod's popularity, i.e., the probability that a pod with this resource profile occurs.
		podRes := pod.TargetPodResource // Retrieve the resources requested by this typical pod.

		// Check if the considered typical pod in the target workload has a probability that makes sense.
		// If not (shouldn't happen!), ignore the pod.
		if podFreq < 0 || podFreq > 1 {
			log.Errorf("pod %v has bad freq: %f\n", podRes, podFreq)
			skipped = true
			continue
		}

		// Check if the node can host this typical pod; if not, skip to the next one.
		if !isPodAllocatableToNode(nodeRes, podRes) {
			test1 := nodeRes.MilliCpuLeft >= podRes.MilliCpu      // Check if the node has enough CPU resources for the POD.
			test2 := utils.IsNodeAccessibleToPod(nodeRes, podRes) // Check if the node has the GPU type requested by the pod (if that's the case).
			test3 := utils.CanNodeHostPodOnGpuMemory(nodeRes, podRes)
			log.Debugf("DEBUG FRA, plugin.pwrexp_score.CalcExpPWRVarNode(): typical pod %s is not allocatable to node %s %t %t %t)\n",
				podRes.Repr(), nodeRes.Repr(), test1, test2, test3)
			skipped = true
			continue
		}

		// Variable used to store the increase in power consumption of the node if the pod is hypotetically scheduled on it.
		new_node_power := math.MaxFloat64

		// Case 1 - the typical pod requests a fraction of the resources of a single GPU.
		if podRes.IsGpuShare() {
			// For each GPU in the node, compute the node's estimated power consumption we would have by hypotetically assigning
			// the considered pod onto that GPU.
			var best_gpu_idx int = -1
			for i := range nodeRes.MilliGpuLeftList {

				// The considered GPU within the node has enough GPU-shared resources to accomodate the pod.
				if nodeRes.MilliGpuLeftList[i] >= podRes.MilliGpu {
					// Simulate how the available resources on a node would change by scheduling the pod onto a specific node's GPU.
					tmpNodeRes := nodeRes.Copy()
					tmpNodeRes.MilliCpuLeft -= podRes.MilliCpu
					tmpNodeRes.MilliGpuLeftList[i] -= podRes.MilliGpu

					// Now compute the increase in power consumption if allocating the current typical pod on this GPU.
					tmp_CPU_power, tmp_GPU_power := tmpNodeRes.GetEnergyConsumptionNode()
					tmp_node_power := tmp_CPU_power + tmp_GPU_power
					log.Debugf("DEBUG FRA, plugin.pwrexp_score.CalcExpPWRVarNode(): Pwr consumption computed for node %s, GPU %d, with sharing-GPU typical pod %s: %f\n",
						tmpNodeRes.NodeName, i, podRes.Repr(), tmp_node_power)

					// ### Update the node's best score ### //
					if tmp_node_power < new_node_power {
						best_gpu_idx = i
						new_node_power = tmp_node_power
					}
				}

			}

			// Sanity check to see if we found a GPU that can accomodate the pod (shouldn't give error!).
			if best_gpu_idx >= 0 {
				// log.Debugf("DEBUG FRA, plugin.pwrexp_score.CalcExpPWRVarNode(): Final expected pwr consumption for node %s: selected GPU %d, power %f\n",
				//	nodeRes.NodeName, best_gpu_idx, new_node_power)
			} else {
				log.Errorf("typical pod %v couldn't be allocated on node %s even if it had resources!\n", pod.TargetPodResource, nodeRes.NodeName)
			}

			// Case 2 - the pod requests no GPU (CPU only), or exactly one GPU, or multiple GPUs.
		} else {
			// Subtract the node's resources that would be taken by the pod if scheduled on it.
			tmpNodeRes, _ := nodeRes.Sub(podRes)

			// Compute the estimated power consumption of the node with the typical pod hypotetically allocated on it.
			tmp_CPU_power, tmp_GPU_power := tmpNodeRes.GetEnergyConsumptionNode()
			new_node_power = tmp_CPU_power + tmp_GPU_power

			log.Debugf("DEBUG FRA, plugin.pwrexp_score.CalcExpPWRVarNode(): Estimated power consumption for node %s with CPU-only or multi-GPU typical pod %s: %f\n",
				nodeRes.NodeName, podRes.Repr(), new_node_power)
		}

		// Save information about the node's power consumption with this typical pod added.
		list_allocatable_pods = append(list_allocatable_pods, Pair{pwr_inc: new_node_power, prob: podFreq})
	}

	// Step 3 - If some typical pods cannot be allocated on this node, renormalize the probabilities of the ones that can be.
	// TODO: deal with case in which no typical pod can be allocated on the node.
	if skipped {
		log.Debugf("DEBUG FRA, plugin.pwrexp_score.CalcExpPWRVarNode(): Renormalizing typical pods probabilities for node %s\n", nodeRes.NodeName)
		sum_probs := 0.
		if len(list_allocatable_pods) > 0 {
			for _, pod_info := range list_allocatable_pods {
				sum_probs += pod_info.prob
			}
			for i := range list_allocatable_pods {
				list_allocatable_pods[i].prob /= sum_probs
			}
		}
		log.Debugf("DEBUG FRA, plugin.pwrexp_score.CalcExpPWRVarNode(): Sum typical pods' probabilities for node %s: %f\n", nodeRes.NodeName, sum_probs)
	}

	// Step 4 - Compute the expected power variation.
	expPwrVariation = 0.
	for i := range list_allocatable_pods {
		// Compute the increase in power consumption of the node if the pod is hypotetically scheduled on it.
		expPwrVariation += (list_allocatable_pods[i].pwr_inc - old_node_power) * list_allocatable_pods[i].prob
	}

	return expPwrVariation
}

// This function selects the best GPU(s) found in a given node. It essentially re-executes the allocateGpuIdBasedOnPWREXPScore function
// executed within Score(), but it considers only the best GPU(s) for a pod found in a node and ignores the computed score.
func allocateGpuIdBasedOnPWREXPScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, _ simontype.GpuPluginCfg, typicalPods *simontype.TargetPodList) (gpuId string) {
	log.Debugf("DEBUG FRA, plugin.pwrexp_score.allocateGpuIdBasedOnPWREXPScore() => Scoring node %s w.r.t. pod!\n", nodeRes.NodeName)
	_, gpuId = calculatePWREXPShareExtendScore(nodeRes, podRes, typicalPods)
	return gpuId
}
