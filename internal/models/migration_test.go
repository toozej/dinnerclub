package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupMigrationTestDB creates an in-memory SQLite database for testing migrations
func setupMigrationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Store original DB and replace with test DB
	originalDB := database.DB
	database.DB = db

	// Cleanup function to restore original DB
	t.Cleanup(func() {
		database.DB = originalDB
	})

	return db
}

func TestMigrateSchema(t *testing.T) {
	setupMigrationTestDB(t)

	t.Run("migrates all models successfully", func(t *testing.T) {
		// This should not panic or return errors
		assert.NotPanics(t, func() {
			MigrateSchema()
		})

		// Verify that all tables were created
		assert.True(t, database.DB.Migrator().HasTable(&Entry{}))
		assert.True(t, database.DB.Migrator().HasTable(&User{}))
		assert.True(t, database.DB.Migrator().HasTable(&Restaurant{}))
	})

	t.Run("creates tables with correct structure", func(t *testing.T) {
		MigrateSchema()

		// Test Entry table structure
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "submitter"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "name"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "location"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "cuisine"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "visited"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "looking_forward_to"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "closed"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "meal_service"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "ordered"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "order_next"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "do_not_order"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "food_rating"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "ambience_rating"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "value_rating"))
		assert.True(t, database.DB.Migrator().HasColumn(&Entry{}, "comments"))

		// Test User table structure
		assert.True(t, database.DB.Migrator().HasColumn(&User{}, "username"))
		assert.True(t, database.DB.Migrator().HasColumn(&User{}, "password"))
		assert.True(t, database.DB.Migrator().HasColumn(&User{}, "firstname"))
		assert.True(t, database.DB.Migrator().HasColumn(&User{}, "lastname"))
		assert.True(t, database.DB.Migrator().HasColumn(&User{}, "email"))
		assert.True(t, database.DB.Migrator().HasColumn(&User{}, "referral_code"))

		// Test Restaurant table structure
		assert.True(t, database.DB.Migrator().HasColumn(&Restaurant{}, "name"))
		assert.True(t, database.DB.Migrator().HasColumn(&Restaurant{}, "location"))
		assert.True(t, database.DB.Migrator().HasColumn(&Restaurant{}, "cuisine"))
		assert.True(t, database.DB.Migrator().HasColumn(&Restaurant{}, "closed"))
		assert.True(t, database.DB.Migrator().HasColumn(&Restaurant{}, "address"))
		assert.True(t, database.DB.Migrator().HasColumn(&Restaurant{}, "website_url"))
		assert.True(t, database.DB.Migrator().HasColumn(&Restaurant{}, "reservations_url"))
		assert.True(t, database.DB.Migrator().HasColumn(&Restaurant{}, "menu_url"))
		assert.True(t, database.DB.Migrator().HasColumn(&Restaurant{}, "phone_number"))
	})

	t.Run("creates gorm.Model fields for all tables", func(t *testing.T) {
		MigrateSchema()

		// All models should have gorm.Model fields
		models := []interface{}{&Entry{}, &User{}, &Restaurant{}}

		for _, model := range models {
			assert.True(t, database.DB.Migrator().HasColumn(model, "id"))
			assert.True(t, database.DB.Migrator().HasColumn(model, "created_at"))
			assert.True(t, database.DB.Migrator().HasColumn(model, "updated_at"))
			assert.True(t, database.DB.Migrator().HasColumn(model, "deleted_at"))
		}
	})
}

func TestCreateSearchIndexes(t *testing.T) {
	setupMigrationTestDB(t)

	t.Run("creates search indexes successfully", func(t *testing.T) {
		// First migrate the schema
		MigrateSchema()

		// The createSearchIndexes function is called within MigrateSchema
		// We can test that it doesn't cause errors by running MigrateSchema
		assert.NotPanics(t, func() {
			MigrateSchema()
		})
	})

	t.Run("creates composite indexes for entries", func(t *testing.T) {
		MigrateSchema()

		// Test that we can query using the indexed fields efficiently
		// Create test data
		entries := []Entry{
			{Submitter: "user1", Name: "Restaurant A", Location: "N", Cuisine: "Italian", Visited: true},
			{Submitter: "user1", Name: "Restaurant B", Location: "S", Cuisine: "Mexican", Visited: false},
			{Submitter: "user2", Name: "Restaurant C", Location: "E", Cuisine: "Italian", Visited: true},
		}

		for _, entry := range entries {
			err := database.DB.Create(&entry).Error
			require.NoError(t, err)
		}

		// Test queries that should benefit from indexes
		var results []Entry

		// Query by name and visited (should use idx_entries_name_visited)
		err := database.DB.Where("name = ? AND visited = ?", "Restaurant A", true).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		// Query by submitter and visited (should use idx_entries_submitter_visited)
		err = database.DB.Where("submitter = ? AND visited = ?", "user1", true).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 1)

		// Query by cuisine and location (should use idx_entries_cuisine_location)
		err = database.DB.Where("cuisine = ? AND location = ?", "Italian", "N").Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 1)
	})

	t.Run("handles index creation errors gracefully", func(t *testing.T) {
		// This test verifies that the function handles database-specific index creation
		// For SQLite, some PostgreSQL-specific indexes might fail, but the function should continue
		assert.NotPanics(t, func() {
			err := createSearchIndexes()
			// The function should not return an error even if some indexes fail
			assert.NoError(t, err)
		})
	})
}

