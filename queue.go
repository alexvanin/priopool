package priopool

// Priority pool based on implementation example from heap package.
// Priority queue itself is not thread safe.
// See https://cs.opensource.google/go/go/+/refs/tags/go1.17.2:src/container/heap/example_pq_test.go

import (
	"container/heap"
)

type priorityQueueTask struct {
	value    func()
	priority int
	index    int // the index is needed by update and is maintained by the heap.Interface methods
}

type priorityQueue []*priorityQueueTask

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].priority > pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*priorityQueueTask)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

func (pq *priorityQueue) update(item *priorityQueueTask, value func(), priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}
