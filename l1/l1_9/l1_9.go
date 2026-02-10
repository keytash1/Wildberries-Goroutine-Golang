package main

import (
	"fmt"
	"sync"
)

func double(in <-chan int, out chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := range in {
		out <- i * 2
	}
	close(out)
}

func print(out <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for i := range out {
		fmt.Println(i)
	}
}

func main() {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	in := make(chan int, 10)
	out := make(chan int, 10)
	go double(in, out, wg)
	go print(out, wg)
	for i := 0; i < 10; i++ {
		in <- i
	}
	close(in)
	wg.Wait()
}

//done
