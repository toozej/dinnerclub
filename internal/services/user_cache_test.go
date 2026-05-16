package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/toozej/dinnerclub/internal/models"
)

func TestNewUserCacheService(t *testing.T) {
	expiration := 10 * time.Minute
	service := NewUserCacheService(expiration)

	assert.NotNil(t, service)
	assert.Equal(t, expiration, service.expiration)
	assert.NotNil(t, service.cache)
}

func TestUserCacheService_SetAndGetUserEntries(t *testing.T) {
	service := NewUserCacheService(10 * time.Minute)

	testEntries := []models.Entry{
		{
			Name:      "Restaurant 1",
			Location:  "City 1",
			Cuisine:   "Italian",
			Submitter: "testuser",
			Visited:   true,
		},
		{
			Name:      "Restaurant 2",
			Location:  "City 2",
			Cuisine:   "Japanese",
			Submitter: "testuser",
			Visited:   false,
		},
	}

	t.Run("set and get entries", func(t *testing.T) {
		service.SetUserEntries("testuser", testEntries)

		entries, found := service.GetUserEntries("testuser")

		assert.True(t, found)
		assert.Len(t, entries, 2)
		assert.Equal(t, "Restaurant 1", entries[0].Name)
		assert.Equal(t, "Restaurant 2", entries[1].Name)
	})

	t.Run("get non-existing user entries", func(t *testing.T) {
		entries, found := service.GetUserEntries("nonexistentuser")

		assert.False(t, found)
		assert.Nil(t, entries)
	})

	t.Run("overwrite existing entries", func(t *testing.T) {
		// Set initial entries
		service.SetUserEntries("testuser2", testEntries)

		// Overwrite with new entries
		newEntries := []models.Entry{
			{
				Name:      "New Restaurant",
				Location:  "New City",
				Cuisine:   "Mexican",
				Submitter: "testuser2",
				Visited:   true,
			},
		}
		service.SetUserEntries("testuser2", newEntries)

		entries, found := service.GetUserEntries("testuser2")

		assert.True(t, found)
		assert.Len(t, entries, 1)
		assert.Equal(t, "New Restaurant", entries[0].Name)
	})
}

func TestUserCacheService_Expiration(t *testing.T) {
	// Use very short expiration for testing
	service := NewUserCacheService(100 * time.Millisecond)

	testEntries := []models.Entry{
		{
			Name:      "Expiring Restaurant",
			Location:  "Expiring City",
			Cuisine:   "Expiring Cuisine",
			Submitter: "testuser",
			Visited:   true,
		},
	}

	t.Run("entries expire after timeout", func(t *testing.T) {
		service.SetUserEntries("expireuser", testEntries)

		// Should be available immediately
		entries, found := service.GetUserEntries("expireuser")
		assert.True(t, found)
		assert.Len(t, entries, 1)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Should be expired now
		entries, found = service.GetUserEntries("expireuser")
		assert.False(t, found)
		assert.Nil(t, entries)
	})

	t.Run("cache entry is removed after expiration", func(t *testing.T) {
		service.SetUserEntries("expireuser2", testEntries)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Try to get expired entry (should remove it from cache)
		_, found := service.GetUserEntries("expireuser2")
		assert.False(t, found)

		// Verify it's actually removed from the internal cache
		service.mutex.RLock()
		_, exists := service.cache["expireuser2"]
		service.mutex.RUnlock()
		assert.False(t, exists)
	})
}

func TestUserCacheService_InvalidateUser(t *testing.T) {
	service := NewUserCacheService(10 * time.Minute)

	testEntries := []models.Entry{
		{
			Name:      "Restaurant to Invalidate",
			Location:  "Invalidate City",
			Cuisine:   "Invalidate Cuisine",
			Submitter: "invalidateuser",
			Visited:   true,
		},
	}

	t.Run("invalidate existing user", func(t *testing.T) {
		service.SetUserEntries("invalidateuser", testEntries)

		// Verify entries are cached
		entries, found := service.GetUserEntries("invalidateuser")
		assert.True(t, found)
		assert.Len(t, entries, 1)

		// Invalidate the user
		service.InvalidateUser("invalidateuser")

		// Verify entries are no longer cached
		entries, found = service.GetUserEntries("invalidateuser")
		assert.False(t, found)
		assert.Nil(t, entries)
	})

	t.Run("invalidate non-existing user", func(t *testing.T) {
		// Should not panic or cause issues
		service.InvalidateUser("nonexistentuser")

		// Verify other cached entries are not affected
		service.SetUserEntries("otheruser", testEntries)
		entries, found := service.GetUserEntries("otheruser")
		assert.True(t, found)
		assert.Len(t, entries, 1)

		service.InvalidateUser("nonexistentuser")

		entries, found = service.GetUserEntries("otheruser")
		assert.True(t, found)
		assert.Len(t, entries, 1)
	})
}

