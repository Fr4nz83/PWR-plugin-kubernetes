The ```data``` folder contains the following elements:

- ```node_yaml``` (folder): contains the YAML with the specifications of the nodes of the GPU datacenter being simulated;
- ```csv``` (folder): contains the traces considered in our paper's experimental evaluation;
- ```bash prepare_input.sh```: bash script that executes the Python script pod_csv_to_yaml.py, which is in charge of translating the traces from CSV to YAML;
- ```pod_csv_to_yaml.py```: Python script executed by prepare_input.sh;
- ```0 - Workloads stats.ipynb```: can be used to explore the characteristics of the traces;
- ```1 - Nodes stats.ipynb```: can be used to explore the characteristics of the nodes of the simulated GPU datacenter;
- ```2 - Add CPU models to YAML nodes``` (not actually used in the experimental evaluation): can be used to add CPU models to the nodes of the simulated GPU datacenter.


# How to generate the YAMLs of the traces used by the simulator

The simulator requires two different YAML files to run an experiment:

- a YAML file that contains the **specifications of the nodes of the simulated GPU datacenter**. This is already provided in the ```node_yaml``` subfolder, hence you don't have to do nothing.
- a YAML file that represents an actual trace, i.e., **a sequence of pods with the resources they require from the GPU datacenter**. In our paper, pods correspond to tasks. During a simulation, the simulator uses a trace in YAML format to generate a workload with the trace's characteristics.

The original traces are stored in CSV files in the ```csv``` folder. The user must therefore translate the CSVs into YAMLs before running simulations, and they can do so by executing the bash script ```bash prepare_input.sh``` as follows:

```bash
$ bash prepare_input.sh
```

This will output a set of folders -- one per trace and with the name of each folder being the same of the CSV trace from which it has been generated. Each folder will contain:

- a YAML containing the specifications of the nodes of the datacenter being simulated -- this is actually the same for all the folders.
- a YAML containing the specifications of the pods of the trace the folder refers to.

Please find below more details on the traces and the nodes of the simulated GPU datacenter.


## Notes on the workload traces

The original traces are stored in CSV format in the `csv` folder. There are 5 different types of traces in this dataset. For more information on them, please refer to our paper. In the following, we briefly report their main characteristics:

- **Default** trace, consisting of 8,152 tasks collected from an Alibaba production-grade GPU datacenter without GPU constraints. Represented by the CSV `openb_pod_list_default.csv`.
- **multi-GPU** traces: the amount of GPU resources requested by tasks that use 1 or more entire GPUs is increased by 20%, 30%, 40%, and 50% compared to the Default trace. This is achieved by increasing the total number of multi-GPU tasks while keeping their internal distribution fixed. The numbers of CPU-only and sharing-GPU tasks remain unchanged. They are represented by the CSVs with the name `openb_pod_list_multigpuXX.csv`.
- **Sharing-GPU** traces: the percentage of GPU resources requested by sharing-GPU tasks is set at 40%, 60%, 80%, and 100% of the total GPU resources requested by GPU tasks. This is done by adjusting the number of sharing-GPU and multi-GPU tasks, while keeping intra-class distributions fixed and maintaining the same percentage of CPU-only tasks. They are represented by the CSVs with the name `openb_pod_list_gpushareXX.csv`.
- **Constrained-GPU** traces: the percentage of GPU tasks that request specific GPU models is set at 10%, 20%, 25%, and 33%. All other characteristics match those of Default. They are represented by the CSVs with the name `openb_pod_list_gpuspecXX.csv`.
- **no-GPU** traces: not considered in the paper. The percentage of tasks that do not request any GPU is varied w.r.t. the Default trace. Represented by the files `openb_pod_list_cpuXXX.csv`.

To illustrate the information contained in these CSVs, we provide a brief overview of the Default trace's content. Text is taken from the README.md of the original repository.

### [openb_pod_list_default.csv](./csv/openb_pod_list_default.csv)

This file contains 8152 tasks submitted to the GPU cluster, listing their resource specifications, QoS, phase and creation/deletion/scheduled times. 

The other openb_pod_list_*.csv files (excluding the gpuspec ones) are sampled from the default one, emphasizing certain types of workloads (e.g., CPU-only tasks, GPU-sharing tasks, multi-GPU tasks).

Trace files with `gpuspec` augment tasks with GPU type requirements. About 33% of GPU tasks in the production cluster have GPU type constraints (see also the information in the paper).

Here's a sample output:

|    | name           |   cpu_milli |   memory_mib |   num_gpu |   gpu_milli | gpu_spec        | qos       | pod_phase   |   creation_time |   deletion_time |   scheduled_time |
|---:|:---------------|------------:|-------------:|----------:|------------:|:----------------|:----------|:------------|----------------:|----------------:|-----------------:|
|  0 | openb-pod-0017 |       88000 |       327680 |         8 |        1000 | nan             | Burstable | Succeeded   |         9437497 |        10769854 |          9437497 |
|  1 | openb-pod-0022 |        4000 |        15258 |         1 |         220 | nan             | BE        | Running     |         9679175 |         9973826 |          9679175 |
|  2 | openb-pod-0035 |       16000 |        32768 |         1 |        1000 | V100M16\|V100M32 | LS        | Running     |         9967058 |         9968575 |          9967063 |

- `cpu_milli`: Number of CPUs requested (in milli)
- `memory_mib`: Main memory requested (in MiB)
- `num_gpu`: Number of GPUs requested (integers from 0 to 8)
- `gpu_milli`: Detailed GPU requested for GPU-sharing workloads (i.e., `num_gpu==1`) (in milli).
- `gpu_spec`: Required GPU types, For example, `nan` means no GPU type constraints while `V100M16|V100M32` means the task can run on [NVIDIA V100](https://www.nvidia.com/en-us/data-center/v100/) with either 16GB VRAM or 32GB VRAM.
- `qos`: [Quality of Service](https://kubernetes.io/docs/concepts/workloads/pods/pod-qos/) (e.g., Burstable, Best Effort (BE), Latency Sensitive (LS))
- `pod_phrase`: Succeeded, Running, Pending, Failed
- `creation_time`: Timestamp of creation (in seconds)
- `deletion_time`: Timestamp of deletion (in seconds)
- `scheduled_time`: Timestamp of being scheduled (in seconds)


## Notes on the nodes of the simulated GPU datacenter

In the following, a brief overview concerning the specifications of the nodes of the simulated GPU datacenter is provided. Text is from the README.md of the original repository, and presents the specifications according to the content and format used in [openb_node_list_all_node.csv](./csv/openb_node_list_all_node.csv).

More precisely, this CSV contains 1523 nodes of a heterogeneous GPU cluster in production, listing their CPU, main memory, GPU specifications and GPU types.

[openb_node_list_gpu_node.csv](./csv/openb_node_list_gpu_node.csv) is a subset excluding non-GPU nodes. 
[openb_node_list_gpu_node.yaml](./node_yaml/openb_node_list_gpu_node.yaml) contains the same data in YAML format.

Here's a sample output:

|    | sn              |   cpu_milli |   memory_mib |   gpu | model   |
|---:|:----------------|------------:|-------------:|------:|:--------|
|  0 | openb-node-0227 |       32000 |       262144 |     0 | nan     |
|  1 | openb-node-0228 |      128000 |       786432 |     8 | G3      |
|  2 | openb-node-0229 |       96000 |       786432 |     8 | V100M32 |

- `cpu_milli`: Number of CPUs (in milli)
- `memory_mib`: Main memory (in MiB)
- `gpu`: Number of GPUs
- `model`: GPU type. G1, G2, G3 are undisclosed internal GPU codes.
