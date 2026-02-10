package main

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

func main() {
	//0 выход по условию
	i := 0
	go func() {
		for {
			if i < 10 {
				fmt.Println(0, i)
				i++
			} else {
				return
			}
		}
	}()
	time.Sleep(time.Second * 3)

	// 1 канал уведомлений
	i = 0
	quit := make(chan int) //или буф?
	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				fmt.Println(1, i)
				i++
			}
		}
	}()
	//дал поработать завершил подождал
	time.Sleep(time.Second * 1)
	quit <- 10
	time.Sleep(time.Second * 3)

	// 2 контекст
	i = 0
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fmt.Println(2, i)
				i++
			}
		}
	}()
	time.Sleep(time.Second * 1)
	cancel()
	time.Sleep(time.Second * 3)

	// 3 Goexit
	i = 0
	go func() {
		for {
			if i < 10 {
				fmt.Println(3, i)
				i++
			} else {
				runtime.Goexit()
			}
		}
	}()
	time.Sleep(time.Second * 3)

	// 4 TimeoutContext
	ctxT, _ := context.WithTimeout(context.Background(), 2*time.Second)
	go func() {
		for {
			select {
			case <-ctxT.Done():
				return
			default:
				fmt.Println(4, i)
				i++
			}
		}
	}()
	time.Sleep(time.Second * 1)
	// 5 close channel
	data := make(chan int)
	go func() {
		for v := range data {
			fmt.Println(5, v)
		}
		fmt.Println(5)
	}()
	for i := 0; i < 10; i++ {
		data <- i
	}
	close(data)
	time.Sleep(1 * time.Second)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println(r)
			}
		}()
		panic("Паника")
	}()
}

//done
