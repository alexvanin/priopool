package priopool

import (
	"container/heap"
	"errors"
	"fmt"
	"sync"

	"github.com/panjf2000/ants/v2"
)

type (
	// PriorityPool is a pool of goroutines with priority queue buffer.
	// Based on panjf2000/ants and stdlib heap libraries.
	PriorityPool struct {
		mu    sync.Mutex // thread safe access to priority queue
		pool  *ants.Pool
		queue priorityQueue
		limit int
	}
)

const defaultQueueCapacity = 10

var (
	// ErrQueueOverload will be returned on submit operation
	// when both goroutines pool and priority queue are full.
	ErrQueueOverload = errors.New("pool and priority queue are full")

	// ErrPoolCapacitySize will be returned when constructor
	// provided with non-positive pool capacity.
	ErrPoolCapacitySize = errors.New("pool capacity must be positive")
)

// New creates instance of priority pool. Pool capacity must be positive.
// Zero queue capacity disables priority queue. Negative queue capacity
// disables priority queue length limit.
func New(poolCapacity, queueCapacity int) (*PriorityPool, error) {
	if poolCapacity <= 0 {
		return nil, ErrPoolCapacitySize
	}

	pool, err := ants.NewPool(poolCapacity, ants.WithNonblocking(true))
	if err != nil {
		return nil, fmt.Errorf("creating pool instance: %w", err)
	}

	var queue priorityQueue

	switch {
	case queueCapacity >= 0:
		queue.tasks = make([]*priorityQueueTask, 0, queueCapacity)
	case queueCapacity < 0:
		queue.tasks = make([]*priorityQueueTask, 0, defaultQueueCapacity)
	}

	return &PriorityPool{
		pool:  pool,
		queue: queue,
		limit: queueCapacity,
	}, nil
}

// Submit sends the task into priority pool. Non-blocking operation. If pool has
// available workers, then task executes immediately. If pool is full, then task
// is stored in priority queue. It will be executed when available worker pops
// the task from priority queue. Tasks from queue do not evict running tasks
// from pool. Tasks with bigger priority number are popped earlier.
// If queue is full, submit returns ErrQueueOverload error.
func (p *PriorityPool) Submit(priority uint32, task func()) error {
	p.mu.Lock() // lock from the beginning to avoid starving
	defer p.mu.Unlock()

	err := p.pool.Submit(func() {
		task()

		// pick the highest priority item from the queue
		// process items until queue is empty
		for {
			p.mu.Lock()
			if p.queue.Len() == 0 {
				p.mu.Unlock()
				return
			}
			queueF := heap.Pop(&p.queue)
			p.mu.Unlock()

			queueF.(*priorityQueueTask).value()
		}
	})
	if err == nil {
		return nil
	}

	if !errors.Is(err, ants.ErrPoolOverload) {
		return fmt.Errorf("pool submit: %w", err)
	}

	if p.limit >= 0 && p.queue.Len() >= p.limit {
		return ErrQueueOverload
	}

	heap.Push(&p.queue, &priorityQueueTask{
		value:    task,
		priority: int(priority),
	})

	return nil
}
