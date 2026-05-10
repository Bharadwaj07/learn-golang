package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]string),
	}
}

func (c *Cache) Set(key, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = val
}

func (c *Cache) Get(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key]
}

func main() {
	rand.Seed(time.Now().UnixNano())
	cache := NewCache()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", rand.Intn(10))
			if rand.Intn(2) == 0 {
				cache.Set(key, fmt.Sprintf("val%d", id))
			} else {
				_ = cache.Get(key)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("All done")
}
