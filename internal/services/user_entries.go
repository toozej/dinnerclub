// Package services provides business logic and data access services
// for the dinnerclub application with optimized database operations.
package services

import (
	"context"
	"fmt"
	"time"

	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
)

// UserEntriesService handles user-specific entry operations with preloading capabilities
type UserEntriesService struct {
	defaultTimeout time.Duration
}

// NewUserEntriesService creates a new instance of UserEntriesService with default timeout
func NewUserEntriesService() *UserEntriesService {
	return &UserEntriesService{
		defaultTimeout: 30 * time.Second, // Default 30 second timeout
	}
}

// PreloadUserEntries fetches all entries for a specific user with timeout and context support.
// This is designed to be called during login to preload user data.
func (s *UserEntriesService) PreloadUserEntries(ctx context.Context, username string) ([]models.Entry, error) {
	// Create a timeout context if none provided
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.defaultTimeout)
		defer cancel()
	}

	var entries []models.Entry

	// Use GORM with context for timeout support
	db := database.WithContext(ctx)
	if db == nil {
		return entries, fmt.Errorf("database not available")
	}

	err := db.Where("submitter = ?", username).
		Order("id desc").
		Find(&entries).Error

	return entries, err
}

// GetUserEntriesPaginated fetches paginated entries for a user with timeout support
func (s *UserEntriesService) GetUserEntriesPaginated(ctx context.Context, username string, limit, offset int) ([]models.Entry, int64, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.defaultTimeout)
		defer cancel()
	}

	var entries []models.Entry
	var count int64

	db := database.WithContext(ctx)

	// Get total count
	if err := db.Model(&models.Entry{}).Where("submitter = ?", username).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated entries
	err := db.Where("submitter = ?", username).
		Order("id desc").
		Limit(limit).
		Offset(offset).
		Find(&entries).Error

	return entries, count, err
}

// GetEntryByIDWithTimeout fetches a single entry by ID with timeout support
func (s *UserEntriesService) GetEntryByIDWithTimeout(ctx context.Context, entryID uint) (*models.Entry, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.defaultTimeout)
		defer cancel()
	}

	var entry models.Entry
	err := database.WithContext(ctx).
		Where("id = ?", entryID).
		First(&entry).Error

	return &entry, err
}

// CreateEntryWithTimeout creates a new entry with timeout support
func (s *UserEntriesService) CreateEntryWithTimeout(ctx context.Context, entry *models.Entry) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.defaultTimeout)
		defer cancel()
	}

	return database.WithContext(ctx).Create(entry).Error
}

// UpdateEntryWithTimeout updates an existing entry with timeout support
func (s *UserEntriesService) UpdateEntryWithTimeout(ctx context.Context, entryID uint, updates *models.Entry) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.defaultTimeout)
		defer cancel()
	}

	return database.WithContext(ctx).
		Model(&models.Entry{}).
		Where("id = ?", entryID).
		Updates(updates).Error
}

// DeleteEntryWithTimeout deletes an entry with timeout support
func (s *UserEntriesService) DeleteEntryWithTimeout(ctx context.Context, entryID uint) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.defaultTimeout)
		defer cancel()
	}

	return database.WithContext(ctx).
		Delete(&models.Entry{}, entryID).Error
}
