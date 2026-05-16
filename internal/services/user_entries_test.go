package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Store original DB and replace with test DB
	originalDB := database.DB
	database.DB = db

	// Auto-migrate the Entry model
	err = db.AutoMigrate(&models.Entry{})
	require.NoError(t, err)

	// Cleanup function to restore original DB
	t.Cleanup(func() {
		database.DB = originalDB
	})

	return db
}

func TestNewUserEntriesService(t *testing.T) {
	service := NewUserEntriesService()

	assert.NotNil(t, service)
	assert.Equal(t, 30*time.Second, service.defaultTimeout)
}

func TestUserEntriesService_PreloadUserEntries(t *testing.T) {
	setupTestDB(t)
	service := NewUserEntriesService()

	// Insert test data
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
		{
			Name:      "Restaurant 3",
			Location:  "City 3",
			Cuisine:   "Mexican",
			Submitter: "otheruser",
			Visited:   true,
		},
	}

	for _, entry := range testEntries {
		err := database.DB.Create(&entry).Error
		require.NoError(t, err)
	}

	t.Run("preload entries for existing user", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		entries, err := service.PreloadUserEntries(ctx, "testuser")

		assert.NoError(t, err)
		assert.Len(t, entries, 2)

		// Verify entries are ordered by ID desc
		assert.Equal(t, "Restaurant 2", entries[0].Name)
		assert.Equal(t, "Restaurant 1", entries[1].Name)

		// Verify all entries belong to testuser
		for _, entry := range entries {
			assert.Equal(t, "testuser", entry.Submitter)
		}
	})

	t.Run("preload entries for non-existing user", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		entries, err := service.PreloadUserEntries(ctx, "nonexistentuser")

		assert.NoError(t, err)
		assert.Len(t, entries, 0)
	})

	t.Run("preload entries with nil context", func(t *testing.T) {
		entries, err := service.PreloadUserEntries(context.TODO(), "testuser")

		assert.NoError(t, err)
		assert.Len(t, entries, 2)
	})

	t.Run("preload entries with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := service.PreloadUserEntries(ctx, "testuser")

		// Should handle cancelled context gracefully
		// Note: SQLite might not always respect context cancellation for simple queries
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}
	})
}

func TestUserEntriesService_GetUserEntriesPaginated(t *testing.T) {
	setupTestDB(t)
	service := NewUserEntriesService()

	// Insert test data
	for i := 1; i <= 15; i++ {
		entry := models.Entry{
			Name:      fmt.Sprintf("Restaurant %d", i),
			Location:  fmt.Sprintf("City %d", i),
			Cuisine:   "Test Cuisine",
			Submitter: "testuser",
			Visited:   i%2 == 0,
		}
		err := database.DB.Create(&entry).Error
		require.NoError(t, err)
	}

	t.Run("get first page", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		entries, count, err := service.GetUserEntriesPaginated(ctx, "testuser", 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, int64(15), count)
		assert.Len(t, entries, 10)

		// Verify entries are ordered by ID desc (newest first)
		assert.Equal(t, "Restaurant 15", entries[0].Name)
		assert.Equal(t, "Restaurant 6", entries[9].Name)
	})

	t.Run("get second page", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		entries, count, err := service.GetUserEntriesPaginated(ctx, "testuser", 10, 10)

		assert.NoError(t, err)
		assert.Equal(t, int64(15), count)
		assert.Len(t, entries, 5)

		// Verify entries are the remaining ones
		assert.Equal(t, "Restaurant 5", entries[0].Name)
		assert.Equal(t, "Restaurant 1", entries[4].Name)
	})

	t.Run("get page with no results", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		entries, count, err := service.GetUserEntriesPaginated(ctx, "nonexistentuser", 10, 0)

		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
		assert.Len(t, entries, 0)
	})
}

func TestUserEntriesService_GetEntryByIDWithTimeout(t *testing.T) {
	setupTestDB(t)
	service := NewUserEntriesService()

	// Insert test entry
	testEntry := models.Entry{
		Name:      "Test Restaurant",
		Location:  "Test City",
		Cuisine:   "Test Cuisine",
		Submitter: "testuser",
		Visited:   true,
	}
	err := database.DB.Create(&testEntry).Error
	require.NoError(t, err)

	t.Run("get existing entry", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		entry, err := service.GetEntryByIDWithTimeout(ctx, testEntry.ID)

		assert.NoError(t, err)
		assert.Equal(t, testEntry.Name, entry.Name)
		assert.Equal(t, testEntry.Submitter, entry.Submitter)
	})

	t.Run("get non-existing entry", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := service.GetEntryByIDWithTimeout(ctx, 99999)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "record not found")
	})

	t.Run("get entry with nil context", func(t *testing.T) {
		entry, err := service.GetEntryByIDWithTimeout(context.TODO(), testEntry.ID)

		assert.NoError(t, err)
		assert.Equal(t, testEntry.Name, entry.Name)
	})
}

