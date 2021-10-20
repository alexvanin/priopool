package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"
)

type (
	PriorityPool struct {
		mu    sync.Mutex // PriorityQueue is not thread safe
		pool  *ants.Pool
		queue PriorityQueue
		limit int
	}
)

const (
	PoolCapacity  = 2
	QueueCapacity = 4
)

func (p *PriorityPool) Submit(pri int, f func()) error {
	err := p.pool.Submit(func() {
		f()

		for {
			// pick the highest priority item from the queue if there is any
			p.mu.Lock()
			if p.queue.Len() == 0 {
				p.mu.Unlock()
				return
			}
			queueF := p.queue.Pop()
			p.mu.Unlock()

			queueF.(*Item).value()
		}
	})
	if err == nil {
		return nil
	}

	if !errors.Is(err, ants.ErrPoolOverload) {
		return err
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	ln := p.queue.Len()
	if ln >= QueueCapacity {
		return errors.New("queue capacity is full")
	}

	p.queue.Push(&Item{
		value:    f,
		priority: pri,
		index:    ln,
	})

	return nil
}

func NewPriorityPool() (*PriorityPool, error) {
	pool, err := ants.NewPool(PoolCapacity, ants.WithNonblocking(true))
	if err != nil {
		return nil, fmt.Errorf("new pool: %w", err)
	}

	return &PriorityPool{
		pool:  pool,
		queue: make(PriorityQueue, 0, QueueCapacity),
		limit: QueueCapacity,
	}, nil
}

func main() {
	p, err := NewPriorityPool()
	if err != nil {
		log.Fatal(err)
	}

	wg := new(sync.WaitGroup)
	wg.Add(QueueCapacity + PoolCapacity)

	for i := 0; i < QueueCapacity+PoolCapacity+1; i++ {
		if i < 4 {
			err = p.Submit(1, func() {
				time.Sleep(1 * time.Second)
				fmt.Println("Low priority task is done")
				wg.Done()
			})
			fmt.Printf("id:%d <low priority task> error:%v\n", i+1, err)
		} else if i < 6 {
			err = p.Submit(10, func() {
				time.Sleep(1 * time.Second)
				fmt.Println("High priority task is done")
				wg.Done()
			})
			fmt.Printf("id:%d <high priority task> error:%v\n", i+1, err)
		} else {
			err = p.Submit(10, func() {})
			fmt.Printf("id:%d <out of capacity task> error:%v\n", i+1, err)
		}
	}
	fmt.Println()
	wg.Wait()

	/*
		id:1 <low priority task> error:<nil>
		id:2 <low priority task> error:<nil>
		id:3 <low priority task> error:<nil>
		id:4 <low priority task> error:<nil>
		id:5 <high priority task> error:<nil>
		id:6 <high priority task> error:<nil>
		id:7 <out of capacity task> error:queue capacity is full

		Low priority task is done
		Low priority task is done
		High priority task is done
		High priority task is done
		Low priority task is done
		Low priority task is done
	*/
}
