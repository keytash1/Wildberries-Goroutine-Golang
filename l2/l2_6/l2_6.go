package main

import (
	"fmt"
)

func main() {
	var s = []string{"1", "2", "3"} // s: len=3 cap=3 data=[1 2 3]
	modifySlice(s)
	fmt.Println(s) // [3 2 3]
}

func modifySlice(i []string) { // i копирует len cap и указатель на массив
	i[0] = "3"         //меняет исходный массив; i=s: len = 3 cap = 3 data=[3 2 3]
	i = append(i, "4") // i: len=4 cap=6 (из-за того что cap вырос создается новый массив, туда копируются элементы старого и i теперь хранит ссылку на новый массив) data=[3 2 3 4]
	i[1] = "5"         //меняем новый массив i: len=4 cap=6 data=[3 5 3 4]
	i = append(i, "6") //i: len=6 cap=6 data=[3 5 3 4 6]
} //исходный s в main: поменялся только в первой строчке манипуляций с i когда они еще ссылались на одну область данных len=3 cap=3 data=[3 2 3 ](сap исходного маасива не меняется)

//done
