package main

import (
	"fmt"
)

func bitOp(n int64, i int, op bool) int64 {
	mask := int64(1)<<i - 1
	if op {
		return n | mask
	}
	return n &^ mask
}

// нумерация битов с 1 справа
func main() {
	n := int64(5)
	fmt.Println(bitOp(n, 1, false))
}

//done
