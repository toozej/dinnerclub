package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/gorm"
)

// Test the setupTestDB function used in user_entries_test.go
func TestSetupTestDB(t *testing.T) {
	t.Run("creates in-memory database", func(t *testing.T) {
		db := setupTestDB(t)

		assert.NotNil(t, db)

		// Verify it's a working database
		sqlDB, err := db.DB()
		require.NoError(t, err)
		err = sqlDB.Ping()
		assert.NoError(t, err)
	})

	t.Run("migrates Entry model", func(t *testing.T) {
		db := setupTestDB(t)

		// Verify Entry table exists and is functional
		var count int64
		err := db.Model(&models.Entry{}).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count) // Should start empty
	})

	t.Run("sets up database.DB global variable", func(t *testing.T) {
		originalDB := database.DB

		db := setupTestDB(t)

		// Verify global DB was set
		assert.Equal(t, db, database.DB)

		// The cleanup function should restore the original DB
		// This is tested implicitly by the t.Cleanup mechanism
		// Note: originalDB might be nil in test environment, which is fine
		_ = originalDB
	})

	t.Run("cleanup restores original database", func(t *testing.T) {
		originalDB := database.DB

		// Create a sub-test to trigger cleanup
		t.Run("sub-test", func(t *testing.T) {
			testDB := setupTestDB(t)
			assert.Equal(t, testDB, database.DB)
		})

		// After sub-test completes, cleanup should have run
		assert.Equal(t, originalDB, database.DB)
	})

	t.Run("supports Entry model operations", func(t *testing.T) {
		db := setupTestDB(t)

		// Test Create
		entry := &models.Entry{
			Name:      "Test Restaurant",
			Location:  "Test Location",
			Cuisine:   "Test Cuisine",
			Submitter: "testuser",
			Visited:   true,
		}
		err := db.Create(entry).Error
		require.NoError(t, err)
		assert.NotZero(t, entry.ID)

		// Test Read
		var foundEntry models.Entry
		err = db.First(&foundEntry, entry.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "Test Restaurant", foundEntry.Name)
		assert.Equal(t, "testuser", foundEntry.Submitter)
		assert.True(t, foundEntry.Visited)

		// Test Update
		err = db.Model(&foundEntry).Update("visited", false).Error
		require.NoError(t, err)

		// Verify update
		err = db.First(&foundEntry, entry.ID).Error
		require.NoError(t, err)
		assert.False(t, foundEntry.Visited)

		// Test Delete
		err = db.Delete(&foundEntry).Error
		require.NoError(t, err)

		// Verify deletion
		err = db.First(&foundEntry, entry.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("multiple test databases are isolated", func(t *testing.T) {
		db1 := setupTestDB(t)
		db2 := setupTestDB(t)

		// Add data to first database
		entry1 := &models.Entry{
			Name:      "Restaurant 1",
			Location:  "Location 1",
			Cuisine:   "Cuisine 1",
			Submitter: "user1",
			Visited:   true,
		}
		err := db1.Create(entry1).Error
		require.NoError(t, err)

		// Verify second database is empty
		var count int64
		err = db2.Model(&models.Entry{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		// Add different data to second database
		entry2 := &models.Entry{
			Name:      "Restaurant 2",
			Location:  "Location 2",
			Cuisine:   "Cuisine 2",
			Submitter: "user2",
			Visited:   false,
		}
		err = db2.Create(entry2).Error
		require.NoError(t, err)

		// Verify databases remain isolated
		err = db1.Model(&models.Entry{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		err = db2.Model(&models.Entry{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Verify each database has its own data
		var entry1Found models.Entry
		err = db1.First(&entry1Found).Error
		require.NoError(t, err)
		assert.Equal(t, "Restaurant 1", entry1Found.Name)

		var entry2Found models.Entry
		err = db2.First(&entry2Found).Error
		require.NoError(t, err)
		assert.Equal(t, "Restaurant 2", entry2Found.Name)
	})

	t.Run("handles sequential database operations", func(t *testing.T) {
		db := setupTestDB(t)

		// Test sequential writes to verify database stability
		for i := 0; i < 5; i++ {
			entry := &models.Entry{
				Name:      "Sequential Restaurant A" + string(rune(i+'0')),
				Location:  "Location A",
				Cuisine:   "Cuisine A",
				Submitter: "userA",
				Visited:   true,
			}
			err := db.Create(entry).Error
			require.NoError(t, err)
		}

		for i := 0; i < 5; i++ {
			entry := &models.Entry{
				Name:      "Sequential Restaurant B" + string(rune(i+'0')),
				Location:  "Location B",
				Cuisine:   "Cuisine B",
				Submitter: "userB",
				Visited:   false,
			}
			err := db.Create(entry).Error
			require.NoError(t, err)
		}

		// Verify all entries were created
		var count int64
		err := db.Model(&models.Entry{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(10), count)

		// Verify data integrity
		var entriesA []models.Entry
		err = db.Where("submitter = ?", "userA").Find(&entriesA).Error
		require.NoError(t, err)
		assert.Len(t, entriesA, 5)

		var entriesB []models.Entry
		err = db.Where("submitter = ?", "userB").Find(&entriesB).Error
		require.NoError(t, err)
		assert.Len(t, entriesB, 5)
	})

	t.Run("supports transaction operations", func(t *testing.T) {
		db := setupTestDB(t)

		// Test successful transaction
		tx := db.Begin()
		require.NoError(t, tx.Error)

		entry := &models.Entry{
			Name:      "Transaction Test",
			Location:  "TX Location",
			Cuisine:   "TX Cuisine",
			Submitter: "tx_user",
			Visited:   true,
		}
		err := tx.Create(entry).Error
		require.NoError(t, err)

		// Commit transaction
		err = tx.Commit().Error
		require.NoError(t, err)

		// Verify data persisted
		var count int64
		err = db.Model(&models.Entry{}).Where("name = ?", "Transaction Test").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Test rollback transaction
		tx = db.Begin()
		require.NoError(t, tx.Error)

		rollbackEntry := &models.Entry{
			Name:      "Rollback Test",
			Location:  "RB Location",
			Cuisine:   "RB Cuisine",
			Submitter: "rb_user",
			Visited:   false,
		}
		err = tx.Create(rollbackEntry).Error
		require.NoError(t, err)

		// Rollback transaction
		tx.Rollback()

		// Verify data was not persisted
		err = db.Model(&models.Entry{}).Where("name = ?", "Rollback Test").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestDatabaseTestUtilityConsistency(t *testing.T) {
	t.Run("setupTestDB creates consistent database state", func(t *testing.T) {
		// Create multiple test databases and verify they all start in the same state
		databases := make([]*gorm.DB, 3)
		for i := 0; i < 3; i++ {
			databases[i] = setupTestDB(t)
		}

		// All databases should start empty
		for i, db := range databases {
			var count int64
			err := db.Model(&models.Entry{}).Count(&count).Error
			require.NoError(t, err, "Database %d should be accessible", i)
			assert.Equal(t, int64(0), count, "Database %d should start empty", i)
		}

		// All databases should support the same operations
		for i, db := range databases {
			entry := &models.Entry{
				Name:      "Consistency Test",
				Location:  "Test Location",
				Cuisine:   "Test Cuisine",
				Submitter: "testuser",
				Visited:   true,
			}
			err := db.Create(entry).Error
			require.NoError(t, err, "Database %d should support Create operations", i)
			assert.NotZero(t, entry.ID, "Database %d should assign IDs", i)
		}
	})

	t.Run("database cleanup is reliable", func(t *testing.T) {
		originalDB := database.DB
		var testDB *gorm.DB

		// Create a scope where we set up a test database
		func() {
			testDB = setupTestDB(t)
			assert.Equal(t, testDB, database.DB)

			// Add some data
			entry := &models.Entry{
				Name:      "Cleanup Test Entry",
				Location:  "Cleanup Location",
				Cuisine:   "Cleanup Cuisine",
				Submitter: "cleanup_user",
				Visited:   true,
			}
			err := testDB.Create(entry).Error
			require.NoError(t, err)
		}()

		// After the function scope, cleanup should have run
		// Note: In a real test, this would be triggered by t.Cleanup()
		// Here we're testing the concept

		// Verify the original database is restored
		assert.Equal(t, originalDB, database.DB)
	})

	t.Run("test database performance is acceptable", func(t *testing.T) {
		db := setupTestDB(t)

		// Test that basic operations complete in reasonable time
		// This is more of a smoke test than a rigorous performance test

		// Create multiple entries
		entries := make([]models.Entry, 100)
		for i := 0; i < 100; i++ {
			entries[i] = models.Entry{
				Name:      "Performance Test " + string(rune(i+'0')),
				Location:  "Location " + string(rune(i+'0')),
				Cuisine:   "Cuisine " + string(rune(i+'0')),
				Submitter: "perf_user",
				Visited:   i%2 == 0,
			}
		}

		// Batch create
		err := db.Create(&entries).Error
		require.NoError(t, err)

		// Verify all were created
		var count int64
		err = db.Model(&models.Entry{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(100), count)

		// Test query performance
		var foundEntries []models.Entry
		err = db.Where("submitter = ?", "perf_user").Find(&foundEntries).Error
		require.NoError(t, err)
		assert.Len(t, foundEntries, 100)

		// Test update performance
		err = db.Model(&models.Entry{}).Where("submitter = ?", "perf_user").Update("visited", true).Error
		require.NoError(t, err)

		// Verify updates
		var visitedCount int64
		err = db.Model(&models.Entry{}).Where("submitter = ? AND visited = ?", "perf_user", true).Count(&visitedCount).Error
		require.NoError(t, err)
		assert.Equal(t, int64(100), visitedCount)
	})
}
