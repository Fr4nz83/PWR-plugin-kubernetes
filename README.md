# *Power- and Fragmentation-aware Online Scheduling for GPU Datacenters*

This repository contains the code behind the paper "[Power- and Fragmentation-aware Online Scheduling for GPU Datacenters](https://arxiv.org/abs/2412.17484)". 
It started as a fork of the repository behind the seminal paper "[Beware of Fragmentation: Scheduling GPU-Sharing Workloads with Fragmentation Gradient Descent](https://www.usenix.org/system/files/atc23-weng.pdf)" from Weng, Qizhen, et al., and now includes the code behind our paper's contributions:

- our power-aware online scheduling policy, **PWR**, in the form of a Kubernetes scoring plugin. The core of the PWR plugin can be found in the Go source file [`pwr_score.go`](pkg/simulator/plugin/pwr_score.go);
- added the power consumption telemetry feature to the simulator -- this required to modify several simulator's source files;
- provide the Python and Bash scripts used for the paper's experimental evaluation, in order to ensure reproducibility.


## How to compile the code

The code can be theoretically compiled on any platform. First ensure that Go is installed. Then:

`go mod vendor` installs the dependencies required to compile the code. 

```bash
$ go mod vendor
```

`make` generates the compiled binary files in the `bin` directory.

```bash
$ make
```

## How to reproduce our experimental evaluation's results

The Python dependencies required to run the Python scripts behind our experimental evaluation are listed in the file `requirements.txt`. They can be installed by executing:

```bash
$ pip install -r requirements.txt
```

Then follow these steps:

1. **translate the production traces from CSV to YAML** -- this is required to run the experiments with the simulator. To this end, read [README](data/README.md) under the `data` directory for more information.
2. **execute the experiments conducted in the paper**. To this end, read [README](experiments/README.md) under the `experiments` directory for more information. **TODO: update the document**.


## Cite us

If you used or appreciated this repository's contributions, please cite the following paper:

Lettich, F., Carlini, E., Nardini, F. M., Perego, R., & Trani, S. (2024). *Power-and Fragmentation-aware Online Scheduling for GPU Datacenters*. arXiv preprint arXiv:2412.17484.

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