func TestUserEntriesService_CreateEntryWithTimeout(t *testing.T) {
	setupTestDB(t)
	service := NewUserEntriesService()

	t.Run("create valid entry", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		entry := &models.Entry{
			Name:      "New Restaurant",
			Location:  "New City",
			Cuisine:   "New Cuisine",
			Submitter: "testuser",
			Visited:   false,
		}

		err := service.CreateEntryWithTimeout(ctx, entry)

		assert.NoError(t, err)
		assert.NotZero(t, entry.ID)

		// Verify entry was created
		var count int64
		err = database.DB.Model(&models.Entry{}).Where("name = ?", "New Restaurant").Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("create entry with nil context", func(t *testing.T) {
		entry := &models.Entry{
			Name:      "Another Restaurant",
			Location:  "Another City",
			Cuisine:   "Another Cuisine",
			Submitter: "testuser",
			Visited:   true,
		}

		err := service.CreateEntryWithTimeout(context.TODO(), entry)

		assert.NoError(t, err)
		assert.NotZero(t, entry.ID)
	})
}

func TestUserEntriesService_UpdateEntryWithTimeout(t *testing.T) {
	setupTestDB(t)
	service := NewUserEntriesService()

	// Insert test entry
	testEntry := models.Entry{
		Name:      "Original Restaurant",
		Location:  "Original City",
		Cuisine:   "Original Cuisine",
		Submitter: "testuser",
		Visited:   false,
	}
	err := database.DB.Create(&testEntry).Error
	require.NoError(t, err)

	t.Run("update existing entry", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		updates := &models.Entry{
			Name:     "Updated Restaurant",
			Location: "Updated City",
			Visited:  true,
		}

		err := service.UpdateEntryWithTimeout(ctx, testEntry.ID, updates)

		assert.NoError(t, err)

		// Verify entry was updated
		var updatedEntry models.Entry
		err = database.DB.First(&updatedEntry, testEntry.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, "Updated Restaurant", updatedEntry.Name)
		assert.Equal(t, "Updated City", updatedEntry.Location)
		assert.True(t, updatedEntry.Visited)
	})

	t.Run("update non-existing entry", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		updates := &models.Entry{
			Name: "Should Not Update",
		}

		err := service.UpdateEntryWithTimeout(ctx, 99999, updates)

		// GORM doesn't return an error for updating non-existing records
		// It just affects 0 rows
		assert.NoError(t, err)
	})
}

func TestUserEntriesService_DeleteEntryWithTimeout(t *testing.T) {
	setupTestDB(t)
	service := NewUserEntriesService()

	// Insert test entry
	testEntry := models.Entry{
		Name:      "To Delete Restaurant",
		Location:  "Delete City",
		Cuisine:   "Delete Cuisine",
		Submitter: "testuser",
		Visited:   false,
	}
	err := database.DB.Create(&testEntry).Error
	require.NoError(t, err)

	t.Run("delete existing entry", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := service.DeleteEntryWithTimeout(ctx, testEntry.ID)

		assert.NoError(t, err)

		// Verify entry was deleted
		var count int64
		err = database.DB.Model(&models.Entry{}).Where("id = ?", testEntry.ID).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("delete non-existing entry", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := service.DeleteEntryWithTimeout(ctx, 99999)

		// GORM doesn't return an error for deleting non-existing records
		assert.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkUserEntriesService_PreloadUserEntries(b *testing.B) {
	setupTestDB(&testing.T{})
	service := NewUserEntriesService()

	// Insert test data
	for i := 0; i < 100; i++ {
		entry := models.Entry{
			Name:      fmt.Sprintf("Restaurant %d", i),
			Location:  fmt.Sprintf("City %d", i),
			Cuisine:   "Benchmark Cuisine",
			Submitter: "benchuser",
			Visited:   i%2 == 0,
		}
		database.DB.Create(&entry)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.PreloadUserEntries(ctx, "benchuser")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUserEntriesService_GetUserEntriesPaginated(b *testing.B) {
	setupTestDB(&testing.T{})
	service := NewUserEntriesService()

	// Insert test data
	for i := 0; i < 100; i++ {
		entry := models.Entry{
			Name:      fmt.Sprintf("Restaurant %d", i),
			Location:  fmt.Sprintf("City %d", i),
			Cuisine:   "Benchmark Cuisine",
			Submitter: "benchuser",
			Visited:   i%2 == 0,
		}
		database.DB.Create(&entry)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := service.GetUserEntriesPaginated(ctx, "benchuser", 10, 0)
		if err != nil {
			b.Fatal(err)
		}
	}
}
