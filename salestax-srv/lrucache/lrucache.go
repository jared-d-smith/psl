// Package lrucache implements a LRU cache server for string keys and double
// values (associated tax value at address).
//
// It is a simple cache server with an LRU (least recently used) eviction policy.
// It utilizes unordered map (i.e. hash table) and list to provide O(1) insertion
// and lookup.
package lrucache

import (
	"container/list"
	"errors"
	"math"
	"sync"
)

// LRUCache is a concurrent/thread safe implementation of a LRU Cache server.
type LRUCache struct {
	size  int
	list  *list.List
	cache map[interface{}]*list.Element
	mutex sync.RWMutex
}

// CacheItem hold the key/value pairs in the LRUCache. Intentionally stayed
// away from interface{} to avoid generics overhead since it does not appear
// to be valuable in this case.
type CacheItem struct {
	key   string
	value float64
}

// LoaderFunc is a function that matches the signiture of sales_tax_lookup.
type LoaderFunc func(string) (float64, error)

// New returns a pointer to an initialized LRUCache structure.
func New(sz int) *LRUCache {
	if sz <= 0 {
		panic("LRUCache size too small (<=0)")
	}
	c := &LRUCache{
		size:  sz,
		list:  list.New(),
		cache: make(map[interface{}]*list.Element, sz+1),
	}
	return c
}

// FastRateLookup implements the requested speed up utlizing the underlying
// LRUCache. It is expected that the user of this function will provide
// sales_tax_lookup routine as the second parameter to the function (fptr). This
// enables automatic slow lookup with caching in the event that a cache miss occurs.
//
// Subtle difference.  Get/Set return *CacheItem / FastRateLookup returns value type (float64)
func (c *LRUCache) FastRateLookup(key string, loader LoaderFunc) (float64, error) {
	taxRate := math.NaN()

	// test to see if key exists in the cache
	if val, err := c.Get(key); err == nil {
		taxRate = val.value
	} else {
		// cache miss but a loader function has been provided
		if loader != nil {
			// slow lookup using user provided routine
			taxRate, err := loader(key)
			if err != nil {
				return math.NaN(), errors.New("Using provided data acquistion routine")
			}
			// insert value retreived from user provided routine into cache
			c.Insert(key, taxRate)
			if err != nil {
				return math.NaN(), errors.New("Value insertion into cache failed")
			}
		} else {
			// Cache miss with no user provided data loader, return error
			return math.NaN(), err
		}
	}

	return taxRate, nil
}

// Get tests to see if a key exists in the cache. If it does not, an error
// is returned. If the key is found, error is set to nil and a pointer to the CacheItem
// is returned.
func (c *LRUCache) Get(key string) (*CacheItem, error) {
	c.mutex.RLock()
	elem, exists := c.cache[key]
	c.mutex.RUnlock()

	if exists {
		item := elem.Value.(*CacheItem)
		c.mutex.Lock()
		defer c.mutex.Unlock()
		c.list.MoveToFront(elem)
		return item, nil
	}
	return nil, errors.New("Key not found")
}

// Insert inserts a key value pair into the LRUCache. It returns an error
// if necessary.
func (c *LRUCache) Insert(key string, value float64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// test to see if elem exists in cache
	if elem, exists := c.cache[key]; exists {
		c.list.MoveToFront(elem)
		item := elem.Value.(*CacheItem)
		item.value = value
	} else {

		// test if cache is full
		if c.list.Len() >= c.size {
			c.prune(1)
		}
		ci := &CacheItem{
			key:   key,
			value: value,
		}
		c.cache[key] = c.list.PushFront(ci)
	}
	return nil
}

func (c *LRUCache) prune(n int) error {
	for i := 0; i < n; i++ {
		elem := c.list.Back()
		if elem == nil {
			return nil
		}
		c.list.Remove(elem)
		item := elem.Value.(*CacheItem)
		delete(c.cache, item.key)
	}
	return nil
}
