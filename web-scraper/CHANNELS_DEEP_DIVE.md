# Channels Deep Dive

Channels are the primary way goroutines communicate in Go. This document provides detailed examples.

## Channel Basics

### Creating Channels
```go
// Unbuffered channel (synchronous)
ch := make(chan int)

// Buffered channel (asynchronous)  
ch := make(chan int, 10)  // Capacity 10

// Channel for any type
ch := make(chan interface{})

// Channel of structs
type Message struct {
    ID    int
    Text  string
}
ch := make(chan Message)
```

### Sending and Receiving

```go
// Send a value to channel
ch <- value

// Receive a value from channel
value := <-ch

// Receive and test if open
value, ok := <-ch  // ok=false if channel closed and empty

// Receive without using value
<-ch

// The arrow always points LEFT when receiving!
value := <-ch   // Correct: receives into value
<-ch := value   // Wrong: syntax error
```

---

## Unbuffered vs Buffered

### Unbuffered Channels (Synchronous)

Both sender and receiver must be ready at the same time.

```go
ch := make(chan int)  // No capacity

// Example 1: Deadlock!
ch <- 42              // Blocks here - no receiver waiting!
value := <-ch         // Never reached
```

```go
// Example 2: Works
ch := make(chan int)

go func() {
    value := <-ch     // Waiting for sender...
    fmt.Println(value)
}()

ch <- 42              // Unblocks the goroutine above
```

**Timing matters:**
- Sender blocks until receiver is ready
- Receiver blocks until sender sends
- Perfect for synchronization

### Buffered Channels (Asynchronous)

Sender can send without a ready receiver (up to buffer capacity).

```go
ch := make(chan int, 3)

ch <- 1              // OK, buffer has space
ch <- 2              // OK, buffer has space
ch <- 3              // OK, buffer has space
ch <- 4              // BLOCKS - buffer full!
```

**Benefits:**
- Decouples sender and receiver
- Sender doesn't wait for receiver
- Prevents goroutine starvation

---

## Closing Channels

You can explicitly close a channel:

```go
close(ch)

// After close:
ch <- 42                  // PANIC! Send on closed channel
value, ok := <-ch         // ok=false
```

### Rules for Closing
1. **Only senders close**: Receiver shouldn't close
2. **Multiple senders**: Use sync.Once or separate close channel
3. **Test before sending**: Check if channel is closed first

```go
// Option 1: Single sender closes
go func() {
    for i := 0; i < 5; i++ {
        ch <- i
    }
    close(ch)  // Sender closes when done
}()

// Option 2: Detect closed channel
select {
case ch <- value:
    // Sent successfully
default:
    // Channel is closed or full
}
```

---

## Channel Patterns in Your Project

### Pattern: Semaphore
Your web scraper uses a buffered channel as a semaphore:

```go
sem := make(chan struct{}, maxConcurrent)

// Acquire
sem <- struct{}{}

// Release
<-sem
```

**How it works:**
- Channel has `maxConcurrent` "slots"
- Send fills a slot (blocks if full)
- Receive empties a slot
- Limited concurrency achieved!

---

## Select Statement

`select` lets a goroutine wait on multiple channel operations:

```go
select {
case value := <-ch1:
    fmt.Println("From ch1:", value)
case ch2 <- 42:
    fmt.Println("Sent to ch2")
case <-time.After(1 * time.Second):
    fmt.Println("Timeout!")
default:
    fmt.Println("No channel ready")
}
```

### Use Cases

**1. Multiplexing Multiple Channels**
```go
ch1 := make(chan int)
ch2 := make(chan int)

go send(ch1)
go send(ch2)

for i := 0; i < 4; i++ {
    select {
    case value := <-ch1:
        fmt.Println("From ch1:", value)
    case value := <-ch2:
        fmt.Println("From ch2:", value)
    }
}
```

**2. Timeout Handling**
```go
select {
case result := <-resultChan:
    return result
case <-time.After(5 * time.Second):
    return errors.New("request timeout")
}
```

**3. Default Case (Non-blocking)**
```go
select {
case value := <-ch:
    fmt.Println("Received:", value)
default:
    fmt.Println("Channel not ready (non-blocking)")
}
```

**4. Graceful Shutdown**
```go
done := make(chan bool)

go worker()

select {
case <-done:
    fmt.Println("Worker completed")
case <-time.After(10 * time.Second):
    fmt.Println("Worker timeout")
}
```

---

## Common Channel Mistakes

### Mistake 1: Sending on Closed Channel
```go
ch := make(chan int)
close(ch)
ch <- 42  // PANIC!
```

### Mistake 2: Closing Closed Channel
```go
ch := make(chan int)
close(ch)
close(ch)  // PANIC!
```

### Mistake 3: Multiple Senders, Unclear Who Closes
```go
// BAD: Who closes the channel?
go sender1(ch)
go sender2(ch)
go receiver(ch)

close(ch)  // Which goroutine should do this?
```

**Solution: Use a done channel**
```go
done := make(chan struct{})
go sender(ch, done)
go receiver(ch, done)

<-done  // Wait for receiver
close(ch)
```

### Mistake 4: Forgetting to Receive
```go
ch := make(chan int, 1)
ch <- 42
value := <-ch
// GOOD

ch := make(chan int)
ch <- 42  // Blocks forever - no receiver!
```

---

## Advanced: Buffered Channel as Queue

Use buffered channels to implement a work queue:

```go
jobs := make(chan string, 100)
results := make(chan Result, 100)

// Start 5 workers
for i := 0; i < 5; i++ {
    go func(id int) {
        for job := range jobs {
            results <- processJob(job)
        }
    }(i)
}

// Send jobs
URLs := []string{"url1", "url2", ...}
for _, url := range URLs {
    jobs <- url
}
close(jobs)

// Collect results
for i := 0; i < len(URLs); i++ {
    result := <-results
    // process result
}
```

---

## Channel Performance Tips

1. **Use buffered channels when possible** (less blocking)
2. **Buffer size = expected concurrent senders** (common heuristic)
3. **Avoid creating thousands of unbuffered channels**
4. **Use `range` to receive until closed**:
   ```go
   for value := range ch {
       // Automatically loops until ch closed
   }
   ```
5. **Don't pass channels if simpler synchronization works like WaitGroups**

---

## Real-World Example: Pipeline

A pipeline is a series of stages connected by channels:

```go
// Stage 1: Generate numbers
func generate(nums ...int) <-chan int {
    out := make(chan int)
    go func() {
        for _, n := range nums {
            out <- n
        }
        close(out)
    }()
    return out
}

// Stage 2: Square numbers
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

// Stage 3: Print
func main() {
    nums := generate(2, 3, 4)
    squared := square(nums)
    for result := range squared {
        fmt.Println(result)  // 4, 9, 16
    }
}
```

Each stage is independent and can be reused in different pipelines!