func TestUserCacheService_ConcurrentAccess(t *testing.T) {
	service := NewUserCacheService(10 * time.Minute)

	testEntries := []models.Entry{
		{
			Name:      "Concurrent Restaurant",
			Location:  "Concurrent City",
			Cuisine:   "Concurrent Cuisine",
			Submitter: "concurrentuser",
			Visited:   true,
		},
	}

	t.Run("concurrent set and get operations", func(t *testing.T) {
		done := make(chan bool, 10)

		// Start multiple goroutines setting entries
		for i := 0; i < 5; i++ {
			go func(id int) {
				username := fmt.Sprintf("user%d", id)
				service.SetUserEntries(username, testEntries)
				done <- true
			}(i)
		}

		// Start multiple goroutines getting entries
		for i := 0; i < 5; i++ {
			go func(id int) {
				username := fmt.Sprintf("user%d", id)
				service.GetUserEntries(username)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all entries were set correctly
		for i := 0; i < 5; i++ {
			username := fmt.Sprintf("user%d", i)
			entries, found := service.GetUserEntries(username)
			assert.True(t, found)
			assert.Len(t, entries, 1)
		}
	})
}

func TestUserCacheService_CacheEntry(t *testing.T) {
	service := NewUserCacheService(10 * time.Minute)

	testEntries := []models.Entry{
		{
			Name:      "Cache Entry Test",
			Location:  "Cache City",
			Cuisine:   "Cache Cuisine",
			Submitter: "cacheuser",
			Visited:   true,
		},
	}

	t.Run("cache entry has correct timestamps", func(t *testing.T) {
		beforeSet := time.Now()
		service.SetUserEntries("timestampuser", testEntries)
		afterSet := time.Now()

		service.mutex.RLock()
		entry, exists := service.cache["timestampuser"]
		service.mutex.RUnlock()

		assert.True(t, exists)
		assert.True(t, entry.CachedAt.After(beforeSet) || entry.CachedAt.Equal(beforeSet))
		assert.True(t, entry.CachedAt.Before(afterSet) || entry.CachedAt.Equal(afterSet))
		assert.True(t, entry.ExpiresAt.After(entry.CachedAt))
		assert.Equal(t, service.expiration, entry.ExpiresAt.Sub(entry.CachedAt))
	})
}

func TestInitUserCache(t *testing.T) {
	// Store original UserCache
	originalUserCache := UserCache
	defer func() {
		UserCache = originalUserCache
	}()

	t.Run("initialize global user cache", func(t *testing.T) {
		InitUserCache()

		assert.NotNil(t, UserCache)
		assert.Equal(t, 15*time.Minute, UserCache.expiration)
	})
}

// Benchmark tests
func BenchmarkUserCacheService_SetUserEntries(b *testing.B) {
	service := NewUserCacheService(10 * time.Minute)

	testEntries := []models.Entry{
		{
			Name:      "Benchmark Restaurant",
			Location:  "Benchmark City",
			Cuisine:   "Benchmark Cuisine",
			Submitter: "benchuser",
			Visited:   true,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		username := fmt.Sprintf("user%d", i%100) // Cycle through 100 users
		service.SetUserEntries(username, testEntries)
	}
}

func BenchmarkUserCacheService_GetUserEntries(b *testing.B) {
	service := NewUserCacheService(10 * time.Minute)

	testEntries := []models.Entry{
		{
			Name:      "Benchmark Restaurant",
			Location:  "Benchmark City",
			Cuisine:   "Benchmark Cuisine",
			Submitter: "benchuser",
			Visited:   true,
		},
	}

	// Pre-populate cache
	for i := 0; i < 100; i++ {
		username := fmt.Sprintf("user%d", i)
		service.SetUserEntries(username, testEntries)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		username := fmt.Sprintf("user%d", i%100) // Cycle through 100 users
		service.GetUserEntries(username)
	}
}

func BenchmarkUserCacheService_ConcurrentAccess(b *testing.B) {
	service := NewUserCacheService(10 * time.Minute)

	testEntries := []models.Entry{
		{
			Name:      "Concurrent Benchmark Restaurant",
			Location:  "Concurrent Benchmark City",
			Cuisine:   "Concurrent Benchmark Cuisine",
			Submitter: "concurrentbenchuser",
			Visited:   true,
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			username := fmt.Sprintf("user%d", i%100)
			if i%2 == 0 {
				service.SetUserEntries(username, testEntries)
			} else {
				service.GetUserEntries(username)
			}
			i++
		}
	})
}
