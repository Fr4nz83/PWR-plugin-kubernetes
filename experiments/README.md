# Experimental evaluation pipeline

First, ensure the binary file `simon` has been generated in the `bin` directory (see "Environment Setup" in [README](../README.md) for details).
Then, you can execute the steps below.


### 1. Generation of the experiments' scripts

```bash
# pwd: kubernetes-scheduler-simulator/experiments
$ python run_scripts/generate_run_scripts.py > run_scripts/run_scripts_0511.sh
```


### 2. Execute

`run_scripts_0511.sh` includes multiple executable commands that are executed sequentially by default.
You can adjust the `--max-procs` parameter in the following command to modify the number of parallel threads based on the CPU resources available on your machine.
It is recommended to configure the parallel thread pool size to **half the number of virtual CPUs available** (i.e., `# of vCPU / 2`).

```bash
# pwd: kubernetes-scheduler-simulator/experiments
$ cd ..
# pwd: kubernetes-scheduler-simulator
$ cat experiments/run_scripts/run_scripts_0511.sh | while read i; do printf "%q" "$i"; done | xargs --max-procs=16 -I CMD bash -c CMD
# "--max-procs=16" where 16 is the degree of PARALLEL suggested above
# bash run_scripts_0511.sh will run experiments sequentially
```

From the original repository's description => to explain the bash script generated (e.g., `run_scripts_0511.sh`):
- Each experiment is conducted via the Python script [scripts/generate_config_and_run.py](../scripts/generate_config_and_run.py). The script conducts an experiment by executing the following three steps:
    - First, the script generates two configuration yaml files in that folder, which are served as input to `bin/simon apply` (i.e., cluster-config and scheduler-config, see "Quickstart Example" in repo [README](../README.md)), 
    - Then, it executes the `bin/simon apply` command (confirmed by passing the `-e` parameter to the script)
    - The simulator's executable, i.e., `bin/simon`, will schedule the tasks and produce a scheduling log file in the corresponding folder.
- Afterwards, [scripts/analysis.py](../scripts/analysis.py) is executed to parse logs and yields multiple `analysis_*` files in the folder.

Please, be aware that executing many simulation in parallel takes a lot of computational and memory resources and, depending on the available resources, can take a lot of time. From the original repository, the authors report that it takes around:
- 10 minutes for 1 experiment on 2 vCPU, 9.4MB disk space for logs.
- 10 hours for 1020 experiments on a 256 vCPU machine with pool size of 128 threads, 9.4GB disk space for logs


### 3. Analysis & Merge

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
