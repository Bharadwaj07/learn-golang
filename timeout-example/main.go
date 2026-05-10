package main

import (
	"fmt"
	"time"
)

func main() {
	ch := make(chan string)
	go func() {
		time.Sleep(5 * time.Second)
		ch <- "work done"
	}()
	select {
	case result := <-ch:
		fmt.Println(result)
	case <-time.After(2 * time.Second):
		fmt.Println("timed out")
	}
}
