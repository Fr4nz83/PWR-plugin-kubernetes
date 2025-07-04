# *Power- and Fragmentation-aware Online Scheduling for GPU Datacenters*

This repository contains the code behind the paper "[Power- and Fragmentation-aware Online Scheduling for GPU Datacenters](https://arxiv.org/abs/2412.17484)" by Francesco Lettich ([CNR-ISTI](https://www.isti.cnr.it/en/)), Emanuele Carlini (CNR-ISTI), Franco Maria Nardini (CNR-ISTI), Raffaele Perego (CNR-ISTI), and Salvatore Trani (CNR-ISTI).

The article has been accepted and presented to the **25th IEEE International Symposium on Cluster, Cloud, and Internet Computing ([CCGrid 2025](https://site.uit.no/ccgrid2025/))** conference. The slides of the presentation given at the conference can be found [here](CCGrid_2025/CCGrid%202025%20presentation.pdf).

This repository started as a fork of the repository behind the seminal 2023 USENIX paper "[Beware of Fragmentation: Scheduling GPU-Sharing Workloads with Fragmentation Gradient Descent](https://www.usenix.org/system/files/atc23-weng.pdf)" from Weng, Qizhen, et al.; now, it includes the customizations and code behind our paper's contributions, which is focused on minimizing power consumption alongside GPU fragmentation. 

More precisely, in this repository you will find:

- our power-aware online scheduling policy, **PWR**, in the form of a Kubernetes scoring plugin. The core of the PWR plugin can be found in the Go source file [`pwr_score.go`](pkg/simulator/plugin/pwr_score.go);
- the power consumption telemetry feature added to Alibaba's open-simulator (this required to modify some of the simulator's source files);
- Python and Bash scripts used for the paper's experimental evaluation, to ensure reproducibility.


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

## How to reproduce our experimental evaluation's pipeline and results

The Python dependencies required to run the Python scripts behind our experimental evaluation are listed in the file `requirements.txt`. They can be installed by executing:

```bash
$ pip install -r requirements.txt
```

Then, to reproduce the experimental pipeline used in our paper, please follow these steps:

1. **translate the production traces from CSV to YAML** -- this is required to run the experiments with the simulator. To this end, read [README](data/README.md) under the `data` directory for more information.
2. **execute the simulations conducted in the paper**. To this end, read Section 1 from the [README](experiments/README.md) under the `experiments` directory for more information. Please, be aware that the simulations can take a lot of time, depending on the amount of resources at your disposal.
3. **extract and plot the simulations' results**. To this end, read Section 2 from the [README](experiments/README.md#2-analysis-of-the-simulations-results-part-of-the-text-adapted-from-the-original-repository) under the `experiments` directory for more information.


## Cite us

Please cite our **CCGRID 2025** article if you have found our contributions useful, or you have used them within your work.

```bibtex
@inproceedings{lettich2025pwr,
  author={Lettich, Francesco and Carlini, Emanuele and Nardini, Franco Maria and Perego, Raffaele and Trani, Salvatore},
  booktitle={2025 IEEE 25th International Symposium on Cluster, Cloud and Internet Computing (CCGrid)}, 
  title={Power- and Fragmentation-Aware Online Scheduling for GPU Datacenters}, 
  year={2025},
  volume={},
  number={},
  pages={43-52},
  keywords={Training;Measurement;Power demand;Costs;Power measurement;Processor scheduling;Graphics processing units;Resource management;Sustainable development;Optimization;GPU Datacenter;Power-aware Scheduling;GPU-sharing;GPU Fragmentation;Online Scheduling;Green Computing;Sustainable Computing},
  doi={10.1109/CCGRID64434.2025.00015}}
```
