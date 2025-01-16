# Experimental evaluation pipeline

First, you need to ensure that the binary file `simon` has been generated in the `bin` directory. This requires to compile the Go code, which is detailed in the "How to compile the Go code" section of [README](../README.md).
Secondly, make sure that the traces have been translated from CSV to YAMLs -- to this end, please refer to the [README](../data/README.md) in the `data` directory.

Once these two operations are done, you are ready to prepare and execute the experimental evaluation pipeline used in our paper.


## 1. Generation of the scripts that execute the simulations

The first step requires to generate the bash script in charge of executing a batch of simulations. The Python script `generate_run_scripts.py` is in charge of doing so, and is located in the `experiments/run_scripts` path. For example, you can do so by executing the lines below from the command line:

```bash
# pwd: kubernetes-scheduler-simulator/experiments
$ python run_scripts/generate_run_scripts.py > run_scripts/run_scripts_0511.sh
```

Please note that **_you can (and should!)_** customize some of the variables in `generate_run_scripts.py` that regulate the bash script's generation. 
The Python script has been documented, hence this customization can be easily done. In particular, you should customize the following variables:

- **DATE**: this is simply the name of the folder that will be used to store the simulations' results.
- **REPEAT**: number of repetitions for simulation conducted with a specific trace. 10 is the value used in our experimental evaluation.
- **PARALLEL_SIMULATIONS**: number of simulations that the bash script will run in parallel at a time. WARNING: you should select a value compatible with the computational and memory resources you have.
- **FILELIST**: list containing the names of the traces that will be considered during the simulations. The list contains the traces we considered in our experimental evaluation.
- **AllMethodList**: the list of scoring plugins considered in the simulations. Note that here you can linearly combine two scoring plugins -- _in our experimental evaluation, we combined our PWR plugin with the FGD one_.


## 2. Execute the simulations

Assuming that `run_scripts_0511.sh` is the name of the bash script generated at step 1, you can now execute the batch of simulations represented by the script from the root of the project.
For example:

```bash
# pwd: kubernetes-scheduler-simulator
$ ./experiments/run_scripts/run_scripts_0511.sh
```

Each simulation executed by the bash script is, in reality, made of two steps:
- First, the bash script invokes the Python script [scripts/generate_config_and_run.py](../scripts/generate_config_and_run.py). The purpose of this script is to run a simulation and it does so by executing the following three sub-steps:
    - First, the script prepares two configuration YAML files for the simulator in a simulation's subfolder, which are served as input to `bin/simon apply` (i.e., cluster-config and scheduler-config, see "Quickstart Example" in repo [README](../README.md)); 
    - Then, it executes the `bin/simon apply` command (confirmed by passing the `-e` parameter to the script)
    - The simulator's executable, i.e., `bin/simon`, will schedule the tasks and produce a scheduling log file in the corresponding simulation's subfolder.
- Afterwards, the bash script executes [scripts/analysis.py](../scripts/analysis.py), which parses logs and yields multiple `analysis_*` files in the smulation's subfolder.

Once again, please be aware that **executing many simulation in parallel takes a lot of computational, memory, and storage resources**. Furthermore, depending on the resources you have and the number of simulations that you can run in parallel, running many simulations in **can take a lot of time**. As a reference, in the original repository the authors of "[Beware of Fragmentation: Scheduling GPU-Sharing Workloads with Fragmentation Gradient Descent](https://www.usenix.org/system/files/atc23-weng.pdf)" report that it takes around:

- 10 minutes for 1 experiment on 2 vCPU, 9.4MB disk space for logs.
- 10 hours for 1020 experiments on a 256 vCPU machine with pool size of 128 threads, 9.4GB disk space for logs


### 3. Analysis of the results

As each experiment has its own folder where the `analysis_*` files are produced, here we bypass all folders and merge all the analysis files of the same category into one file under the `analysis/analysis_results` folder.

