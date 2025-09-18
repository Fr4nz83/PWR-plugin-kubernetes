The ```EFRA``` folder contains all the material necessary to perform the simulations of the EFRA datacenter and EFRA workload considered in the deliverable 4.1.3.


# How to generate the YAMLs of the EFRA datacenter and workload used by the simulator

1. The user must first execute the jupyter notebooks ```1 - Create simulated EFRA workload.ipynb``` and ```2 - Translate simulated EFRA datacenter CSV to YAML.ipynb```: these will create a trace of the workload to be simulated (in CSV format) and will translate the EFRA datacenter specifications from CSV to YAML, respectively. 

2. The user must then execute the bash script ```3 - prepare_input.sh```, which will take in input the files generated at step 1 and will output a folder named ```openb_pod_list_default``` containing the YAMLs with the specifications of the EFRA datacenter and workload.

3. The user must copy the generated ```openb_pod_list_default``` folder into the ```data``` folder under the project's root directory.

4. At this point, the user must follow the usual steps needed to prepare the execution of batches of simulations; we thus refer the reader to the documentation in the project's root folder.
