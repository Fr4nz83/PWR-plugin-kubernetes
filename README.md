This repository contains the code behind the main contributions and experimental evaluation presented in the paper "*Power- and Fragmentation-aware Online Scheduling for GPU Datacenters*". **TODO: add arXiv reference.**

Our repository is a fork of the repository behind the paper "[Beware of Fragmentation: Scheduling GPU-Sharing Workloads with Fragmentation Gradient Descent](https://www.usenix.org/system/files/atc23-weng.pdf)". More specifically, we build their code to:

1. introduce our power-aware online scheduling policy, **PWR**, in the form of a Kubernetes scoring plugin;
2. slightly customize the simulator to support power consumption telemetry;
3. provide the Python scripts behind the extensive experimental evaluation conducted in our paper, to make it reproducible.

## How to compile the code

Please ensure that Go is installed.

`go mod vendor` installs the dependencies required for the simulator. 

```bash
$ go mod vendor
```

`make` generates the compiled binary files in the `bin` directory.

```bash
$ make
```

## How to reproduce our experimental evaluation

The Python dependencies required to run the Python scripts behind our experimental evaluation are listed in the file requirements.txt. They can be installed by executing:

```bash
$ pip install -r requirements.txt
```

1. Please refer to [README](data/README.md) under the `data` directory to prepare production traces. **TODO: update the document**.
2. Then refer to [README](experiments/README.md) under the `experiments` directory to reproduce the results reported in the paper. **TODO: update the document**.


## Cite us

**TODO**.
