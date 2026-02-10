package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for num := range jobs {
		fmt.Println(id, " ", num)
	}
}

func main() {
	number := flag.Int("workers", 3, "number of workers")
	flag.Parse()

	var wg sync.WaitGroup
	jobs := make(chan int, 10)

	for i := 0; i < *number; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	counter := 0
	for i := 0; i < 100000; i++ {
		jobs <- counter
		counter++
		time.Sleep(100 * time.Millisecond)
	}
	close(jobs)
	wg.Wait()
}

//done
