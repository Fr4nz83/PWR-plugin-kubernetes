package utils

const (
	CpuModelName = "alibabacloud.com/cpu-model"
	ResourceName = "alibabacloud.com/gpu-milli"      // GPU milli, i.e., 1000 == 1 GPU, for pod only, node is 1000 by default
	CountName    = "alibabacloud.com/gpu-count"      // GPU number request (or allocatable), for pod and node
	DeviceIndex  = "alibabacloud.com/gpu-index"      // Exists when the pod are assigned/predefined to a GPU device
	ModelName    = "alibabacloud.com/gpu-card-model" // GPU card model, for pod and node
	AssumeTime   = "alibabacloud.com/assume-time"    // To retrieve the scheduling latency
	CreationTime = "alibabacloud.com/creation-time"  // creation timestamp
	DeletionTime = "alibabacloud.com/deletion-time"  // deletion timestamp
	PodNsNameSep = "/"
	DevIdSep     = "-"
	MILLI        = 1000

	MaxSpecCpu  = 128000  // CPU MILLI
	MaxSpecMem  = 1048576 // Mem MiB
	MaxSpecGpu  = 8000    // GPU MILLI
	NoGpuTag    = "no-gpu"
	ShareGpuTag = "share-gpu"
)

var MapGpuTypeMemoryMiB = map[string]int64{
	"P4":      int64(7980711936 / 1024 / 1024),  //  7611 MiB, "Tesla-P4"
	"2080":    int64(11554258944 / 1024 / 1024), // 11019 MiB, "GeForce-RTX-2080-Ti", "NVIDIA-GeForce-RTX-2080-Ti"
	"1080":    int64(11720982528 / 1024 / 1024), // 11178 MiB, "GeForce-GTX-1080-Ti"
	"M40":     int64(12004098048 / 1024 / 1024), // 11448 MiB, "Tesla-M40"
	"T4":      int64(15842934784 / 1024 / 1024), // 15109 MiB, "Tesla-T4"
	"V100M16": int64(16944988160 / 1024 / 1024), // 16160 MiB, "Tesla-V100-SXM2-16GB"
	"P100":    int64(17070817280 / 1024 / 1024), // 16280 MiB, "Tesla-P100-PCIE-16GB"
	"A10":     int64(23835181056 / 1024 / 1024), // 22731 MiB, "A10", "NVIDIA-A10"
	"3090":    int64(25446842368 / 1024 / 1024), // 24268 MiB, "GeForce-RTX-3090"
	"V100M32": int64(34089205760 / 1024 / 1024), // 32510 MiB, "Tesla-V100-SXM2-32GB", "Tesla-V100S-PCIE-32GB"
	"A100":    int64(85198045184 / 1024 / 1024), // 81251 MiB, "A100", "A100-SXM4-80GB"
	"G1":      int64(1048576000 / 1024 / 1024),  // 10000 MiB,
	"G2":      int64(20971520000 / 1024 / 1024), // 20000 MiB,
	"G3":      int64(31457280000 / 1024 / 1024), // 30000 MiB,
}

// The map of maps below stores the idle/max power consumption of several GPUs (in watt).
// We only consider the case in which a GPU's cores are not physically partitioned, e.g., MiG.
// TDP ratings can be easily found in official docs, however power consumptions when idling are not easy to find.
//
// TODO2: SALVO => in futuro, sostituire le costanti con funzioni che modellano il consumo energetico in funzione del workload.
var MapGpuTypeEnergyConsumption = map[string]map[string]float64{
	"T4":   {"idle": float64(10), "full": float64(70)},  // From https://www.nvidia.com/it-it/data-center/tesla-t4/
	"A10":  {"idle": float64(30), "full": float64(150)}, // From https://www.nvidia.com/content/dam/en-zz/Solutions/Data-Center/a10/pdf/a10-datasheet.pdf
	"P100": {"idle": float64(30), "full": float64(250)}, // From https://sc20.supercomputing.org/proceedings/tech_poster/poster_files/rpost131s2-file2.pdf
	"V100": {"idle": float64(30), "full": float64(300)}, // From https://sc20.supercomputing.org/proceedings/tech_poster/poster_files/rpost131s2-file2.pdf
	"A100": {"idle": float64(50), "full": float64(400)}, // From https://images.nvidia.com/aem-dam/en-zz/Solutions/data-center/nvidia-ampere-architecture-whitepaper.pdf
}

// The map of maps below stores the idle/max power consumption (in watts), as well as the number of cores, of several CPU models.
//
// NOTE: information concerning the instances typically found in Alibaba clusters can be found at:
// https://www.alibabacloud.com/help/en/ecs/user-guide/overview-of-instance-families
// https://www.alibabacloud.com/en/product/ecs-pricing-list/en?_p_lc=1#/?_k=8oavlr
// The first link does not contain any references to instances with A100, but the second one does.
//
// TODO: populate the table with the CPU(s) used in typical Alibaba clusters.
var MapCpuTypeEnergyConsumption = map[string]map[string]float64{
	"":      {"idle": float64(35), "full": float64(300), "ncores": float64(32)}, // If no node's CPU type is provided (shouldn't happen!), assume this CPU profile.
	"Intel": {"idle": float64(15), "full": float64(200), "ncores": float64(16)}, // If no node's CPU type is provided (shouldn't happen!), assume this CPU profile.
}
