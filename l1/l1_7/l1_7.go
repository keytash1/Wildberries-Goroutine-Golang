package main

import "sync"

func main() {
	//fatal error: concurrent map writes

	// wg := sync.WaitGroup{}
	// wg.Add(10)
	// badMap := make(map[int]int)
	// for i := 0; i < 10; i++ {
	// 	go func() {
	// 		for j := 0; j < 100; j++ {
	// 			badMap[0] = i
	// 		}
	// 	}()
	// }
	// wg.Wait()

	//через мьютекс
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	wg.Add(10)
	badMap := make(map[int]int)
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				mu.Lock()
				badMap[0] = i
				mu.Unlock()
			}
		}(i)
		//в новых версиях с замыканием траблов нет
	}
	wg.Wait()

	//через sync.Map
	wg1 := sync.WaitGroup{}
	safeMap := sync.Map{}
	wg1.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg1.Done()
			for j := 0; j < 100; j++ {
				safeMap.Store(0, j)
			}
		}(i)
	}
	wg1.Wait()
}

//done
