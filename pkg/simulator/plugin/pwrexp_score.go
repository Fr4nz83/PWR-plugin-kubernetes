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

	// Find the minimum (largest negative) score. The more the negative, the worst.
	// NOTE: The best (largest) possible score is known to be 0, i.e., no increase in expected power increase.
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
		log.Debugf("DEBUG FRA, plugin.pwrexp_score.NormalizeScore(): all the scores are equal.\n")

		for i, _ := range scores {
			scores[i].Score = 100
			log.Debugf("DEBUG FRA, plugin.pwrexp_score.NormalizeScore(): normalized score for node %s: %d\n", scores[i].Name, scores[i].Score)
		}

		return framework.NewStatus(framework.Success)
	}

	// Normalize the scores to the range [0, 100].
	for i, _ := range scores {
		// Normalization formula: normalized_score = (score - minScore) / (0 - minScore) * 100
		scores[i].Score = (scores[i].Score - minScore) * 100 / (maxScore - minScore)
		log.Debugf("DEBUG FRA, plugin.pwrexp_score.NormalizeScore(): normalized score for node %s: %d\n", scores[i].Name, scores[i].Score)
	}

	return framework.NewStatus(framework.Success)
}

// This function computes the score of a node w.r.t. an unscheduled pod. This is done by hypotetically scheduling the pod on the node,
// and then measure how much the node's power consumption increases.
func calculatePWREXPShareExtendScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, typicalPods *simontype.TargetPodList) (score int64, gpuId string) {

	// TODO: use the typicalPods variable to compute the expected increase in power consumption onto this node.
	// Compute the node's current expected power consumption increase.
	curr_exp_pwr_inc := CalcExpPWRIncNode(nodeRes, typicalPods)

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

				// Compute the expected increase in power consumption with the pod hypotetically allocated onto the node.
				hyp_exp_pwr_inc := CalcExpPWRIncNode(newNodeRes, typicalPods)
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

		// Compute the expected increase in power consumption with the pod hypotetically allocated onto the node.
		hyp_exp_pwr_inc := CalcExpPWRIncNode(newNodeRes, typicalPods)
		pwrScore := int64(curr_exp_pwr_inc - hyp_exp_pwr_inc)
		log.Debugf("DEBUG FRA, plugin.pwrexp_score.calculatePWREXPShareFragExtendScore(): Scoring node %s with CPU-only or multi-GPU pod: %d\n",
			nodeRes.NodeName, pwrScore)

		return pwrScore, simontype.AllocateExclusiveGpuId(nodeRes, podRes)
	}
}

// This function selects the best GPU(s) found in a given node. It essentially re-executes the allocateGpuIdBasedOnPWREXPScore function
// executed within Score(), but it considers only the best GPU(s) for a pod found in a node and ignores the computed score.
func allocateGpuIdBasedOnPWREXPScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, _ simontype.GpuPluginCfg, typicalPods *simontype.TargetPodList) (gpuId string) {
	log.Debugf("DEBUG FRA, plugin.pwrexp_score.allocateGpuIdBasedOnPWREXPScore() => Scoring node %s w.r.t. pod!\n", nodeRes.NodeName)
	_, gpuId = calculatePWREXPShareExtendScore(nodeRes, podRes, typicalPods)
	return gpuId
}

// This function computes the expected power increase onto a node when considering the pods of a typical workload represented by 'typicalPods'.
func CalcExpPWRIncNode(nodeRes simontype.NodeResource, typicalPods *simontype.TargetPodList) (pwrExpIncrease float32) {

	// Consider the pods in the target workload.
	for _, pod := range *typicalPods {
		// Check if the current pod in the target workload has a probability that makes sense. If not, ignore the pod in
		// the target workload (shouldn't happen!).
		freq := pod.Percentage
		if freq < 0 || freq > 1 {
			log.Errorf("pod %v has bad freq: %f\n", pod.TargetPodResource, freq)
			continue
		}

		// TODO: to be continued...
	}
	return pwrExpIncrease
}
