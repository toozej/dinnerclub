// Package services provides caching functionality for user data
// to improve performance during login and subsequent requests.
package services

import (
	"sync"
	"time"

	"github.com/toozej/dinnerclub/internal/models"
)

// UserCacheEntry represents cached user data with expiration
type UserCacheEntry struct {
	Entries   []models.Entry
	CachedAt  time.Time
	ExpiresAt time.Time
}

// UserCacheService provides in-memory caching for user entries
type UserCacheService struct {
	cache      map[string]*UserCacheEntry
	mutex      sync.RWMutex
	expiration time.Duration
}

// NewUserCacheService creates a new user cache service with specified expiration
func NewUserCacheService(expiration time.Duration) *UserCacheService {
	service := &UserCacheService{
		cache:      make(map[string]*UserCacheEntry),
		expiration: expiration,
	}

	// Start cleanup goroutine
	go service.cleanupExpired()

	return service
}

// SetUserEntries caches user entries with expiration
func (s *UserCacheService) SetUserEntries(username string, entries []models.Entry) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	s.cache[username] = &UserCacheEntry{
		Entries:   entries,
		CachedAt:  now,
		ExpiresAt: now.Add(s.expiration),
	}
}

// GetUserEntries retrieves cached user entries if not expired
func (s *UserCacheService) GetUserEntries(username string) ([]models.Entry, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	entry, exists := s.cache[username]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		// Entry expired, remove it
		delete(s.cache, username)
		return nil, false
	}

	return entry.Entries, true
}

// InvalidateUser removes cached entries for a specific user
func (s *UserCacheService) InvalidateUser(username string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.cache, username)
}

// cleanupExpired periodically removes expired cache entries
func (s *UserCacheService) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		s.mutex.Lock()
		now := time.Now()
		for username, entry := range s.cache {
			if now.After(entry.ExpiresAt) {
				delete(s.cache, username)
			}
		}
		s.mutex.Unlock()
	}
}

// Global cache service instance
var UserCache *UserCacheService

// InitUserCache initializes the global user cache service
func InitUserCache() {
	UserCache = NewUserCacheService(15 * time.Minute) // Cache for 15 minutes
}
