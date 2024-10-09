package pokeCache

import (
    "sync";
    "time";
    "fmt";
)

type Cache struct {
    Entries     map[string]cacheEntry
    mu          sync.Mutex
    interval    time.Duration 
}

type cacheEntry struct {
    createdAt   time.Time
    val         []byte
}

//Main function of this internal package for setting up a Cache 
func NewCache(interval int) *Cache {
    return &Cache{
        Entries: make(map[string]cacheEntry),
        interval: time.Duration(interval) * time.Second,
    }
}

//frucntion to add new entries to the map
func (c *Cache) Add(key string, val []byte) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    if _, ok := c.Entries[key]; ok {
        return fmt.Errorf("The key %v already exists in the Cache.", key)
    }
    entry := cacheEntry{
        createdAt:  time.Now(),
        val:        val,
    }
    c.Entries[key] = entry
    return nil 
}

//Function to retrieve a Cache entry 
func (c *Cache) Get(key string) ([]byte, bool) {
     
    c.mu.Lock()
    defer c.mu.Unlock()
    entry, ok := c.Entries[key]
    if !ok {
        return nil, false
    }
    return entry.val, true
}
 
//Function to clean up entries after a certain duration specified in the NewCache function 
func (c *Cache) reapLoop() {
    ticker := time.NewTicker(c.interval)
    for {
        <-ticker.C 
        c.mu.Lock()
        for key, entry := range c.Entries {
            if time.Since(entry.createdAt) > c.interval {
                delete(c.Entries, key)
            }
        }
    c.mu.Unlock()
    }
}
