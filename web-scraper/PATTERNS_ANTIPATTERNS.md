# Concurrency Patterns & Anti-Patterns

This document covers proven patterns and common mistakes to avoid.

## Good Patterns

### Pattern 1: Fan-Out/Fan-In

Multiple goroutines process data in parallel, results collected afterward.

**Your web scraper uses this pattern:**
```go
// Fan-out: Launch all goroutines
func checkLinks(links []string, maxConcurrent int) []Result {
    results := make([]Result, len(links))
    var wg sync.WaitGroup
    sem := make(chan struct{}, maxConcurrent)
    
    for i, link := range links {
        wg.Add(1)
        go func(i int, link string) {  // Fan-out
            defer wg.Done()
            sem <- struct{}{}
            defer func() { <-sem }()
            results[i] = check(client, link)
        }(i, link)
    }
    
    wg.Wait()  // Fan-in: Collect all results
    return results
}
```

**Advantages:**
- ✅ Simple to understand
- ✅ Good for independent tasks
- ✅ Results available in order (array indexing)
- ✅ Easy error propagation (store in Result struct)

**When to use:**
- Map operations (apply function to each element)
- Web scraping (check each URL)
- Batch processing
- Independent tasks

---

### Pattern 2: Worker Pool

Fixed number of workers process tasks from a queue.

```go
func workerPool(urls []string, numWorkers int) []Result {
    jobs := make(chan string, len(urls))
    results := make(chan Result)
    var wg sync.WaitGroup
    
    // Start workers
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for job := range jobs {
                results <- check(client, job)
            }
        }()
    }
    
    // Send jobs
    go func() {
        for _, url := range urls {
            jobs <- url
        }
        close(jobs)
    }()
    
    // Collect results
    var allResults []Result
    go func() {
        wg.Wait()
        close(results)
    }()
    
    for result := range results {
        allResults = append(allResults, result)
    }
    
    return allResults
}
```

**Advantages:**
- ✅ Reuses goroutines (efficient)
- ✅ Works with infinite streams
- ✅ Easy to scale workers
- ✅ Natural error handling

**When to use:**
- Stream processing
- Handling requests from queue
- Long-running services
- When data arrives over time

---

### Pattern 3: Pipeline

Stages connected by channels, each processing data.

```go
// Stage 1: Generate URLs
func generateURLs(urls []string) <-chan string {
    out := make(chan string)
    go func() {
        defer close(out)
        for _, url := range urls {
            out <- url
        }
    }()
    return out
}

// Stage 2: Fetch and parse
func fetchAndParse(in <-chan string) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for url := range in {
            out <- check(client, url)
        }
    }()
    return out
}

// Stage 3: Filter alive
func filterAlive(in <-chan Result) <-chan Result {
    out := make(chan Result)
    go func() {
        defer close(out)
        for result := range in {
            if result.IsAlive() {
                out <- result
            }
        }
    }()
    return out
}

// Usage
func main() {
    urls := []string{"url1", "url2", ...}
    urlsChan := generateURLs(urls)
    results := fetchAndParse(urlsChan)
    aliveOnly := filterAlive(results)
    
    for result := range aliveOnly {
        fmt.Println(result)
    }
}
```

**Advantages:**
- ✅ Highly composable
- ✅ Each stage is independent
- ✅ Easy to test
- ✅ Good for streaming data

**When to use:**
- Data transformation pipelines
- Multi-step processing
- Composable operations
- Streaming data

---

### Pattern 4: Timeout & Cancellation

Cancel long-running operations.

```go
// With timeout
func checkWithTimeout(url string, timeout time.Duration) (Result, error) {
    done := make(chan Result, 1)
    
    go func() {
        done <- check(client, url)
    }()
    
    select {
    case result := <-done:
        return result, nil
    case <-time.After(timeout):
        return Result{}, errors.New("timeout")
    }
}

// With context (better)
func checkWithContext(ctx context.Context, url string) (Result, error) {
    done := make(chan Result, 1)
    
    go func() {
        done <- check(client, url)
    }()
    
    select {
    case result := <-done:
        return result, nil
    case <-ctx.Done():
        return Result{}, ctx.Err()  // Respects parent timeout
    }
}

// Usage with context
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

result, err := checkWithContext(ctx, "https://example.com")
if err != nil {
    fmt.Println("Failed:", err)
}
```

---

### Pattern 5: Graceful Shutdown

Stop gracefully, allowing cleanup.

```go
type Server struct {
    done chan struct{}
    wg   sync.WaitGroup
}

func (s *Server) Start() {
    s.wg.Add(1)
    go func() {
        defer s.wg.Done()
        for {
            select {
            case <-s.done:
                return  // Stop when signaled
            default:
                s.handleRequest()
            }
        }
    }()
}

func (s *Server) Stop() {
    close(s.done)      // Signal all goroutines
    s.wg.Wait()        // Wait for completion
    fmt.Println("Server stopped")
}

// Usage
server := &Server{done: make(chan struct{})}
server.Start()

time.Sleep(1 * time.Second)
server.Stop()  // Cleanup before exit
```

---

## Anti-Patterns (What NOT to Do)

### Anti-Pattern 1: Fire and Forget (No Wait)

❌ **WRONG:**
```go
for i, link := range links {
    go func(link string) {
        results[i] = check(client, link)  // Race condition!
    }(link)
}
return results  // May return before any goroutines run!
```

**Problems:**
- Results incomplete
- `i` changes before goroutine reads it
- Function returns immediately

✅ **CORRECT:**
```go
var wg sync.WaitGroup
for i, link := range links {
    wg.Add(1)
    go func(i int, link string) {
        defer wg.Done()
        results[i] = check(client, link)
    }(i, link)
}
wg.Wait()
return results
```

