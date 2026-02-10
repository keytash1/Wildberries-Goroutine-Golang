package main

import "fmt"

func groupTemps(temps []float64) map[int][]float64 {
	ans := make(map[int][]float64)
	for _, temp := range temps {
		currInterval := int(temp) / 10 * 10
		ans[int(currInterval)] = append(ans[int(currInterval)], temp)
	}
	return ans
}

func main() {
	ans := groupTemps([]float64{-25.4, -27.0, 13.0, 19.0, 15.5, 24.5, -21.0, 32.5})
	for key, value := range ans {
		fmt.Println(key, value)
	}
}

//done
