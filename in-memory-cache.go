package cache

import (
	"errors"
	"sync"
	"time"
)

type InMemoryCache[T any] struct {
	timeToLive     time.Duration
	cleaningTicker *time.Ticker
	done           chan struct{}
	mutex          sync.RWMutex
	isClosed       bool
	hashmap        map[string]*entry[T]
	queue          []*entry[T]
}

func NewInMemoryCache[T any](timeToLive, autoCleanInterval time.Duration) *InMemoryCache[T] {
	c := &InMemoryCache[T]{
		timeToLive:     timeToLive,
		cleaningTicker: time.NewTicker(autoCleanInterval),
		done:           make(chan struct{}),
		hashmap:        make(map[string]*entry[T]),
	}
	go c.autoClean()
	return c
}

func (c *InMemoryCache[T]) Set(key string, v T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.mustBeNotClosed()
	if _, exists := c.hashmap[key]; exists {
		// Remove entry with the same key from the queue.
		for i := range c.queue {
			if c.queue[i].Key == key {
				c.queue = append(c.queue[:i], c.queue[i+1:]...)
				break
			}
		}
	}
	e := &entry[T]{
		Key:        key,
		Value:      v,
		Expiration: time.Now().Add(c.timeToLive),
	}
	c.hashmap[key] = e
	c.queue = append(c.queue, e)
}

func (c *InMemoryCache[T]) Get(key string) (T, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	c.mustBeNotClosed()
	var v T
	e, exists := c.hashmap[key]
	if !exists || e.IsExpired() {
		return v, errors.New("cache miss")
	}
	return e.Value, nil
}

func (c *InMemoryCache[T]) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.mustBeNotClosed()
	close(c.done)
	c.cleaningTicker.Stop()
	c.isClosed = true
}

func (c *InMemoryCache[T]) mustBeNotClosed() {
	if c.isClosed {
		panic("cache is closed")
	}
}

func (c *InMemoryCache[T]) autoClean() {
	for {
		select {
		case <-c.cleaningTicker.C:
			c.deleteExpiredEntries()
		case <-c.done:
			return
		}
	}
}

func (c *InMemoryCache[T]) deleteExpiredEntries() {
	if len(c.queue) == 0 {
		return
	}
	c.mutex.Lock()
	for len(c.queue) != 0 && c.queue[0].IsExpired() {
		c.queue = c.queue[1:]
	}
	c.mutex.Unlock()
}

type entry[T any] struct {
	Key        string
	Value      T
	Expiration time.Time
}

func (e entry[T]) IsExpired() bool {
	return e.Expiration.Before(time.Now())
}
