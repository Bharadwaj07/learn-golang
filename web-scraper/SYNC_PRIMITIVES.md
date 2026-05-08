# Go Synchronization Primitives

Go provides several tools to coordinate goroutines. This document covers the most important ones.

## WaitGroup (sync.WaitGroup)

Waits for a collection of goroutines to finish.

### API
```go
var wg sync.WaitGroup

wg.Add(n)        // Tell WaitGroup to wait for n goroutines
wg.Done()        // Signal that one goroutine is done
wg.Wait()        // Block until all Done() calls received
```

### Usage Pattern

```go
var wg sync.WaitGroup

// Add 3 before starting
wg.Add(3)

go func() {
    defer wg.Done()
    // Work 1
}()

go func() {
    defer wg.Done()
    // Work 2
}()

go func() {
    defer wg.Done()
    // Work 3
}()

wg.Wait()  // Waits for all 3 goroutines
fmt.Println("All done!")
```

### In Your Project
```go
var wg sync.WaitGroup

for i, link := range links {
    wg.Add(1)  // Add 1 per link
    
    go func(i int, link string) {
        defer wg.Done()  // Signal done
        results[i] = check(client, link)
    }(i, link)
}

wg.Wait()  // All links checked before returning
```

### Key Rules
1. **Add before launching goroutine**: `wg.Add(1)` then `go func()`
2. **Use `defer wg.Done()`**: Ensures it's called even on panic
3. **Never use negative Add()**: `wg.Add(-1)` only after `wg.Add(2)` for example

### Common Mistake
```go
// WRONG: Add called from goroutine
go func() {
    wg.Add(1)  // Race condition!
    defer wg.Done()
}()
```

---

## Mutex (sync.Mutex)

Mutex (mutual exclusion) protects shared data from concurrent access.

### API
```go
var mu sync.Mutex

mu.Lock()        // Acquire lock (blocks if locked)
mu.Unlock()      // Release lock

// Better: deferred unlock
mu.Lock()
defer mu.Unlock()
// Safe access to shared data
```

### Example: Counter
```go
var count int
var mu sync.Mutex

// Unsafe (race condition)
go func() { count++ }()
go func() { count++ }()

// Safe (with mutex)
go func() {
    mu.Lock()
    count++  // Protected
    mu.Unlock()
}()

go func() {
    mu.Lock()
    count++  // Protected
    mu.Unlock()
}()
```

### RWMutex (Read-Write Mutex)
Allows multiple readers OR one writer:

```go
var data string
var rwmu sync.RWMutex

// Multiple readers OK
go func() {
    rwmu.RLock()
    fmt.Println(data)  // Safe read
    rwmu.RUnlock()
}()

go func() {
    rwmu.RLock()
    fmt.Println(data)  // Another safe read
    rwmu.RUnlock()
}()

// Writer (exclusive)
go func() {
    rwmu.Lock()
    data = "new value"  // Exclusive write
    rwmu.Unlock()
}()
```

### When NOT to Use Mutex
Your web scraper doesn't need mutex because:
- Each goroutine writes to unique array index: `results[i]`
- No overlapping writes to same memory
- Array is only read after `wg.Wait()`

This avoids mutex overhead and is cleaner.

---

## Atomic Operations (sync/atomic)

Lock-free synchronization for simple values.

```go
var counter int64
atomic.AddInt64(&counter, 1)     // Increment safely
value := atomic.LoadInt64(&counter)  // Read safely
```

### When to Use
- Simple integer counters
- Boolean flags
- When mutex would be overkill

### Example
```go
var requests int64

go func() {
    for i := 0; i < 1000; i++ {
        atomic.AddInt64(&requests, 1)
    }
}()

go func() {
    time.Sleep(100 * time.Millisecond)
    count := atomic.LoadInt64(&requests)
    fmt.Println("Requests so far:", count)
}()
```

---

## Semaphore (via Buffered Channel)

Limits concurrent access to a resource.

### Implementation
```go
sem := make(chan struct{}, numWorkers)

for i := 0; i < totalTasks; i++ {
    sem <- struct{}{}        // Acquire
    defer func() { <-sem }() // Release
    
    // Critical section
    doWork()
}
```

### In Your Project
```go
sem := make(chan struct{}, maxConcurrent)

go func(i int, link string) {
    defer wg.Done()
    
    sem <- struct{}{}        // Acquire semaphore
    defer func() { <-sem }() // Release semaphore
    
    results[i] = check(client, link)
}(i, link)
```

