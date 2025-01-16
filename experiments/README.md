# Experimental evaluation pipeline

First, you need to ensure that the binary file `simon` has been generated in the `bin` directory. This requires to compile the Go code, which is detailed in the "How to compile the Go code" section of [README](../README.md).
Secondly, make sure that the traces have been translated from CSV to YAMLs -- to this end, please refer to the [README](../data/README.md) in the `data` directory.

Once these two operations are done, you are ready to prepare and execute the experimental evaluation pipeline used in our paper.



## 1. Running the simulations

### 1.1. Generation of the scripts that execute the simulations

The first step requires to generate the bash script in charge of executing a batch of simulations. The Python script `generate_run_scripts.py` is in charge of doing so, and is located in the `experiments/run_scripts` path. For example, assuming that this repository is located in the `PWR-plugin-kubernetes` folder in your machine, you can generate the bash script by executing the lines below from the command line:

```bash
# pwd: PWR-plugin-kubernetes/experiments
$ python run_scripts/generate_run_scripts.py > run_scripts/run_scripts_0511.sh
```

Please note that **_you can (and should!)_** customize some of the variables in `generate_run_scripts.py` that regulate the bash script's generation. 
`generate_run_scripts.py` is self-documented, hence this customization can be easily done. In particular, you should customize the following variables:

- **DATE**: this is simply the name of the folder that will be used to store the simulations' results.
- **REPEAT**: number of repetitions for simulation conducted with a specific trace. 10 is the value used in our experimental evaluation.
- **PARALLEL_SIMULATIONS**: number of simulations that the bash script will run in parallel at a time. WARNING: you should select a value compatible with the computational and memory resources you have.
- **FILELIST**: list containing the names of the traces that will be considered during the simulations. The list contains the traces we considered in our experimental evaluation.
- **AllMethodList**: the list of scoring plugins considered in the simulations. Note that here you can linearly combine two scoring plugins -- _in our experimental evaluation, we combined our PWR plugin with the FGD one_.

Finally, note that the output of the simulations is designed to be **deterministic** (and thus reproducible); this is ensured by using a given range of values (i.e., [42, 42 + REPEAT] for the `--tuneseed` flag passed to the simulator.


### 1.2. Execute the simulations

Assuming that `run_scripts_0511.sh` is the name of the bash script generated at step 1, you can now execute the batch of simulations represented by the script from the root of the project.
For example:

```bash
# pwd: PWR-plugin-kubernetes
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


## 2. Analysis of the simulations' results (part of the text adapted from the original repository)

The folder in which the results of the simulations will be stored will be located in `experiments`. Recall from the [generation of scripts](#11-generation-of-the-scripts-that-execute-the-simulations) Section that the folder's name depends on the DATE variable -- e.g., if `DATE=2023_0511`, then this will be the name used for the results' folder. In this folder, you will find the results of the simulations. Here, each simulation has its own subfolder, where you will find several files with the simulation results. **The files you will be mainly interested into are the `analysis_*` ones**. 

Assume that we are evaluating 6 scheduling policies, and that we are evaluating each policy on 17 different traces, each with its own workload's distributions. Furthermore, asume that for each policy-trace combination, we repeat the simulation 10 times to ensure results' reliability. Thus, in this example we will have a total number of `6 x 17 x 10 = 1020` simulations to conduct. The results of these 1020 simulations will have the following structure:

```bash
├── 01-Random
│   ├── dataset_01_default
│   │   ├── random_seed_42
│   │   ├── random_seed_43
│   │   ├── random_seed_..
│   │   └── random_seed_51
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

As mentioned before, within each simulation's subfolder you will find several files. The interesting ones are some of the `analysis_*` files, each focusing on a specific metric considered during the simulation. More precisely, **in our experimental evaluation we consider**:

- `analysis_allo.csv`: contains information regarding the resources requested and allocated onto the GPU datacenter as the simulation progresses.
- `analysis_frag.csv`: contains information regarding GPU fragmentation as the simulation progresses.
- `analysis_pwr.csv`: contains information regarding the esitmated CPU and GPU power consumption within the simulated GPU datacenter as the simulation progresses.
