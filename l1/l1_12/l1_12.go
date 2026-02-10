package main

import "fmt"

func unicStrsMake(strs []string) []string {
	unicStrs := make([]string, 0)
	mapMeet := make(map[string]struct{})
	for i := 0; i < len(strs); i++ {
		if _, ok := mapMeet[strs[i]]; !ok {
			unicStrs = append(unicStrs, strs[i])
			mapMeet[strs[i]] = struct{}{}
		}
	}
	return unicStrs
}

func main() {
	strs := make([]string, 0)
	strs = append(strs, []string{"cat", "cat", "dog", "tree"}...)
	fmt.Println(unicStrsMake(strs))
}

//done
