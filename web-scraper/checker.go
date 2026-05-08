package main

import (
	"net/http"
	"sync"
	"time"
)

// Result holds the outcome of checking a single link
type Result struct {
	URL    string
	Status int
	Err    error
}

func (r Result) IsAlive() bool {
	return r.Err == nil && r.Status < 400
}

// checkLinks checks all links concurrently and returns the results.
// maxConcurrent controls how many requests run at the same time.
func checkLinks(links []string, maxConcurrent int) []Result {
	// a buffered channel acts as a semaphore —
	// it limits how many goroutines run at once
	sem := make(chan struct{}, maxConcurrent)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	results := make([]Result, len(links))
	var wg sync.WaitGroup

	for i, link := range links {
		wg.Add(1)

		go func(i int, link string) {
			defer wg.Done()

			sem <- struct{}{}        // acquire slot
			defer func() { <-sem }() // release slot when done

			results[i] = check(client, link)
		}(i, link)
	}

	wg.Wait() // block until all goroutines finish
	return results
}

func check(client *http.Client, link string) Result {
	resp, err := client.Head(link)
	if err != nil {
		return Result{URL: link, Err: err}
	}
	defer resp.Body.Close()
	return Result{URL: link, Status: resp.StatusCode}
}
