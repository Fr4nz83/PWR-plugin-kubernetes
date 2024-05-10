package plugin

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	simontype "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type"
)

type PWRScorePlugin struct {
	handle      framework.Handle
	typicalPods *simontype.TargetPodList
}

var _ framework.ScorePlugin = &PWRScorePlugin{} // This assignment is used at compile-time to check if the class implements the plugin interface.

// NOTE: typical pods should represent the target workload, i.e., pods passed via YAMLs before workload inflation.
// These are required to compute the cluster fragmentation.
func NewPWDScorePlugin(_ runtime.Object, handle framework.Handle, typicalPods *simontype.TargetPodList) (framework.Plugin, error) {
	fmt.Printf("DEBUG FRA, plugin.fgd_score.NewPWDScorePlugin() => Instantiating PWD plugin!\n")

	plugin := &PWRScorePlugin{
		handle:      handle,
		typicalPods: typicalPods,
	}

	allocateGpuIdFunc[plugin.Name()] = allocateGpuIdBasedOnPWRScore
	return plugin, nil
}

func (plugin *PWRScorePlugin) Name() string {
	return simontype.FGDScorePluginName
}

// TODO: da completare.
func (plugin *PWRScorePlugin) Score(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) (int64, *framework.Status) {
	return 0, framework.NewStatus(framework.Success)
}

func (plugin *PWRScorePlugin) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// This function computes the score of a node w.r.t. an unscheduled pod. This is done by hypotetically scheduling the pod on the node,
// and then measure how much the node's fragmentation changes w.r.t. the target workload.
func calculatePWRShareFragExtendScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, typicalPods *simontype.TargetPodList) (score int64, gpuId string) {
	return 0, ""
}

// Understand if it makes sense that the PWR plugin uses this function.
func allocateGpuIdBasedOnPWRScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, _ simontype.GpuPluginCfg, typicalPods *simontype.TargetPodList) (gpuId string) {
	_, gpuId = calculatePWRShareFragExtendScore(nodeRes, podRes, typicalPods)
	return gpuId
}
