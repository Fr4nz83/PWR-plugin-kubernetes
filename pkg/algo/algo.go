package algo


// This interface defines the requirements of a sorting mechanism for pods in a scheduling queue.
type SchedulingQueueSort interface {
	Len() int
	Swap(i, j int)
	Less(i, j int) bool
}
