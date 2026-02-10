package main

import "fmt"

func intersect(nums1 []int, nums2 []int) []int {
	mapa := make(map[int]int)
	for _, elem := range nums1 {
		mapa[elem] += 1
	}
	ans := make([]int, 0)
	for _, elem := range nums2 {
		if count, ok := mapa[elem]; ok {
			if count > 0 {
				ans = append(ans, elem)
				mapa[elem] -= 1
			}
		}
	}
	return ans
}

func main() {
	fmt.Println(intersect([]int{1, 2, 3}, []int{2, 3, 4}))
}

//done
