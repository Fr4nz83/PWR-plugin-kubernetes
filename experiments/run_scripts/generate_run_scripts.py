# Usage: python3 generate_run_scripts.py > run_scripts.sh

DATE = "2024_0606" # Used as the folder name under experiments/ to hold all log results. To avoid collision of repeated experiments, may change date or append _v1, _v2, etc.
REMARK = "Artifacts"
REPEAT = 10 # Number of repeats for each experiment.
PARALLEL_SIMULATIONS = 12 # Number of simulations that will be run in parallel.

FILELIST = [
    # Default trace
    "data/openb_pod_list_default",
    # no-GPU traces (not used in the paper's experimental evaluation)
#    "data/openb_pod_list_cpu050",
#    "data/openb_pod_list_cpu100",
#    "data/openb_pod_list_cpu200",
#    "data/openb_pod_list_cpu250",
    # GPU-sharing traces
    "data/openb_pod_list_gpushare100",
    "data/openb_pod_list_gpushare40",
    "data/openb_pod_list_gpushare60",
    "data/openb_pod_list_gpushare80",
    # GPU-constrained traces
    "data/openb_pod_list_gpuspec10",
    "data/openb_pod_list_gpuspec20",
    "data/openb_pod_list_gpuspec25",
    "data/openb_pod_list_gpuspec33",
    # multi-GPU traces
    "data/openb_pod_list_multigpu20",
    "data/openb_pod_list_multigpu30",
    "data/openb_pod_list_multigpu40",
    "data/openb_pod_list_multigpu50",
]

AllMethodList = [
    # Experiments with only one active scoring plugin.
    ["01", "Random", "random", "<none>", "<none>"],
    ["02", "DotProd", "best", "merge", "max"],
    ["03", "GpuClustering", "<none>", "<none>", "<none>"],
    ["04", "GpuPacking", "<none>", "<none>", "<none>"],
    ["05", "BestFit", "<none>", "<none>", "<none>"],
    ["06", "FGD", "<self>", "share", "max"],
    ["07", "PWR", "<self>", "share", "max"],
    # Experiments with linear combinations of PWR and FGD scoring plugins.
    ["08", "PWR 500 FGD 500", "FGD", "share", "max"],
    ["09", "PWR 200 FGD 800", "FGD", "share", "max"],
    ["11", "PWR 100 FGD 900", "FGD", "share", "max"],
    ["12", "PWR 50 FGD 950", "FGD", "share", "max"],
]

AllMethodDict = {}
for item in AllMethodList:
    AllMethodDict[item[0]] = item

#####################################################################
#####################################################################
#####################################################################

MethodList = AllMethodList.copy()


def get_dir_name_from_method(method_input):
    if len(method_input) != 5:
        print("[ERROR] get_dir_name_from_method: len(method) == 5, including id, policy, gsm, dem, nm")
        return "default_name"
    id, policy, gsm, dem, nm = method_input
    gsm = policy if gsm == "<self>" else gsm # no need to adjust, except that <self> is not allowed in bash. generate_config_and_run will recover the policy's full name
    dir_name = "%s-%s" % (id, policy.replace(' ', '_')) # If a policy contains multiple plugins, and thus multiple weights, replace the blank spaces with '_'.
    suffix = ""
    suffix += '_%s' % gsm if gsm != "<none>" else ''
    suffix += '_%s' % dem if dem != "<none>" else ''
    suffix += '_%s' % nm if nm != "<none>" else ''
    return dir_name # + suffix


def get_method_from_policy_id_list(id_list):
    if type(id_list) == list:
        return [AllMethodDict.get("%02d" % id, None) if type(id)==int else AllMethodDict.get("%s" % id, None) for id in id_list]
    else:
        return [AllMethodDict.get("%02d" % id_list, None) if type(id)==int else AllMethodDict.get("%s" % id_list, None)]


