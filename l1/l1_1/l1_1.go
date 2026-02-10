package main

import "fmt"

type Human struct {
	name string
	age  int
}

func NewHuman(name string, age int) *Human {
	return &Human{
		name: name,
		age:  age,
	}
}

func (h *Human) Name() string {
	return h.name
}

func (h *Human) Age() int {
	return h.age
}

func (h *Human) String() string {
	return fmt.Sprintf("%s (%d years old)", h.name, h.age)
}

type Action struct {
	Human
	left  int
	right int
}

func main() {
	human := NewHuman("Anton", 19)
	action := Action{Human: *human,
		left:  1,
		right: 2}
	fmt.Println(action.Name())
}

//done
