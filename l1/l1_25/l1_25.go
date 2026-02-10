package main

import (
	"time"
)

func sleep(duration time.Duration) {
	timer := time.NewTimer(duration)
	<-timer.C
	// or <-time.After(duration)
}

func main() {
	sleep(5 * time.Second)
}

//done
