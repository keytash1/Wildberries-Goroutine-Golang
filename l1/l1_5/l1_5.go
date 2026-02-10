package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

func Recieve(ch <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for nums := range ch {
		fmt.Println(nums)
	}
}

func main() {
	n := flag.Int("seconds", 10, "working time")
	flag.Parse()
	ch := make(chan int, 10)
	i := 0
	wg := &sync.WaitGroup{}
	timer := time.After(time.Duration(*n) * time.Second)
	wg.Add(1)
	go Recieve(ch, wg)
	for {
		select {
		case <-timer:
			close(ch)
			wg.Wait()
			return
		case ch <- i:
			i++
			time.Sleep(1 * time.Second)
		}
	}
}

//done
