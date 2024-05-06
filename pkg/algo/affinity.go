package algo

import (
	corev1 "k8s.io/api/core/v1"
)

// AffinityQueue is used to sort pods by Affinity
type AffinityQueue struct {
	pods []*corev1.Pod
}

// NewAffinityQueue return a AffinityQueue
func NewAffinityQueue(pods []*corev1.Pod) *AffinityQueue {
	return &AffinityQueue{
		pods: pods,
	}
}

func (aff *AffinityQueue) Len() int      { return len(aff.pods) }
func (aff *AffinityQueue) Swap(i, j int) { aff.pods[i], aff.pods[j] = aff.pods[j], aff.pods[i] }
func (aff *AffinityQueue) Less(i, j int) bool {
	
	// NOTE: If the pod at index i has a non-nil node selector (i.e., it requires a specific node affinity), it's considered to have a higher
	// priority than the pod at index j. This is because a pod with a node selector specified likely has stricter requirements about where
	// it can be scheduled, so it should be placed higher in the scheduling queue.
	
	// NOTE 2: // NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	//            https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
	return aff.pods[i].Spec.NodeSelector != nil
}
