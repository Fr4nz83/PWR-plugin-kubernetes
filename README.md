This repository contains the code behind the main contributions and experimental evaluation presented in the paper "[Power- and Fragmentation-aware Online Scheduling for GPU Datacenters](https://arxiv.org/abs/2412.17484)".

This repository started as a fork of the repository behind the paper "[Beware of Fragmentation: Scheduling GPU-Sharing Workloads with Fragmentation Gradient Descent](https://www.usenix.org/system/files/atc23-weng.pdf)", and includes the code behind our paper's contributions:

1. introduce our power-aware online scheduling policy, **PWR**, in the form of a Kubernetes scoring plugin;
2. slightly customized the simulator to support power consumption telemetry;
3. provide the Python and Bash scripts used for the paper's experimental evaluation, to ensure reproducibility.


## How to compile the code

First ensure that Go is installed. Then:

`go mod vendor` installs the dependencies required to compile the code. 

```bash
$ go mod vendor
```

`make` generates the compiled binary files in the `bin` directory.

```bash
$ make
```

## How to reproduce our experimental evaluation's results

The Python dependencies required to run the Python scripts behind our experimental evaluation are listed in the file requirements.txt. They can be installed by executing:

```bash
$ pip install -r requirements.txt
```

Then follow these steps:

1. read [README](data/README.md) under the `data` directory to prepare the production traces required to run the experiments. **TODO: update the document**.
2. read [README](experiments/README.md) under the `experiments` directory to execute the same experiments conducted in the paper. **TODO: update the document**.


## Cite us

If you used or appreciated this repository's contributions, please cite the following paper:

Lettich, F., Carlini, E., Nardini, F. M., Perego, R., & Trani, S. (2024). Power-and Fragmentation-aware Online Scheduling for GPU Datacenters. arXiv preprint arXiv:2412.17484.

```
@misc{lettich2024powerfragmentationawareonlinescheduling,
      title={Power- and Fragmentation-aware Online Scheduling for GPU Datacenters}, 
      author={Francesco Lettich and Emanuele Carlini and Franco Maria Nardini and Raffaele Perego and Salvatore Trani},
      year={2024},
      eprint={2412.17484},
      archivePrefix={arXiv},
      primaryClass={cs.DC},
      url={https://arxiv.org/abs/2412.17484}, 
}
```