### Real-World Use: Database Connection Pool
```go
// Limit 10 concurrent database connections
connPool := make(chan struct{}, 10)

for i := 0; i < 100; i++ {
    go func(id int) {
        connPool <- struct{}{}  // Get connection slot
        defer func() { <-connPool }()
        
        executeQuery(id)
    }(i)
}
```

---

## Once (sync.Once)

Executes code exactly once, no matter how many goroutines call it.

```go
var once sync.Once
var instance *Singleton

func getInstance() *Singleton {
    once.Do(func() {
        instance = &Singleton{}
    })
    return instance
}

// Called from multiple goroutines
go func() { getInstance() }()
go func() { getInstance() }()
go func() { getInstance() }()

// instance is created exactly once
```

### Use Cases
- Singleton pattern (create once, share many)
- Lazy initialization
- One-time setup

---

## Cond (sync.Cond)

Coordinates goroutines that need to wait for an event.

```go
var mu sync.Mutex
var cond = sync.NewCond(&mu)
var done bool

// Waiter
go func() {
    cond.L.Lock()
    for !done {
        cond.Wait()  // Release lock, wait for signal
    }
    cond.L.Unlock()
    fmt.Println("Done!")
}()

// Signaler (after some work)
time.Sleep(1 * time.Second)
cond.L.Lock()
done = true
cond.Broadcast()  // Wake all waiters
cond.L.Unlock()
```

### Variants
```go
cond.Wait()       // Wait for signal
cond.Signal()     // Wake 1 waiting goroutine
cond.Broadcast()  // Wake all waiting goroutines
```

---

## Comparison Table

| Tool | Use Case | Overhead |
|------|----------|----------|
| WaitGroup | Wait for goroutines to finish | Low |
| Mutex | Protect shared data | Medium |
| RWMutex | Readers + writer | Medium-High |
| Atomic | Simple counters | Very Low |
| Semaphore | Limit concurrency | Low |
| Once | Execute once | Very Low |
| Cond | Event signaling | Medium |
| Channel | Communication + sync | Low-Medium |

---

## Design Patterns

### Pattern: Work Pool with WaitGroup

```go
var wg sync.WaitGroup
jobs := []string{"job1", "job2", ...}

// Launch workers
for i := 0; i < numWorkers; i++ {
    wg.Add(1)
    go worker(&wg)
}

func worker(wg *sync.WaitGroup) {
    defer wg.Done()
    // Do work
}
```

### Pattern: Graceful Shutdown

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
        // Process work
        }
    }
}()

// Shutdown
close(done)
wg.Wait()
```

### Pattern: Rate Limiter

```go
limiter := make(chan struct{}, ratePerSecond)

go func() {
    tick := time.Tick(1 * time.Second)
    for range tick {
        for i := 0; i < ratePerSecond; i++ {
            select {
            case limiter <- struct{}{}:
            default:
            }
        }
    }
}()

// Use limiter before each request
<-limiter
makeRequest()
```

---

## Best Practices

1. **Prefer channels to mutexes** for goroutine communication
2. **Use WaitGroup for "fork-join" patterns** (start many, wait for all)
3. **Use Mutex only for shared data** without clear ownership
4. **Always `defer Unlock()`** to prevent deadlocks
5. **Use `sync.Once` for singletons**
6. **Run with `-race` flag** to detect synchronization bugs
7. **Avoid nested locks** (mutex inside mutex) - prone to deadlocks
8. **Keep critical sections small** - minimize lock contention

---

## Deadlock Prevention

A deadlock occurs when goroutines wait indefinitely for each other.

### Cause: Unbuffered Channel Deadlock
```go
ch := make(chan int)
ch <- 42      // Deadlock! No receiver waiting
value := <-ch  // Never reached
```

### Solution: Use Goroutine
```go
ch := make(chan int)

go func() {
    value := <-ch  // Receiver ready
}()

ch <- 42  // Can send now
```

### Cause: Mutex Deadlock
```go
mu.Lock()
doSomething()
mu.Lock()  // Deadlock! Same thread, lock still held
mu.Unlock()
```

### Solution: Different locks or redesign
```go
var mu1, mu2 sync.Mutex

mu1.Lock()
doA()
mu1.Unlock()  // Release before acquiring another

mu2.Lock()
doB()
mu2.Unlock()
```

---

## Testing for Concurrency Bugs

```bash
# Run with race detector
go run -race main.go

# Race detector enabled tests
go test -race ./...

# Coverage with race detection
go test -race -cover ./...
```

The race detector checks for:
- Concurrent reads/writes to same memory
- Concurrent writes to same memory
- Unsynchronized access to variables

Your web scraper should pass `-race` with no issues!
