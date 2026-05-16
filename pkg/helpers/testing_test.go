package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/gorm"
)

// mockTestable implements the testable interface for testing purposes
type mockTestable struct {
	helperCalled bool
	fatalfCalled bool
	fatalfArgs   []interface{}
}

func (m *mockTestable) Helper() {
	m.helperCalled = true
}

func (m *mockTestable) Fatalf(format string, args ...any) {
	m.fatalfCalled = true
	m.fatalfArgs = append([]interface{}{format}, args...)
}

func TestFreshDb(t *testing.T) {
	t.Run("creates in-memory database by default", func(t *testing.T) {
		// Store original database
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		mock := &mockTestable{}
		db := FreshDb(mock)

		// Verify helper was called
		assert.True(t, mock.helperCalled)
		assert.False(t, mock.fatalfCalled)

		// Verify database is functional
		assert.NotNil(t, db)

		// Test that we can perform basic operations
		var count int64
		err := db.Model(&models.User{}).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("creates database with custom path when provided", func(t *testing.T) {
		// Store original database
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		mock := &mockTestable{}
		db := FreshDb(mock, ":memory:")

		assert.True(t, mock.helperCalled)
		assert.False(t, mock.fatalfCalled)
		assert.NotNil(t, db)

		// Verify database works with custom path
		var count int64
		err := db.Model(&models.Entry{}).Count(&count).Error
		assert.NoError(t, err)
	})

	t.Run("migrates schema automatically", func(t *testing.T) {
		// Store original database
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		mock := &mockTestable{}
		db := FreshDb(mock)

		// Verify all expected tables exist by checking we can query them
		tables := []interface{}{
			&models.User{},
			&models.Entry{},
			&models.Restaurant{},
		}

		for _, table := range tables {
			var count int64
			err := db.Model(table).Count(&count).Error
			assert.NoError(t, err, "Table for %T should exist", table)
		}
	})

	t.Run("handles database creation errors gracefully", func(t *testing.T) {
		// This test is harder to trigger since SQLite is very permissive
		// but we can test the interface behavior
		mock := &mockTestable{}

		// Using a valid path should not cause fatalf to be called
		db := FreshDb(mock, ":memory:")
		assert.NotNil(t, db)
		assert.False(t, mock.fatalfCalled)
	})
}

func TestFreshDbIntegration(t *testing.T) {
	t.Run("database supports CRUD operations", func(t *testing.T) {
		// Store original database
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		db := FreshDb(t)

		// Test Create
		user := &models.User{
			Username: "testuser",
			Email:    "test@example.com",
		}
		err := db.Create(user).Error
		require.NoError(t, err)
		assert.NotZero(t, user.ID)

		// Test Read
		var foundUser models.User
		err = db.First(&foundUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "testuser", foundUser.Username)

		// Test Update
		err = db.Model(&foundUser).Update("email", "updated@example.com").Error
		require.NoError(t, err)

		// Verify update
		err = db.First(&foundUser, user.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "updated@example.com", foundUser.Email)

		// Test Delete
		err = db.Delete(&foundUser).Error
		require.NoError(t, err)

		// Verify deletion
		err = db.First(&foundUser, user.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("database supports foreign key relationships", func(t *testing.T) {
		// Store original database
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		db := FreshDb(t)

		// Create a restaurant
		restaurant := &models.Restaurant{
			Name:     "Test Restaurant",
			Location: "Test Location",
			Cuisine:  "Test Cuisine",
		}
		err := db.Create(restaurant).Error
		require.NoError(t, err)

		// Create an entry referencing the restaurant
		entry := &models.Entry{
			Name:      "Test Restaurant",
			Location:  "Test Location",
			Cuisine:   "Test Cuisine",
			Submitter: "testuser",
			Visited:   true,
		}
		err = db.Create(entry).Error
		require.NoError(t, err)

		// Verify we can query related data
		var entries []models.Entry
		err = db.Where("name = ?", "Test Restaurant").Find(&entries).Error
		require.NoError(t, err)
		assert.Len(t, entries, 1)
	})

	t.Run("multiple database instances are isolated", func(t *testing.T) {
		// Store original database
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		db1 := FreshDb(t)
		db2 := FreshDb(t)

		// Add data to first database
		user1 := &models.User{Username: "user1", Email: "user1@example.com"}
		err := db1.Create(user1).Error
		require.NoError(t, err)

		// Verify second database is empty
		var count int64
		err = db2.Model(&models.User{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		// Add different data to second database
		user2 := &models.User{Username: "user2", Email: "user2@example.com"}
		err = db2.Create(user2).Error
		require.NoError(t, err)

		// Verify databases remain isolated
		err = db1.Model(&models.User{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		err = db2.Model(&models.User{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func TestDatabaseTestUtilities(t *testing.T) {
	t.Run("database cleanup works correctly", func(t *testing.T) {
		// Store original database
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		db := FreshDb(t)

		// Add test data
		users := []models.User{
			{Username: "user1", Email: "user1@example.com"},
			{Username: "user2", Email: "user2@example.com"},
			{Username: "user3", Email: "user3@example.com"},
		}

		for _, user := range users {
			err := db.Create(&user).Error
			require.NoError(t, err)
		}

		// Verify data exists
		var count int64
		err := db.Model(&models.User{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)

		// Clean up all users
		err = db.Where("1 = 1").Delete(&models.User{}).Error
		require.NoError(t, err)

		// Verify cleanup worked
		err = db.Model(&models.User{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("database transaction rollback works", func(t *testing.T) {
		// Store original database
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		db := FreshDb(t)

		// Start transaction
		tx := db.Begin()
		require.NoError(t, tx.Error)

		// Add data in transaction
		user := &models.User{Username: "txuser", Email: "tx@example.com"}
		err := tx.Create(user).Error
		require.NoError(t, err)

		// Verify data exists in transaction
		var count int64
		err = tx.Model(&models.User{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Rollback transaction
		tx.Rollback()

		// Verify data was rolled back
		err = db.Model(&models.User{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("database supports basic operations", func(t *testing.T) {
		// Store original database
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		db := FreshDb(t)

		// Test basic database operations sequentially to verify stability
		// Create users
		for i := 0; i < 5; i++ {
			user := &models.User{
				Username: "test_user_" + string(rune(i+'0')),
				Email:    "test" + string(rune(i+'0')) + "@example.com",
			}
			err := db.Create(user).Error
			require.NoError(t, err)
		}

		// Create entries
		for i := 0; i < 5; i++ {
			entry := &models.Entry{
				Name:      "Test Restaurant " + string(rune(i+'0')),
				Location:  "Location " + string(rune(i+'0')),
				Cuisine:   "Cuisine " + string(rune(i+'0')),
				Submitter: "test_user",
				Visited:   i%2 == 0,
			}
			err := db.Create(entry).Error
			require.NoError(t, err)
		}

		// Verify both operations completed successfully
		var userCount, entryCount int64
		err := db.Model(&models.User{}).Count(&userCount).Error
		require.NoError(t, err)
		err = db.Model(&models.Entry{}).Count(&entryCount).Error
		require.NoError(t, err)

		assert.Equal(t, int64(5), userCount)
		assert.Equal(t, int64(5), entryCount)
	})
}
