package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Counter struct {
	num atomic.Int64
}

func (c *Counter) Increment() {
	c.num.Add(1)
}

func (c *Counter) Value() int64 {
	return c.num.Load()
}

func main() {
	wg := sync.WaitGroup{}
	c := Counter{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			c.Increment()
		}()
	}
	wg.Wait()
	fmt.Println(c.Value())
}

//done
