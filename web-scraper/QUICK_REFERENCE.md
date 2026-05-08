# Concurrency Quick Reference

Quick guide with practical examples for common concurrency scenarios.

## Quick Start Examples

### 1. Run N Tasks Concurrently (Your Project)
```go
var wg sync.WaitGroup
results := make([]Result, len(tasks))

for i, task := range tasks {
    wg.Add(1)
    go func(i int, task Task) {
        defer wg.Done()
        results[i] = process(task)
    }(i, task)
}

wg.Wait()
return results
```

**Use:** When you have independent tasks to process in parallel.

---

### 2. Limit Concurrent Operations (Semaphore)
```go
sem := make(chan struct{}, maxConcurrent)

for _, task := range tasks {
    sem <- struct{}{}  // Acquire
    go func(t Task) {
        defer func() { <-sem }()  // Release
        process(t)
    }(task)
}
```

**Use:** When you need to limit concurrency (your project does this!).

---

### 3. Timeout on Long Operation
```go
result := make(chan Result, 1)

go func() {
    result <- process()
}()

select {
case r := <-result:
    return r
case <-time.After(5 * time.Second):
    return errors.New("timeout")
}
```

**Use:** Prevent hanging on unresponsive operations.

---

### 4. Cancel All Goroutines
```go
done := make(chan struct{})
var wg sync.WaitGroup

wg.Add(1)
go func() {
    defer wg.Done()
    for {
        select {
        case <-done:
            return
        default:
            doWork()
        }
    }
}()

// Later, cancel
close(done)
wg.Wait()
```

**Use:** Graceful shutdown of goroutines.

---

### 5. Fire Multiple Requests, Get First Response
```go
results := make(chan Result, 1)
defer close(results)

go func() { results <- fetch(url1) }()
go func() { results <- fetch(url2) }()
go func() { results <- fetch(url3) }()

fastest := <-results  // First one wins!
```

**Use:** Get fastest response from redundant sources.

---

### 6. Process Results as They Arrive
```go
jobs := make(chan Job)
go func() {
    for job := range jobs {
        // Process each value as it arrives
        fmt.Println(job)
    }
    fmt.Println("All done!")
}()

jobs <- Job{1}
jobs <- Job{2}
close(jobs)  // Signal no more jobs
```

**Use:** Stream processing, event handling.

---

### 7. Rate Limiter
```go
limiter := time.Tick(100 * time.Millisecond)

for _, task := range tasks {
    <-limiter  // Wait until next tick
    go process(task)
}
```

**Use:** Control request rate to avoid overwhelming servers.

---

### 8. Merge Multiple Channels
```go
func merge(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    out := make(chan int)
    
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            for val := range c {
                out <- val
            }
            wg.Done()
        }(ch)
    }
    
    go func() {
        wg.Wait()
        close(out)
    }()
    
    return out
}
```

**Use:** Combine results from multiple sources.

---

## Common Patterns Reference

| Pattern | Code | Use Case |
|---------|------|----------|
| **Wait for completion** | `wg.Add(1); defer wg.Done(); wg.Wait()` | Run N tasks |
| **Limit concurrency** | `sem := make(chan struct{}, N)` | Bounded parallelism |
| **Timeout** | `select { case <-done: case <-time.After(...) }` | Prevent hangs |
| **Cancel** | `close(done)` channel | Stop all workers |
| **First response** | Buffered channel with `select` | Race requests |
| **Stream results** | `range channel` | Process as arrive |
| **Rate limit** | `time.Tick()` | Throttle requests |
| **Merge channels** | Loop with WaitGroup | Combine sources |

---

## Go Concurrency Decision Tree

```
START: Need to do something concurrently?
    |
    ├─ YES: Need to wait for completion?
    │   ├─ YES → Use sync.WaitGroup
    │   │   |
    │   │   ├─ Need to limit concurrency?
    │   │   │   ├─ YES → Use Semaphore (buffered channel)
    │   │   │   └─ NO → Plain goroutine + WaitGroup
    │   │   |
    │   │   └─ Done!
    │   |
    │   └─ NO: Fire and forget?
    │       └─ Make sure you understand why no wait needed!
    │
    └─ Need goroutines to communicate?
        ├─ YES → Use channels
        │   ├─ Send/receive simple data?
        │   │   ├─ YES → Regular channel
        │   │   └─ NO → Buffered channel
        │   │
        │   └─ Multiple sources to sync?
        │       └─ Use select
        |
        └─ NO → You don't need concurrency!
```

---

## Troubleshooting Guide

