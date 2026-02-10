package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for num := range jobs {
		fmt.Println(id, " ", num)
	}
}

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT)

	number := flag.Int("workers", 3, "number of workers")
	flag.Parse()

	var wg sync.WaitGroup
	jobs := make(chan int, 10)

	for i := 0; i < *number; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	counter := 0
	for {
		select {
		//выбирается тот кейс операция которого может быть выполнена без блокировки
		case <-quit:
			close(jobs)
			wg.Wait()
			return

		case jobs <- counter:
			counter++
			time.Sleep(100 * time.Millisecond)
		}
	}
}

//выбрал канал тк это обеспечивает простое и гарантированное завершение воркеров после обработки уже отправленных данных
//перехватываем сигнал ОС об остановке и его уже можем переслать горутинам двумя(нет) обработать способами контекст и прерывающий канал
//done
