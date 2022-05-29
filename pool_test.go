package priopool_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/alexvanin/priopool"
)

type syncList struct {
	mu   sync.Mutex
	list []int
}

func (s *syncList) append(i int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.list = append(s.list, i)
}

func TestPriorityPool_New(t *testing.T) {
	const pos, zero, neg = 1, 0, -1

	cases := [...]struct {
		pool, queue int
		err         error
	}{
		{pool: pos, queue: pos, err: nil},
		{pool: pos, queue: zero, err: nil},
		{pool: pos, queue: neg, err: nil},
		{pool: zero, queue: pos, err: priopool.ErrPoolCapacitySize},
		{pool: neg, queue: pos, err: priopool.ErrPoolCapacitySize},
	}

	for _, c := range cases {
		_, err := priopool.New(c.pool, c.queue)
		require.Equal(t, c.err, err, c)
	}
}

func TestPriorityPool_Submit(t *testing.T) {
	const (
		poolCap, queueCap = 2, 4
		totalCap          = poolCap + queueCap
		highPriority      = 10
		midPriority       = 5
		lowPriority       = 1
	)

	p, err := priopool.New(poolCap, queueCap)
	require.NoError(t, err)

	result := new(syncList)
	wg := new(sync.WaitGroup)
	wg.Add(totalCap)

	for i := 0; i < totalCap; i++ {
		var priority uint32

		switch {
		case i < poolCap:
			priority = midPriority // first put two middle priority tasks
		case i < poolCap+queueCap/2:
			priority = lowPriority // then add to queue two low priority tasks
		default:
			priority = highPriority // in the end fill queue with high priority tasks
		}

		err = p.Submit(priority, taskGenerator(int(priority), result, wg))
		require.NoError(t, err)
	}

	err = p.Submit(highPriority, func() {})
	require.Error(t, err, priopool.ErrQueueOverload)

	wg.Wait()

	expected := []int{
		midPriority, midPriority, // first tasks that took workers from pool
		highPriority, highPriority, // tasks from queue with higher priority
		lowPriority, lowPriority, // remaining tasks from queue
	}
	require.Equal(t, expected, result.list)

	t.Run("disabled queue", func(t *testing.T) {
		p, err := priopool.New(poolCap, 0)
		require.NoError(t, err)

		wg := new(sync.WaitGroup)
		wg.Add(poolCap)

		for i := 0; i < poolCap; i++ {
			id := i
			err = p.Submit(lowPriority, taskGenerator(id, nil, wg))
			require.NoError(t, err)
		}

		err = p.Submit(highPriority, func() {})
		require.Error(t, err, priopool.ErrQueueOverload)

		wg.Wait()
	})

	t.Run("disabled queue limit", func(t *testing.T) {
		const n = 50

		p, err := priopool.New(poolCap, -1)
		require.NoError(t, err)

		wg := new(sync.WaitGroup)
		wg.Add(n)

		for i := 0; i < n; i++ {
			id := i
			err = p.Submit(lowPriority, taskGenerator(id, nil, wg))
			require.NoError(t, err)
		}

		wg.Wait()
	})

	t.Run("non-priority order", func(t *testing.T) {
		const n = 5

		p, err := priopool.New(1, -1)
		require.NoError(t, err)

		wg := new(sync.WaitGroup)
		wg.Add(n)
		result := new(syncList)

		for i := 0; i < n; i++ {
			id := i
			err = p.Submit(lowPriority, taskGenerator(id, result, wg))
			require.NoError(t, err)
		}

		wg.Wait()

		fmt.Println(result.list)
		for i := 0; i < n; i++ {
			require.Equal(t, i, result.list[i])
		}
	})
}

func taskGenerator(ind int, output *syncList, wg *sync.WaitGroup) func() {
	return func() {
		time.Sleep(100 * time.Millisecond)
		if output != nil {
			output.append(ind)
		}
		wg.Done()
	}
}
