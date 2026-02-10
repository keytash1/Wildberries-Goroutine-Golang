package main

import (
	"fmt"
	"strings"
)

func checkUnic(s string) bool {
	//hashset
	mapa := make(map[rune]struct{}, len(s))
	for _, char := range strings.ToLower(s) {
		if _, ok1 := mapa[char]; ok1 {
			return false
		}
		mapa[char] = struct{}{}
	}
	return true
}

func main() {
	fmt.Println(checkUnic("abcd"), checkUnic("abCdefAaf"), checkUnic("aabcd"))
}

//done
