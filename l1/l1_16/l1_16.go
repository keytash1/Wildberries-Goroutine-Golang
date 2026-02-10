package main

import (
	"fmt"
	"math/rand"
	"time"
)

func quickSort(arr []int) []int {
	if len(arr) < 2 {
		return arr
	}
	pivot := arr[len(arr)/2]
	var left, right, middle []int
	for _, x := range arr {
		if x < pivot {
			left = append(left, x)
		} else if x > pivot {
			right = append(right, x)
		} else {
			middle = append(middle, x)
		}
	}
	left = quickSort(left)
	right = quickSort(right)
	return append(append(left, middle...), right...)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	c := make([]int, 0, 100)
	for range 100 {
		c = append(c, rand.Intn(100)+1)
	}
	fmt.Println(quickSort(c))
}

//done
