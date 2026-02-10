package main

import "fmt"

func typeof(a interface{}) {
	switch a.(type) {
	case int:
		fmt.Println("int")
	case string:
		fmt.Println("string")
	case bool:
		fmt.Println("bool")
	case chan interface{}:
		fmt.Println("chan")
	}

}

func main() {
	integ := 12
	str := "12"
	boo := true
	c := make(chan interface{})
	typeof(integ)
	typeof(str)
	typeof(boo)
	typeof(c)
}

//done
