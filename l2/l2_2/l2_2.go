package main

import "fmt"

func test() (x int) { // 1. именованный возвращаемый параметр, default = 0
	defer func() {
		x++ // 3. defer меняет вовзращаемый параметр x = 2
	}()
	x = 1  // 2. параметр теперь x = 1
	return // после вызова return вызывается defer
}

func anotherTest() int { // неименованный возвращаемый параметр
	var x int // 1. x = 0
	defer func() {
		x++ // 3. defer меняет локальную переменную x, но return возвращает x на момент вызова (x = 1)
	}()
	x = 1    // 2. x = 1
	return x // return копирует переменную x на момент вызова в вовзращаемое значение, x = 1
} // после return выполняется defer, но возврат уже сделан

func main() {
	fmt.Println(test())        // 2
	fmt.Println(anotherTest()) // 1
}

//done
