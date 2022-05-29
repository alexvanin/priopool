package priopool

// Priority pool based on implementation example from heap package.
// Priority queue itself is not thread safe.
// See https://cs.opensource.google/go/go/+/refs/tags/go1.17.2:src/container/heap/example_pq_test.go

type priorityQueueTask struct {
	value    func()
	priority int
	index    uint64 // monotonusly increasing index to sort values with same priority
}

type priorityQueue struct {
	nextIndex uint64
	tasks     []*priorityQueueTask
}

func (pq priorityQueue) Len() int { return len(pq.tasks) }

func (pq priorityQueue) Less(i, j int) bool {
	if pq.tasks[i].priority == pq.tasks[j].priority {
		return pq.tasks[i].index < pq.tasks[j].index
	}
	return pq.tasks[i].priority > pq.tasks[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq.tasks[i], pq.tasks[j] = pq.tasks[j], pq.tasks[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*priorityQueueTask)
	item.index = pq.nextIndex
	pq.nextIndex++
	pq.tasks = append(pq.tasks, item)
}

func (pq *priorityQueue) Pop() interface{} {
	n := len(pq.tasks)
	item := pq.tasks[n-1]
	pq.tasks[n-1] = nil
	pq.tasks = pq.tasks[0 : n-1]
	return item
}
