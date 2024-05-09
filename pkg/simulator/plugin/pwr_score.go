package plugin

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	simontype "github.com/hkust-adsl/kubernetes-scheduler-simulator/pkg/type"
)

type PWRScorePlugin struct {
	handle      framework.Handle
	typicalPods *simontype.TargetPodList
}

// var _ framework.ScorePlugin = &PWRScorePlugin{} // This assignment is used at compile-time to check if the class implements the plugin interface.

// NOTE: typical pods should represent the target workload, i.e., pods passed via YAMLs before workload inflation.
// These are required to compute the cluster fragmentation.
func NewPWDScorePlugin(_ runtime.Object, handle framework.Handle, typicalPods *simontype.TargetPodList) (framework.Plugin, error) {
	fmt.Printf("DEBUG FRA, plugin.fgd_score.NewPWDScorePlugin() => Instantiating PWD plugin!\n")

	plugin := &PWRScorePlugin{
		handle:      handle,
		typicalPods: typicalPods,
	}

	allocateGpuIdFunc[plugin.Name()] = allocateGpuIdBasedOnFGDScore
	return plugin, nil
}

func (plugin *PWRScorePlugin) Name() string {
	return simontype.FGDScorePluginName
}

// TODO: da completare.

// Understand if it makes sense that the PWR plugin uses this function.
func allocateGpuIdBasedOnPWDScore(nodeRes simontype.NodeResource, podRes simontype.PodResource, _ simontype.GpuPluginCfg, typicalPods *simontype.TargetPodList) (gpuId string) {
	_, gpuId = calculateGpuShareFragExtendScore(nodeRes, podRes, typicalPods)
	return gpuId
}
