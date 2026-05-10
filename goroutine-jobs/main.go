package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	ch := make(chan int)
	for i := 1; i <= 5; i++ {
		go func(id int) {
			// fake job with random sleep
			time.Sleep(time.Duration(rand.Intn(5)+1) * time.Second)
			ch <- id
		}(i)
	}
	for i := 0; i < 5; i++ {
		result := <-ch
		fmt.Printf("Goroutine %d finished\n", result)
	}
}
