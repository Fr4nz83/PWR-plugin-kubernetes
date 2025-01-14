# How to generate the YAMLs of the traces used in the experiments

This folder contains the following elements:

- node_yaml (folder): this folder contains the YAML containing the specifications of the set of nodes of the GPU datacenter being simulated.
- csv (folder): this folder contains the traces considered in the experimental evaluation.
- bash prepare_input.sh: this bash script executes the Python script in charge of generating the YAML files of the traces.
- pod_csv_to_yaml.py: Python script executed by prepare_input.sh
- a bunch of Jupyter notebooks that can be used to explore the characteristics of the traces as well as the nodes of the simulated GPU datacenter.


## Information on the nodes of the simulated GPU datacenter

### [openb_node_list_all_node.csv](./csv/openb_node_list_all_node.csv)

This CSV contains 1523 nodes of a heterogeneous GPU cluster in production, listing their CPU, main memory, GPU specifications and GPU types.

[openb_node_list_gpu_node.csv](./csv/openb_node_list_gpu_node.csv) is a subset excluding non-GPU nodes. [openb_node_list_gpu_node.yaml](./node_yaml/openb_node_list_gpu_node.yaml) contains the same data in YAML format.

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


## Workload traces

The original traces are stored in the csv folder. YAML files must be generated from this data (see below). To illustrate the information contained in these csv, as an example
we report the brief overview on the content of the Default trace (from the README.md of the original repository).

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


## How to generate the YAMLs of the traces, which are required to run the simulations.

While the YAML file containing the nodes' specifications is already provided in the "node_yaml" folder, the users needs to generate the YAMLs of the traces from the CSVs -- this is required to run the various experiments in the simulator. To generate the YAMLs, the user can execute the Python script below.

```bash
$ bash prepare_input.sh
```
