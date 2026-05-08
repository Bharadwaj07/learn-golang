# Go Concurrency Guide

This guide covers the concurrency patterns used in your web scraper project.

## Table of Contents
1. [Goroutines](#goroutines)
2. [Channels](#channels)
3. [WaitGroups](#waitgroups)
4. [Semaphores](#semaphores)
5. [Race Conditions](#race-conditions)
6. [Best Practices](#best-practices)

---

## Goroutines

A goroutine is a lightweight thread managed by the Go runtime. They're much cheaper than OS threads and allow you to run thousands concurrently.

### Creating a Goroutine
```go
go functionName()        // Launch in background
go func() { }()          // Launch anonymous function
```

### In Your Project
```go
go func(i int, link string) {
    // This runs concurrently in the background
    results[i] = check(client, link)
}(i, link)
```

### Key Points
- Goroutines don't have IDs or handles - you can't control them directly
- Started with `go` keyword
- Program exits when main() ends, even if goroutines are still running
- Use WaitGroups to ensure goroutines complete before main exits

---

## Channels

Channels enable safe communication between goroutines. They allow goroutines to send and receive values.

### Channel Types

**Unbuffered Channel** (synchronous):
```go
ch := make(chan int)
ch <- 42              // Sender blocks until receiver is ready
value := <-ch         // Receiver blocks until sender sends
```

**Buffered Channel** (asynchronous):
```go
ch := make(chan int, 5)  // Buffer capacity 5
ch <- 42                 // Non-blocking if buffer has space
value := <-ch            // Blocks only if buffer empty
```

### In Your Project
```go
sem := make(chan struct{}, maxConcurrent)  // Buffered channel
sem <- struct{}{}                          // Send (acquire)
<-sem                                      // Receive (release)
```

### Channel Operations
| Operation | Unbuffered | Buffered |
|-----------|-----------|----------|
| Send | Blocks until received | May block if full |
| Receive | Blocks until sent | May block if empty |
| Close | Safe if no blocked senders | Safe if all senders done |

---

## WaitGroups

`sync.WaitGroup` synchronizes goroutines. It waits for a collection of goroutines to finish executing.

### API
```go
var wg sync.WaitGroup

wg.Add(n)        // Add n goroutines to wait for
wg.Done()        // Signal one goroutine is done (decrement)
wg.Wait()        // Block until all goroutines call Done()
```

### In Your Project
```go
var wg sync.WaitGroup

for i, link := range links {
    wg.Add(1)                    // Increment counter
    
    go func(i int, link string) {
        defer wg.Done()           // Always decrement (even on error)
        results[i] = check(client, link)
    }(i, link)
}

wg.Wait()  // Block until all goroutines finish
return results
```

### Why `defer wg.Done()` is important
- Ensures `Done()` is called even if goroutine panics
- Clean, idiomatic Go pattern
- Prevents race conditions where main returns before work completes

---

## Semaphores

A semaphore is a concurrency primitive that limits resource access. In Go, we use buffered channels as semaphores.

### How It Works
```go
sem := make(chan struct{}, maxConcurrent)

// Acquire slot (blocks if all slots taken)
sem <- struct{}{}

// Do work...

// Release slot
<-sem
```

### In Your Project
```go
sem := make(chan struct{}, maxConcurrent)  // maxConcurrent slots available

go func(i int, link string) {
    sem <- struct{}{}                // Acquire - blocks if full
    defer func() { <-sem }()         // Release - always runs
    
    results[i] = check(client, link) // Only maxConcurrent run simultaneously
}(i, link)
```

### Why Use Semaphores?
1. **Rate limiting**: Prevent overwhelming target servers
2. **Resource control**: Limit open connections
3. **Avoid IP bans**: Serve responsible requests
4. **Prevent errors**: Avoid socket exhaustion

### Without Semaphore (Bad)
```go
// ALL 10,000 goroutines start HTTP requests at once!
for _, link := range links {
    go func(link string) {
        check(client, link)  // Uncontrolled concurrency
    }(link)
}
```

### With Semaphore (Good)
```go
// Only 5 requests at a time
sem := make(chan struct{}, 5)
for _, link := range links {
    go func(link string) {
        sem <- struct{}{}
        defer func() { <-sem }()
        check(client, link)
    }(link)
}
```

---

## Race Conditions

A race condition occurs when multiple goroutines access shared data simultaneously and at least one modifies it.

### Example of a Race Condition (BAD)
```go
var count int

go func() { count++ }()      // Goroutine 1: increment
go func() { count++ }()      // Goroutine 2: increment
```

The final value of `count` is unpredictable! Could be 1 or 2 depending on timing.

### In Your Project (SAFE)
```go
results := make([]Result, len(links))

go func(i int, link string) {
    results[i] = check(client, link)  // Each goroutine writes to different index
}(i, link)
```

This is safe because:
- Each goroutine writes to unique index `i`
- No concurrent reads/writes to same memory
- Array is only read after `wg.Wait()`

### Detecting Race Conditions
Run your program with `-race` flag:
```bash
go run -race .
```

This adds runtime checks to detect race conditions.

---

## Best Practices

### 1. Always Use WaitGroups
```go
// GOOD
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    // work
}()
wg.Wait()
```

### 2. Defer Close Operations
```go
// GOOD
resp, _ := http.Get(url)
defer resp.Body.Close()

// BAD
resp, _ := http.Get(url)
// resp.Body not closed!
```

### 3. Use Buffered Channels for Semaphores
```go
// GOOD - Limits concurrency
sem := make(chan struct{}, maxConcurrent)

// BAD - Unbuffered channel (only one at a time)
sem := make(chan struct{})
```

### 4. Pass Parameters to Goroutines
```go
// GOOD - Copy values to goroutine
for i, link := range links {
    go func(i int, link string) {
        // i and link are copies
    }(i, link)
}

// BAD - Closure over loop variables
for i, link := range links {
    go func() {
        // i and link may change before goroutine runs!
    }()
}
```

### 5. Check for Race Conditions
```bash
go run -race .
go test -race ./...
```

### 6. Set Timeouts on HTTP Clients
```go
// GOOD
client := &http.Client{
    Timeout: 10 * time.Second,
}

// BAD
client := &http.Client{}  // No timeout - can hang forever
```

### 7. Handle Errors in Goroutines
```go
// GOOD
results := make([]Result, len(links))
go func(i int, link string) {
    result := check(client, link)
    results[i] = result  // Includes error if any
}(i, link)

// BAD
go func(link string) {
    check(client, link)  // Errors are ignored!
}(link)
```

---

## Common Concurrency Patterns

### Pattern 1: Fan-Out/Fan-In
Multiple goroutines execute the same task (fan-out), then results are collected (fan-in).

```go
// Fan-out: Start multiple goroutines
results := make([]Result, len(links))
var wg sync.WaitGroup
for i, link := range links {
    wg.Add(1)
    go func(i int, link string) {
        defer wg.Done()
        results[i] = process(link)  // Fan-out
    }(i, link)
}

// Fan-in: Wait for all and collect results
wg.Wait()
return results  // Collected results
```

Your web scraper uses this pattern!

### Pattern 2: Worker Pool
A fixed number of workers process tasks from a queue.

```go
jobs := make(chan string, 100)
results := make(chan Result, 100)

// Start N workers
numWorkers := 5
for i := 0; i < numWorkers; i++ {
    go worker(jobs, results)
}

// Send jobs
for _, link := range links {
    jobs <- link
}
close(jobs)

// Collect results
for i := 0; i < len(links); i++ {
    result := <-results
    // process result
}
```

### Pattern 3: Timeout Handler
Use time.After() to set timeouts.

```go
select {
case result := <-resultChan:
    return result
case <-time.After(5 * time.Second):
    return errors.New("timeout")
}
```

---

## Summary

Your web scraper successfully uses:
- ✅ Goroutines for concurrent execution
- ✅ WaitGroups for synchronization
- ✅ Semaphores (buffered channels) for concurrency control
- ✅ Safe memory access (no race conditions)
- ✅ Proper error handling
- ✅ Resource cleanup (`defer resp.Body.Close()`)
- ✅ Timeouts on HTTP client

This is a solid, production-ready concurrency pattern!
