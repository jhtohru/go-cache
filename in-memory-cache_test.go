package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEntry(t *testing.T) {
	t.Run("entry is expired", func(t *testing.T) {
		entry := entry[string]{
			Key:        "key",
			Value:      "value",
			Expiration: time.Now().Add(-1 * time.Hour),
		}
		assert.True(t, entry.IsExpired())
	})

	t.Run("entry is not expired", func(t *testing.T) {
		entry := entry[string]{
			Key:        "key",
			Value:      "value",
			Expiration: time.Now().Add(1 * time.Hour),
		}
		assert.False(t, entry.IsExpired())
	})
}

func TestInMemoryCache(t *testing.T) {
	t.Run("missing entry", func(t *testing.T) {
		cache := NewInMemoryCache[string](1*time.Hour, 1*time.Hour)
		defer cache.Close()

		value, err := cache.Get("key")

		assert.Empty(t, value)
		assert.Error(t, err)
	})

	t.Run("expired entry", func(t *testing.T) {
		timeToLive := 1 * time.Millisecond
		cleaningInterval := 1 * time.Hour
		cache := NewInMemoryCache[string](timeToLive, cleaningInterval)
		defer cache.Close()

		cache.Set("key", "value")

		time.Sleep(10 * time.Millisecond) // Wait entry to expire.

		value, err := cache.Get("key")

		assert.Empty(t, value)
		assert.Error(t, err)
	})

	t.Run("deleted expired entry", func(t *testing.T) {
		timeToLive := 1 * time.Millisecond
		cleaningInterval := 1 * time.Millisecond
		cache := NewInMemoryCache[string](timeToLive, cleaningInterval)
		defer cache.Close()

		cache.Set("key", "value")

		time.Sleep(10 * time.Millisecond) // Wait entry to expire and be deleted.

		value, err := cache.Get("key")

		assert.Empty(t, value)
		assert.Error(t, err)
	})

	t.Run("cache hit", func(t *testing.T) {
		cache := NewInMemoryCache[string](1*time.Hour, 1*time.Hour)
		defer cache.Close()

		cache.Set("key", "value")

		value, err := cache.Get("key")

		assert.Equal(t, "value", value)
		assert.NoError(t, err)
	})

	t.Run("updating an entry", func(t *testing.T) {
		cache := NewInMemoryCache[string](1*time.Hour, 1*time.Hour)
		defer cache.Close()

		cache.Set("key", "value")
		cache.Set("key", "updated value")

		value, err := cache.Get("key")

		assert.Equal(t, "updated value", value)
		assert.NoError(t, err)
	})

	t.Run("using a closed cache", func(t *testing.T) {
		cache := NewInMemoryCache[string](1*time.Hour, 1*time.Hour)

		cache.Close()

		assert.PanicsWithValue(t, "cache is closed", func() { cache.Set("key", "value") })
		assert.PanicsWithValue(t, "cache is closed", func() { cache.Get("key") })
		assert.PanicsWithValue(t, "cache is closed", func() { cache.Close() })
	})
}
