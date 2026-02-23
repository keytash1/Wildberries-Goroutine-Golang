package main

type customError struct {
	msg string
}

// customError соответствует интерфейсу error
func (e *customError) Error() string {
	return e.msg
}

func test() *customError {
	// ... do something
	return nil
	//это как:
	//var x *customError // x = nil
	//return x           // возвращаем nil указатель
}

func main() {
	var err error // err: (nil, nil) - тип, значение
	err = test()  // err: (*customError, nil)
	// интерфейс = nil если и тип и значение nil
	if err != nil { // проверяет не значение а всю структуру (тип+значение)
		println("error") // выводится "error" тк тип != nil
		return
	}
	println("ok")
}

//done
