package main

import "fmt"

func main() {
	a := [5]int{76, 77, 78, 79, 80}
	// b - слайс с индекса 1 по 4 исходного массива a
	// slice b ссылается на массив а, одна область памяти. len(b) = 3 cap(b) = 4
	var b []int = a[1:4]
	fmt.Println(b) // [77 78 79]
}

//done
