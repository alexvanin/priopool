<p align="center">
<img src="/.github/logo.svg" width="500px">
</p>
<p align="center">
  Goroutines pool with priority queue buffer.
</p>

---

## Overview

Package `priopool` provides goroutines pool based on
[panjf2000/ants](https://github.com/panjf2000/ants) library with priority queue 
buffer based on [stdlib heap](https://pkg.go.dev/container/heap) package.

Priority pool:
- is non-blocking,
- prioritizes tasks with higher priority value,
- can be configured with unlimited queue buffer.

## Install

```
go get -u github.com/alexvanin/priopool
```

## Example

```go
package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alexvanin/priopool"
)

func main() {
	regularJob := func(i int) {
		time.Sleep(1 * time.Second)
		fmt.Printf("Job %d is done\n", i)
	}

	highPriorityJob := func() {
		fmt.Println("High priority job is done")
	}

	pool, err := priopool.New(2, -1) // pool for two parallel executions
	if err != nil {
		log.Fatal(err)
	}

	wg := new(sync.WaitGroup)
	wg.Add(5 + 1)
	for i := 0; i < 5; i++ {
		ind := i + 1
		// enqueue 5 regular jobs
		pool.Submit(1, func() { regularJob(ind); wg.Done() })
	}
	// after 5 regular jobs enqueue high priority job
	pool.Submit(10, func() { highPriorityJob(); wg.Done() })
	wg.Wait()

	/*
		Output:
		Job 2 is done
		Job 1 is done
		High priority job is done
		Job 4 is done
		Job 3 is done
		Job 5 is done
	*/
}
```

## License

Source code is available under the [MIT License](/LICENSE).