The top folder of the experiment varies with DATE (e.g., `2023_0511`), while the `analysis_merge.sh` is hard-coded to bypass and merge folders under the `data` folder. Therefore, we need to softlink the top folder to be merged to `data` (e.g., `$ ln -s 2023_0511 data`) before executing `$ bash analysis_merge.sh`.

```bash
# pwd: kubernetes-scheduler-simulator/experiments
$ ln -s 2023_0511 data # softlink it to data
# pwd: kubernetes-scheduler-simulator/experiments/analysis
$ cd analysis
# The output will be put under "analysis_results" folder
$ bash analysis_merge.sh
```

#### Structure of the 1020 Experiments and Results

Since we have 6 scheduling policies to evaluate (including baselines); for each policy, we have 17 traces (see `data`) where workloads have different distributions; for each policy-trace setting, we repeat the experiments 10 times to, as you pointed out, ensure the reliability of the results. Therefore, there are `6 x 17 x 10 = 1020` experiments to conduct.  These 1020 experiments have the following structure:

```bash
├── 01-Random
│   ├── dataset_01_default
│   │   ├── random_seed_01
│   │   ├── random_seed_02
│   │   ├── random_seed_..
│   │   └── random_seed_10
│   ├── dataset_02_fig11_1
│   ├── dataset_..
│   └── dataset_17_fig14_4
│
├── 02-DotProd
│   ├── dataset_..
│   │   ├── random_seed_..
│
├── 03-GpuClustering
│   ├── dataset_..
│   │   ├── random_seed_..
│
├── 04-GpuPacking
│   ├── dataset_..
│   │   ├── random_seed_..
│
├── 05-BestFit
│   ├── dataset_..
│   │   ├── random_seed_..
│
└── 06-FGD
    ├── dataset_..
    │   ├── random_seed_..
```

Nevertheless, the inner logic of the simulator is the same, regardless of the input traces and deployed scheduling policies. Instead of conducting all the experiments, it is recommended to **randomly pick any one of the experiments for execution** and compare its result with the files summarized in `analysis/expected_results`.

To better understand the result files, we take `analysis_allo_discrete.csv` as a example: The headers are the workload (traces), the scheduling policy, the workload inflation ratio (1.3 means 130% of cluster's GPU capacity), the random seed (42, 43, ..., 51), the total number of GPUs in the cluster, and a series of numbers from 0 to 130.

The 0-130 (`130 = tune (1.3) * 100`) represents the total GPU request (normalized by the `total_gpus`) of submitted arrived workloads; its corresponding values (e.g., 91.53 under column:130) is GPU allocation ratio in the cluster---which is the **key performance indicator (the higher the better)** shown in Fig. 9(a). The following selected results show that our policy 06-FGD outperforms 01-Random by 4% in allocation ratio given workloads with 130% GPU requests in `openb_pod_list_cpu050` traces.

| workload              | sc_policy | tune | seed | total_gpus | 0    | 1    | 2   | ... | 129   | 130       |
| --------------------- | --------- | ---- | ---- | ---------- | ---- | ---- | --- | --- | ----- | --------- |
| openb_pod_list_cpu050 | 01-Random | 1.3  | 42   | 6212       | 0.25 | 0.99 | 2.0 | ... | 91.51 | **91.53** |
| openb_pod_list_cpu050 | 06-FGD    | 1.3  | 42   | 6212       | 0.25 | 0.99 | 2.0 | ... | 95.34 | **95.34** |

The other result files show the fragmentation ratio (`analysis_frag_ratio_discrete.csv`, Fig 7(a)), fragmentation amount (`analysis_frag_discrete.csv`, Fig 7(b)), and distribution of failed pods (`analysis_fail.csv`, Fig 9(c)).

> Our results of the extensive 1020 experiments are cached in [analysis/expected_results](./analysis/expected_results/) (9.8MB) for your reference.