def get_dir_name_from_policy_id_list(id_list):
    return [get_dir_name_from_method(x) for x in get_method_from_policy_id_list(id_list)]

###########################################################
###########################################################
###########################################################


def generate_run_scripts(asyncc=True, parallel=16):
    DateAndRemark = DATE + "-" + REMARK.replace(' ', "_").replace('(',"_").replace(')',"_")
    numJobs=0
    if asyncc:
        print('#!/bin/bash\n# screen -dmS sim-%s bash -c "bash run_scripts_%s.sh"\n' % (DateAndRemark, DATE[-4:]))
    else:
        print('#!/bin/bash\n# cat run_scripts_%s.sh | while read i; do printf "%%q\\n" "$i"; done | xargs --max-procs=16 -I CMD bash -c CMD\n' % (DATE[-4:]))
    for tune_ratio in [1.3]:
        tune_seed_end = 42 + REPEAT if REPEAT >= 1 else 43
        for tune_seed in range(42, tune_seed_end, 1):
            for file in FILELIST:
                filename = file.split('/')[-1]
                for id, policy, gsm, dem, nm in MethodList:  # GpuSelMethod, DimExtMethod, NormMethod
                    dir_name = get_dir_name_from_method([id, policy, gsm, dem, nm])
                    gsm = policy if gsm == "<self>" else gsm
                    OUTPUT_YAML = False
                    SHUFFLE_POD = True
                    outstr = "# %s, %s, %s, %s, %s @ %s\n" % (id, policy, gsm, dem, nm, filename)
                    outstr += 'EXPDIR="experiments/%s/%s/%s/%s/%s' % (DATE, filename, dir_name, tune_ratio, tune_seed)
                    outstr += '" && mkdir -p "${EXPDIR}" && touch "${EXPDIR}/terminal.out" && '
                    outstr += 'python3 scripts/generate_config_and_run.py -d "${EXPDIR}" '
                    outstr += '-e -b '
                    outstr += '-f %s ' % file

                    # Case 1 - we are using a single scoring plugin
                    policies_args = policy.split()
                    if len(policies_args) == 1 :
                        outstr += '-%s 1000 ' % policy
                    # Case 2 - we are using multiple scoring plugins, each with its own weight
                    else :
                        for i in range(len(policies_args) // 2) :
                            outstr += f'-{policies_args[i*2]} {policies_args[i*2 + 1]} '


                    outstr += '-gpusel %s ' % gsm if gsm != "<none>" else ''
                    outstr += '-dimext %s ' % dem if dem != "<none>" else ''
                    outstr += '-norm %s ' % nm if nm != "<none>" else ''
                    outstr += '-tune %s ' % tune_ratio if tune_ratio else ''
                    outstr += '-tuneseed %s ' % tune_seed if tune_seed else ''
                    outstr += "--shuffle-pod=true " if SHUFFLE_POD else ""
                    outstr += '-y "${EXPDIR}/snapshot/yaml" ' if OUTPUT_YAML else ""
                    outstr += '-z "${EXPDIR}/snapshot/ds01" '
                    outstr += '| tee -a "${EXPDIR}/terminal.out" '
                    outstr += '&& python3 scripts/analysis.py -f -g ${EXPDIR} '
                    outstr += '| tee -a "${EXPDIR}/terminal.out" '
                    if asyncc:
                        outstr += " &"
                    print(outstr + "\n")

                    numJobs += 1
                    if asyncc and (numJobs % parallel == 0):
                        print("date & wait\n")  # force them to sync
    if asyncc:
        print("wait && date")


if __name__=='__main__':
    generate_run_scripts(asyncc=True, parallel=PARALLEL_SIMULATIONS)
    #: $ bash run_scripts.txt
    # generate_run_scripts(asyncc=False)
    #: $ cat run_scripts.txt | while read i; do printf "%q\n" "$i"; done | xargs --max-procs=16 -I CMD bash -c CMD
