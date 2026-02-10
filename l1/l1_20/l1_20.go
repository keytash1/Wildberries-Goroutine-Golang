package main

import "fmt"

func reverseString(str string) string {
	runes := []byte(str)
	for i := 0; i < len(runes)/2; i++ {
		runes[i], runes[len(runes)-1-i] = runes[len(runes)-1-i], runes[i]
	}
	return string(runes)
}

func reverseWorldsInString(s string) string {
	s = reverseString(s)
	lastSpace := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			s = s[:lastSpace] + reverseString(s[lastSpace:i]) + s[i:]
			lastSpace = i + 1
		}
	}
	if lastSpace < len(s) {
		s = s[:lastSpace] + reverseString(s[lastSpace:])
	}
	return s
}

func main() {
	fmt.Println(reverseWorldsInString("ABAB ABAB ABAB ABAB"))
}

//done
