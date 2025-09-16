The ```EFRA``` folder contains the following elements:

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
