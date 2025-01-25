package pokecache

import (
	"time"
	"sync"
)

type Cache struct {
	entry map[string]cacheEntry
	mutex sync.Mutex
}

type cacheEntry struct {
	createdAt time.Time
	val []byte
}


func New(interval time.Duration) *Cache {
	cache := Cache{
		entry: make(map[string]cacheEntry),
	}

	cache.reapLoop(interval)

	return &cache
}

func NewCache(interval time.Duration) *Cache {
	cache := Cache{
		entry: make(map[string]cacheEntry),
	}

	cache.reapLoop(interval)

	return &cache
}

func (cache *Cache) Add(key string, val []byte) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.entry[key] = cacheEntry{
		createdAt: time.Now(),
		val:	   val,
	}
}

func (cache *Cache) Get(key string) ([]byte, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	val, found := cache.entry[key] 
	if !found {
		return nil, false
	}

	return val.val, true
}

func (cache *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()

		for {
			<-ticker.C
			cache.mutex.Lock()
			for key, entry := range cache.entry {
				if time.Since(entry.createdAt) > interval {
					delete(cache.entry, key)
				}
			}
			cache.mutex.Unlock()
		}
	}()
}