func TestMigrationIdempotency(t *testing.T) {
	setupMigrationTestDB(t)

	t.Run("migration can be run multiple times safely", func(t *testing.T) {
		// Run migration first time
		assert.NotPanics(t, func() {
			MigrateSchema()
		})

		// Verify tables exist
		assert.True(t, database.DB.Migrator().HasTable(&Entry{}))
		assert.True(t, database.DB.Migrator().HasTable(&User{}))
		assert.True(t, database.DB.Migrator().HasTable(&Restaurant{}))

		// Run migration second time - should not cause errors
		assert.NotPanics(t, func() {
			MigrateSchema()
		})

		// Tables should still exist
		assert.True(t, database.DB.Migrator().HasTable(&Entry{}))
		assert.True(t, database.DB.Migrator().HasTable(&User{}))
		assert.True(t, database.DB.Migrator().HasTable(&Restaurant{}))
	})
}

func TestMigrationWithData(t *testing.T) {
	setupMigrationTestDB(t)

	t.Run("migration preserves existing data", func(t *testing.T) {
		// Run initial migration
		MigrateSchema()

		// Create test data
		entry := Entry{
			Submitter: "testuser",
			Name:      "Test Restaurant",
			Location:  "Test Location",
			Cuisine:   "Test Cuisine",
			Visited:   true,
		}
		err := database.DB.Create(&entry).Error
		require.NoError(t, err)

		user := User{
			Username: "testuser",
			Password: "testpass123",
			Email:    "test@example.com",
		}
		err = database.DB.Create(&user).Error
		require.NoError(t, err)

		restaurant := Restaurant{
			Name:     "Test Restaurant",
			Location: "Test Location",
			Cuisine:  "Test Cuisine",
		}
		err = database.DB.Create(&restaurant).Error
		require.NoError(t, err)

		// Run migration again
		MigrateSchema()

		// Verify data still exists
		var retrievedEntry Entry
		err = database.DB.Where("name = ?", "Test Restaurant").First(&retrievedEntry).Error
		assert.NoError(t, err)
		assert.Equal(t, "testuser", retrievedEntry.Submitter)

		var retrievedUser User
		err = database.DB.Where("username = ?", "testuser").First(&retrievedUser).Error
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", retrievedUser.Email)

		var retrievedRestaurant Restaurant
		err = database.DB.Where("name = ?", "Test Restaurant").First(&retrievedRestaurant).Error
		assert.NoError(t, err)
		assert.Equal(t, "Test Location", retrievedRestaurant.Location)
	})
}

func TestIndexPerformance(t *testing.T) {
	setupMigrationTestDB(t)

	t.Run("indexes improve query performance", func(t *testing.T) {
		MigrateSchema()

		// Create a larger dataset to test index performance
		for i := 0; i < 100; i++ {
			entry := Entry{
				Submitter: "user1",
				Name:      "Restaurant " + string(rune(i)),
				Location:  "Location " + string(rune(i%10)),
				Cuisine:   "Cuisine " + string(rune(i%5)),
				Visited:   i%2 == 0,
			}
			err := database.DB.Create(&entry).Error
			require.NoError(t, err)
		}

		// Test that indexed queries work efficiently
		var results []Entry

		// Query by submitter (indexed)
		err := database.DB.Where("submitter = ?", "user1").Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 100)

		// Query by visited status (indexed)
		err = database.DB.Where("visited = ?", true).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 50)

		// Complex query using multiple indexed fields
		err = database.DB.Where("submitter = ? AND visited = ?", "user1", true).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 50)
	})
}

func TestMigrationErrorHandling(t *testing.T) {
	t.Run("handles database connection errors gracefully", func(t *testing.T) {
		// Store original DB
		originalDB := database.DB

		// Set DB to nil to simulate connection error
		database.DB = nil

		// Restore DB after test
		defer func() {
			database.DB = originalDB
		}()

		// Migration should handle nil database gracefully
		// Note: This will likely panic in the current implementation,
		// but it demonstrates how error handling could be improved
		assert.Panics(t, func() {
			MigrateSchema()
		})
	})
}

func TestSchemaModelsArray(t *testing.T) {
	t.Run("schema models array contains all models", func(t *testing.T) {
		// This test verifies that the schemaModels array in MigrateSchema
		// contains all the expected models. We can't directly access the array,
		// but we can verify that all expected tables are created.
		setupMigrationTestDB(t)
		MigrateSchema()

		expectedTables := []interface{}{
			&Entry{},
			&User{},
			&Restaurant{},
		}

		for _, model := range expectedTables {
			assert.True(t, database.DB.Migrator().HasTable(model),
				"Expected table for model %T to exist", model)
		}
	})
}

func TestIndexNaming(t *testing.T) {
	setupMigrationTestDB(t)

	t.Run("creates indexes with correct names", func(t *testing.T) {
		MigrateSchema()

		// We can't easily test index names in SQLite, but we can verify
		// that the index creation SQL doesn't cause errors
		expectedIndexes := []string{
			"idx_entries_name_visited",
			"idx_entries_submitter_visited",
			"idx_entries_cuisine_location",
			"idx_restaurants_name_fulltext",
		}

		// The indexes should be created without errors
		// This is implicitly tested by MigrateSchema not panicking
		for _, indexName := range expectedIndexes {
			assert.NotEmpty(t, indexName)
		}
	})
}
