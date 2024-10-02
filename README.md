# 🚀 Kubernetes Scheduler Simulator

The simulator evaluates different scheduling policies in GPU-sharing clusters.
It includes the Fragmentation Gradient Descent (FGD) policy proposed in the [USENIX ATC 2023](https://www.usenix.org/conference/atc23) paper "[Beware of Fragmentation: Scheduling GPU-Sharing Workloads with Fragmentation Gradient Descent](https://www.usenix.org/conference/atc23/presentation/weng)", along with other baseline policies (e.g., Best-fit, Dot-product, GPU Packing, GPU Clustering, Random-fit). 

## 🚧 Environment Setup

### Build from stratch

Please ensure that Go is installed.

`go mod vendor` installs the dependencies required for the simulator. 

```bash
$ go mod vendor
```

`make` generates the compiled binary files in the `bin` directory.

```bash
$ make
```

## 🔮 Experiments on Production Traces

Install the required Python dependency environment.

```bash
$ pip install -r requirements.txt
```

1. Please refer to [README](data/README.md) under the `data` directory to prepare production traces.
2. Then refer to [README](experiments/README.md) under the `experiments` directory to reproduce the results reported in the paper.


## 🙏🏻 Acknowledge

Our simulator is developed based on [open-simulator](https://github.com/alibaba/open-simulator) by Alibaba, a simulator used for cluster capacity planning. 
This repository primarily evaluates the performance of different scheduling polices on production traces.
GPU-related plugin has been merged into the main branch of [open-simulator](https://github.com/alibaba/open-simulator).
