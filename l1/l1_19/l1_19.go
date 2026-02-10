package main

import "fmt"

func reverseString(str string) string {
	runes := []rune(str)
	for i := 0; i < len(runes)/2; i++ {
		runes[i], runes[len(runes)-1-i] = runes[len(runes)-1-i], runes[i]
	}
	return string(runes)
}

func main() {
	str := "АБВГД"
	fmt.Println(reverseString(str))
}

//done