---

### Anti-Pattern 2: Not Passing Parameters to Goroutine

❌ **WRONG:**
```go
for i, link := range links {
    go func() {
        // i and link are shared! Loop continues before goroutine runs
        results[i] = check(client, link)
    }()
}
```

**Problem:** All goroutines might see final values of `i` and `link`.

✅ **CORRECT:**
```go
for i, link := range links {
    go func(i int, link string) {
        // i and link are copied to goroutine
        results[i] = check(client, link)
    }(i, link)  // Pass parameters!
}
```

---

### Anti-Pattern 3: Unbounded Goroutine Creation

❌ **WRONG:**
```go
for _, task := range millionTasks {
    go processTask(task)  // Allocates 1 million goroutines!
}
// Out of memory or system overload
```

**Problem:** Each goroutine uses memory. 1 million goroutines = resource exhaustion.

✅ **CORRECT:**
```go
sem := make(chan struct{}, numWorkers)  // Limit concurrency
for _, task := range millionTasks {
    sem <- struct{}{}
    go func(t Task) {
        defer func() { <-sem }()
        processTask(t)
    }(task)
}
```

---

### Anti-Pattern 4: Ignoring Errors in Goroutines

❌ **WRONG:**
```go
go func(link string) {
    result, err := makeRequest(link)
    if err != nil {
        return  // Error silently lost!
    }
    results[i] = result
}(link)
```

**Problem:** Errors are invisible in logs.

✅ **CORRECT:**
```go
type Result struct {
    URL    string
    Status int
    Err    error  // Store error
}

go func(i int, link string) {
    resp, err := makeRequest(link)
    results[i] = Result{
        URL: link,
        Err: err,  // Preserve error
    }
}(i, link)

// Later: inspect errors
for _, result := range results {
    if result.Err != nil {
        log.Printf("Error checking %s: %v", result.URL, result.Err)
    }
}
```

---

### Anti-Pattern 5: Deadlocking Unbuffered Channel

❌ **WRONG:**
```go
ch := make(chan int)
ch <- 42              // Blocks forever - no receiver!
value := <-ch
fmt.Println(value)
```

**Problem:** Sender blocks, no receiver ever arrives. Deadlock!

✅ **CORRECT:**
```go
// Option 1: Use goroutine receiver
ch := make(chan int)
go func() {
    value := <-ch
    fmt.Println(value)
}()
ch <- 42

// Option 2: Use buffered channel
ch := make(chan int, 1)
ch <- 42
value := <-ch
fmt.Println(value)
```

---

### Anti-Pattern 6: Holding Lock During I/O

❌ **WRONG:**
```go
var mu sync.Mutex
mu.Lock()
response := http.Get(url)  // I/O while holding lock!
processResponse(response)  // More work with lock held
mu.Unlock()
// Other goroutines blocked the entire time
```

**Problem:** Locks held longer than needed; poor concurrency.

✅ **CORRECT:**
```go
var mu sync.Mutex
response := http.Get(url)  // I/O without lock

mu.Lock()
processResponse(response)  // Only protect shared data
mu.Unlock()
```

---

### Anti-Pattern 7: Not Closing Channels Properly

❌ **WRONG:**
```go
ch := make(chan int)
go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    // Forgot to close!
}()

for value := range ch {  // Range never ends
    fmt.Println(value)   // Goroutine panics
}
```

**Problem:** Range infinitely waits for channel data.

✅ **CORRECT:**
```go
ch := make(chan int)
go func() {
    defer close(ch)  // Always close
    for i := 0; i < 5; i++ {
        ch <- i
    }
}()

for value := range ch {  // Properly terminates
    fmt.Println(value)
}
```

---

## Your Project: Pattern Analysis

Your web scraper uses the **Fan-Out/Fan-In** pattern with a **Semaphore**:

```go
// ✅ Fan-out: Launch maxConcurrent goroutines
for i, link := range links {
    go func(i int, link string) {
        // ✅ Semaphore: Limit concurrent requests
        sem <- struct{}{}
        defer func() { <-sem }()
        
        // ✅ Fan-in: Collect results
        results[i] = check(client, link)
    }(i, link)
}

wg.Wait()  // ✅ Proper synchronization
return results
```

**Why this is good:**
- ✅ Controlled concurrency (won't overwhelm server)
- ✅ Preserves result order (array indexing)
- ✅ Proper error handling (stored in Result)
- ✅ No resource leaks (proper cleanup)
- ✅ Predictable performance

---

## Debugging Concurrency Issues

### Run with Race Detector
```bash
go run -race .
```

### Use Verbose Logging
```go
log.Printf("Goroutine %d: Starting", id)
log.Printf("Goroutine %d: Acquired semaphore", id)
log.Printf("Goroutine %d: Completed", id)
```

### Add Timeouts
```go
client := &http.Client{
    Timeout: 10 * time.Second,  // Detect hung requests
}
```

### Monitor Goroutines
```go
runtime.NumGoroutine()  // Current goroutine count
```

---

## Summary Checklist

Before shipping concurrent code:

- [ ] Running with `-race` shows no issues?
- [ ] WaitGroups used for synchronization?
- [ ] Errors are propagated/logged?
- [ ] Resources are cleaned up (defer)?
- [ ] No unbounded goroutine creation?
- [ ] Timeouts set on blocking operations?
- [ ] Parameters passed to goroutines (not closure)?
- [ ] Channels properly closed?
- [ ] No mutex deadlocks?
- [ ] Stress tested with realistic load?
