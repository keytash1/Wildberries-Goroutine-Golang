package main

func main() {
	ch := make(chan int) // небуферизированный канал

	go func() {
		for i := 0; i < 10; i++ {
			ch <- i // отправляем в канал 10 значений
		}
		// канал не закрывается
		// нужно добавить close(ch)
	}()

	for n := range ch { // range читает из канала пока он не закроется
		println(n) //  выводит числа от 0 до 9, затем deadlock all goroutines are sleep
	}
	// после 10 значений range ждет новые данные, но их никто не отправит
	// все горутины сптя - deadlock
}

//done
