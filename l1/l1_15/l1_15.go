package main

var justString string

func createHugeString(n int) string {
	result := make([]byte, n)
	for i := range result {
		result[i] = 'a' + byte(i%26)
	}
	return string(result)
}

func someFunc() {
	v := createHugeString(1 << 10)
	justString = string([]byte(v[:100]))
	//gc сможет очистить v
}

// если нужна именно глобальная переменная а не возвращать
func main() {
	someFunc()
}

//done
