package main

import (
	"fmt"
	"sync"
)

func main() {
	array := [5]int{2, 4, 6, 8, 10}
	var wg sync.WaitGroup
	for _, elem := range array {
		wg.Add(1)
		go func(elem int) {
			defer wg.Done()
			fmt.Println(elem * elem)
		}(elem)
	}
	wg.Wait()
}
