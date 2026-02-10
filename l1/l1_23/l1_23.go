package main

import "fmt"

func deleteI(arr []int, i int) []int {
	copy(arr[i:], arr[i+1:])
	return arr[:len(arr)-1]
}

func main() {
	fmt.Println(deleteI([]int{1, 2, 3}, 1))
}

//done