### Problem: Program hangs / Deadlock

**Suspect:** Unbuffered channel with no goroutine
```go
ch := make(chan int)
ch <- 42  // Hangs!
```

**Solutions:**
1. Use buffered channel: `make(chan int, 1)`
2. Launch receiver: `go func() { <-ch }()`

**Check:** Run with timeout, enable `-race` flag

---

### Problem: Results incomplete / Missing

**Suspect:** Returning before goroutines finish
```go
for i, task := range tasks {
    go func(t Task) {
        results[i] = process(t)
    }(task)
}
return results  // Too fast!
```

**Solution:** Wait with WaitGroup
```go
wg.Wait()
return results
```

**Check:** Use race detector: `go run -race .`

---

### Problem: Using wrong loop variable

**Suspect:** Closure over loop variable
```go
for i, task := range tasks {
    go func() {
        process(i, task)  // i might be wrong!
    }()
}
```

**Solution:** Pass as parameter
```go
go func(i int, t Task) {
    process(i, t)  // Safe copy
}(i, task)
```

---

### Problem: Memory leak / Resource exhaustion

**Suspect:** Not cleaning up resources
```go
for _, url := range urls {
    go func(u string) {
        resp, _ := http.Get(u)
        // resp.Body never closed!
    }(url)
}
```

**Solution:** Always defer cleanup
```go
defer resp.Body.Close()
```

---

## Testing Concurrent Code

### Check for Race Conditions
```bash
go run -race main.go
go test -race ./...
```

### Load Test
```go
const concurrency = 1000
const iterations = 100

for i := 0; i < concurrency; i++ {
    go func(id int) {
        for j := 0; j < iterations; j++ {
            testConcurr()
        }
    }(i)
}
```

### Timeout Test
```bash
timeout 30s go run -race main.go
```

---

## Performance Tips

1. **Goroutines are cheap** - Create thousands if needed
2. **Channels have overhead** - Use for coordination, not data passing
3. **Buffered channels reduce blocking** - Size = expected concurrent senders
4. **Mutexes are slower than channels** - Channels preferred
5. **sync.Once for one-time setup** - More efficient than mutex
6. **Atomic for simple counters** - Lower overhead than mutex

---

## Your Web Scraper Checklist

Your project implements good patterns. Verify:

```go
✅ sync.WaitGroup for coordination
✅ Buffered channel as semaphore for rate limiting
✅ Parameters passed to goroutines (no closure bugs)
✅ Results collected in array (order preserved)
✅ Errors stored in Result struct
✅ Resources cleaned up (defer resp.Body.Close())
✅ Timeouts configured (10 second timeout)
✅ No unbounded goroutine creation
```

Run these commands to verify:
```bash
# Check for race conditions
go run -race main.go https://www.example.com

# Run tests
go test -race ./...

# Build and verify
go build
```

---

## Further Learning Resources

### Official Documentation
- [Goroutines and Channels](https://go.dev/doc/effective_go#concurrency)
- [sync Package](https://pkg.go.dev/sync)
- [Context Package](https://pkg.go.dev/context)

### Important Concepts to Study
1. **Context** - Better than channels for cancellation
2. **Select** - Powerful for multiple channel operations
3. **Pipelines** - Efficient data processing
4. **Worker Pools** - Scalable concurrency

### Practice Exercise

Extend your web scraper:
1. Add context with timeout for the whole operation
2. Make results return a channel instead of array
3. Add progress reporting (X of N checked)
4. Implement graceful shutdown with signal handling

---

## One-Liners to Remember

```go
// Start goroutine
go doWork()

// Wait for all
var wg sync.WaitGroup
wg.Add(1)
defer wg.Done()
wg.Wait()

// Semaphore
sem := make(chan struct{}, maxConcurrent)
sem <- struct{}{}; defer func() { <-sem }()

// Timeout
select {
case result := <-ch: return result
case <-time.After(timeout): return err
}

// Cancel
done := make(chan struct{})
close(done)  // Unblocks all <-done

// Range until closed
for val := range ch { }

// First response
results := make(chan T, 1)
return <-results
```

---

## Summary

- **Goroutines**: Lightweight concurrency, started with `go`
- **WaitGroup**: Synchronize goroutine completion
- **Channels**: Goroutine communication (preferred over mutex)
- **Semaphore**: Limit concurrent execution with buffered channels
- **Select**: Coordinate multiple channel operations
- **Race Detector**: Always use `-race` to catch bugs

Your web scraper demonstrates production-quality concurrent Go code!
