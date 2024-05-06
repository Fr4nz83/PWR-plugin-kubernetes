package v1alpha1

// This package defines several Go structs that define the structure of the objects used within the Kubernetes scheduler simulator. 
// These structs are used to define the configuration and specifications for simulating Kubernetes clusters and workloads.



// Main struct representing a simulation object.
// APIVersion: API version.
// Kind: Kind of the object.
// MetaData: Metadata for the simulation.
// Spec: Specifications for the simulation.
type Simon struct {
	APIVersion string        `json:"apiVersion"`
	Kind       string        `json:"kind"`
	MetaData   SimonMetaData `json:"metadata"`
	Spec       SimonSpec     `json:"spec"`
}

// Defines the specifications for the simulation.
// Cluster: Cluster configuration.
// AppList: List of applications.
// NewNode: Configuration for adding a new node.
// CustomConfig: Custom configurations.
type SimonSpec struct {
	Cluster      Cluster      `json:"cluster"`
	AppList      []AppInfo    `json:"appList"`
	NewNode      string       `json:"newNode"`
	CustomConfig CustomConfig `json:"customConfig,omitempty"`
}

// Metadata for the simulation object.
// Name: Name of the simulation.
type SimonMetaData struct {
	Name string `json:"name"`
}

// Defines information about an application.
// Name: Name of the application.
// Path: Path to the application.
// Chart: Indicates whether the application is chart-based.
type AppInfo struct {
	Name  string `json:"name"`
	Path  string `json:"path"`
	Chart bool   `json:"chart,omitempty"`
}

// Defines cluster-specific configurations.
// CustomCluster: Custom configuration for the cluster (probably the one to be given to the simulator?).
// KubeConfig: Kubernetes configuration for the cluster (this is probably the case in which we are testing a real cluster).
type Cluster struct {
	CustomCluster string `json:"customConfig,omitempty"`
	KubeConfig    string `json:"kubeConfig,omitempty"`
}

// Custom configurations for the simulation. Fields represent various configurations related to workload inflation, tuning, pod shuffling, descheduling, etc.
type CustomConfig struct {
	ShufflePod              bool                    `json:"shufflePod,omitempty"`
	ExportConfig            ExportConfig            `json:"exportConfig,omitempty"`
	WorkloadInflationConfig WorkloadInflationConfig `json:"workloadInflationConfig,omitempty"`
	WorkloadTuningConfig    WorkloadTuningConfig    `json:"workloadTuningConfig,omitempty"`
	NewWorkloadConfig       string                  `json:"newWorkloadConfig,omitempty"`
	DescheduleConfig        DescheduleConfig        `json:"descheduleConfig,omitempty"`
	TypicalPodsConfig       TypicalPodsConfig       `json:"typicalPodsConfig,omitempty"`
}

// Configuration for exporting snapshots.
// PodSnapshotYamlFilePrefix: Prefix for pod snapshot YAML files.
// NodeSnapshotCSVFilePrefix: Prefix for node snapshot CSV files.
type ExportConfig struct {
	PodSnapshotYamlFilePrefix string `json:"podSnapshotYamlFilePrefix,omitempty"`
	NodeSnapshotCSVFilePrefix string `json:"nodeSnapshotCSVFilePrefix,omitempty"`
}

// Configuration for workload inflation.
// Ratio: Ratio for workload inflation.
// Seed: Seed for workload inflation.
type WorkloadInflationConfig struct {
	Ratio float64 `json:"ratio,omitempty"`
	Seed  int64   `json:"seed,omitempty"`
}

// Configuration for workload tuning.
// Ratio: Ratio for workload tuning.
// Seed: Seed for workload tuning.
type WorkloadTuningConfig struct { // prune or append pods to match the Ratio * (cluster_GPU_capacity)
	Ratio float64 `json:"ratio,omitempty"` // <= 0 means no effects
	Seed  int64   `json:"seed,omitempty"`
}

// Configuration for descheduling.
// Ratio: Ratio for descheduling.
// Policy: Policy for descheduling.
type DescheduleConfig struct {
	Ratio  float64 `json:"ratio,omitempty"`
	Policy string  `json:"policy,omitempty"`
}

// Configuration for typical pods.
// IsInvolvedCpuPods: Indicates involvement of CPU pods.
// PodPopularityThreshold: Threshold for pod popularity.
// PodIncreaseStep: Step for increasing pods.
// GpuResWeight: Weight for GPU resources.
type TypicalPodsConfig struct {
	IsInvolvedCpuPods      bool    `json:"isInvolvedCpuPods,omitempty"`
	PodPopularityThreshold int     `json:"podPopularityThreshold,omitempty"` // [0-100]
	PodIncreaseStep        int     `json:"podIncreaseStep,omitempty"`
	GpuResWeight           float64 `json:"gpuResWeight,omitempty"`
}
