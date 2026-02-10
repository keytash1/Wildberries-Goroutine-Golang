package main

import (
	"fmt"
	"math/rand"
	"time"
)

func binarySearch(c []int, found int) int {
	low := 0
	high := len(c) - 1
	for low <= high {
		center := (high + low) / 2
		if found == c[center] {
			return center
		}
		if found > c[center] {
			low = center + 1
		} else {
			high = center - 1
		}
	}
	return -1
}

func main() {
	rand.Seed(time.Now().UnixNano())
	c := make([]int, 0, 100)
	for i := range 100 {
		c = append(c, i)
	}
	found := c[rand.Intn(100)]
	for _, elem := range c {
		fmt.Print(elem, " ")
	}
	fmt.Println()
	fmt.Println(found)
	fmt.Println(binarySearch(c, found))
	fmt.Println(binarySearch(c, 100))
}

//done
