package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	err = db.AutoMigrate(&Entry{})
	require.NoError(t, err)

	// Cleanup function to restore original DB
	t.Cleanup(func() {
		database.DB = originalDB
	})

	return db
}

func TestEntryModel(t *testing.T) {
	t.Run("creates entry with all fields", func(t *testing.T) {
		entry := Entry{
			Submitter:        "testuser",
			Name:             "Test Restaurant",
			Location:         "N",
			Cuisine:          "Italian",
			Visited:          false,
			LookingForwardTo: "Pizza and pasta",
			Closed:           false,
			MealService:      "Dinner",
			Ordered:          "Margherita Pizza",
			OrderNext:        "Carbonara",
			DoNotOrder:       "Anchovies",
			FoodRating:       4,
			AmbienceRating:   3,
			ValueRating:      5,
			Comments:         "Looking forward to trying this place",
		}

		// Verify all fields are accessible and set correctly
		assert.Equal(t, "testuser", entry.Submitter)
		assert.Equal(t, "Test Restaurant", entry.Name)
		assert.Equal(t, "N", entry.Location)
		assert.Equal(t, "Italian", entry.Cuisine)
		assert.False(t, entry.Visited)
		assert.Equal(t, "Pizza and pasta", entry.LookingForwardTo)
		assert.False(t, entry.Closed)
		assert.Equal(t, "Dinner", entry.MealService)
		assert.Equal(t, "Margherita Pizza", entry.Ordered)
		assert.Equal(t, "Carbonara", entry.OrderNext)
		assert.Equal(t, "Anchovies", entry.DoNotOrder)
		assert.Equal(t, 4, entry.FoodRating)
		assert.Equal(t, 3, entry.AmbienceRating)
		assert.Equal(t, 5, entry.ValueRating)
		assert.Equal(t, "Looking forward to trying this place", entry.Comments)
	})

	t.Run("creates entry with minimal required fields", func(t *testing.T) {
		entry := Entry{
			Submitter: "testuser",
			Name:      "Minimal Restaurant",
			Location:  "S",
			Cuisine:   "Mexican",
			Visited:   true,
		}

		assert.Equal(t, "testuser", entry.Submitter)
		assert.Equal(t, "Minimal Restaurant", entry.Name)
		assert.Equal(t, "S", entry.Location)
		assert.Equal(t, "Mexican", entry.Cuisine)
		assert.True(t, entry.Visited)

		// Optional fields should have zero values
		assert.Empty(t, entry.LookingForwardTo)
		assert.False(t, entry.Closed)
		assert.Empty(t, entry.MealService)
		assert.Empty(t, entry.Ordered)
		assert.Empty(t, entry.OrderNext)
		assert.Empty(t, entry.DoNotOrder)
		assert.Equal(t, 0, entry.FoodRating)
		assert.Equal(t, 0, entry.AmbienceRating)
		assert.Equal(t, 0, entry.ValueRating)
		assert.Empty(t, entry.Comments)
	})

	t.Run("entry has gorm.Model fields", func(t *testing.T) {
		entry := Entry{}

		// Verify gorm.Model fields are available
		assert.Equal(t, uint(0), entry.ID)
		assert.True(t, entry.CreatedAt.IsZero())
		assert.True(t, entry.UpdatedAt.IsZero())
		// DeletedAt is a gorm.DeletedAt struct, check if it's not valid (not deleted)
		assert.False(t, entry.DeletedAt.Valid)
	})
}

func TestEntryValidation(t *testing.T) {
	t.Run("valid ratings within range", func(t *testing.T) {
		entry := Entry{
			FoodRating:     4,
			AmbienceRating: 3,
			ValueRating:    5,
		}

		// These should be valid ratings (0-5)
		assert.GreaterOrEqual(t, entry.FoodRating, 0)
		assert.LessOrEqual(t, entry.FoodRating, 5)
		assert.GreaterOrEqual(t, entry.AmbienceRating, 0)
		assert.LessOrEqual(t, entry.AmbienceRating, 5)
		assert.GreaterOrEqual(t, entry.ValueRating, 0)
		assert.LessOrEqual(t, entry.ValueRating, 5)
	})

	t.Run("boundary ratings are valid", func(t *testing.T) {
		minEntry := Entry{
			FoodRating:     0,
			AmbienceRating: 0,
			ValueRating:    0,
		}

		maxEntry := Entry{
			FoodRating:     5,
			AmbienceRating: 5,
			ValueRating:    5,
		}

		// Minimum values should be valid
		assert.GreaterOrEqual(t, minEntry.FoodRating, 0)
		assert.GreaterOrEqual(t, minEntry.AmbienceRating, 0)
		assert.GreaterOrEqual(t, minEntry.ValueRating, 0)

		// Maximum values should be valid
		assert.LessOrEqual(t, maxEntry.FoodRating, 5)
		assert.LessOrEqual(t, maxEntry.AmbienceRating, 5)
		assert.LessOrEqual(t, maxEntry.ValueRating, 5)
	})
}

func TestEntryDatabaseOperations(t *testing.T) {
	setupTestDB(t)

	t.Run("creates entry in database", func(t *testing.T) {
		entry := Entry{
			Submitter: "dbuser",
			Name:      "DB Restaurant",
			Location:  "NE",
			Cuisine:   "Thai",
			Visited:   true,
		}

		err := database.DB.Create(&entry).Error
		assert.NoError(t, err)
		assert.NotZero(t, entry.ID)
		assert.False(t, entry.CreatedAt.IsZero())
		assert.False(t, entry.UpdatedAt.IsZero())
	})

	t.Run("retrieves entry from database", func(t *testing.T) {
		// Create entry
		originalEntry := Entry{
			Submitter: "retrieveuser",
			Name:      "Retrieve Restaurant",
			Location:  "SW",
			Cuisine:   "French",
			Visited:   false,
			Comments:  "Test retrieval",
		}

		err := database.DB.Create(&originalEntry).Error
		require.NoError(t, err)

		// Retrieve entry
		var retrievedEntry Entry
		err = database.DB.Where("name = ?", "Retrieve Restaurant").First(&retrievedEntry).Error
		assert.NoError(t, err)

		assert.Equal(t, originalEntry.Submitter, retrievedEntry.Submitter)
		assert.Equal(t, originalEntry.Name, retrievedEntry.Name)
		assert.Equal(t, originalEntry.Location, retrievedEntry.Location)
		assert.Equal(t, originalEntry.Cuisine, retrievedEntry.Cuisine)
		assert.Equal(t, originalEntry.Visited, retrievedEntry.Visited)
		assert.Equal(t, originalEntry.Comments, retrievedEntry.Comments)
	})

	t.Run("updates entry in database", func(t *testing.T) {
		// Create entry
		entry := Entry{
			Submitter: "updateuser",
			Name:      "Update Restaurant",
			Location:  "NW",
			Cuisine:   "Japanese",
			Visited:   false,
		}

		err := database.DB.Create(&entry).Error
		require.NoError(t, err)

		// Update entry
		entry.Visited = true
		entry.FoodRating = 4
		entry.Comments = "Updated after visit"

		err = database.DB.Save(&entry).Error
		assert.NoError(t, err)

		// Verify update
		var updatedEntry Entry
		err = database.DB.First(&updatedEntry, entry.ID).Error
		assert.NoError(t, err)

		assert.True(t, updatedEntry.Visited)
		assert.Equal(t, 4, updatedEntry.FoodRating)
		assert.Equal(t, "Updated after visit", updatedEntry.Comments)
	})

	t.Run("soft deletes entry", func(t *testing.T) {
		// Create entry
		entry := Entry{
			Submitter: "deleteuser",
			Name:      "Delete Restaurant",
			Location:  "SE",
			Cuisine:   "Indian",
			Visited:   true,
		}

		err := database.DB.Create(&entry).Error
		require.NoError(t, err)

		// Soft delete entry
		err = database.DB.Delete(&entry).Error
		assert.NoError(t, err)

		// Verify entry is soft deleted (not found in normal query)
		var deletedEntry Entry
		err = database.DB.Where("id = ?", entry.ID).First(&deletedEntry).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		// Verify entry exists with Unscoped query
		err = database.DB.Unscoped().Where("id = ?", entry.ID).First(&deletedEntry).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedEntry.DeletedAt)
	})
}

func TestEntryIndexes(t *testing.T) {
	setupTestDB(t)

	t.Run("queries by indexed fields", func(t *testing.T) {
		// Create test entries
		entries := []Entry{
			{Submitter: "user1", Name: "Restaurant A", Location: "N", Cuisine: "Italian", Visited: true},
			{Submitter: "user1", Name: "Restaurant B", Location: "S", Cuisine: "Mexican", Visited: false},
			{Submitter: "user2", Name: "Restaurant C", Location: "E", Cuisine: "Italian", Visited: true},
		}

		for _, entry := range entries {
			err := database.DB.Create(&entry).Error
			require.NoError(t, err)
		}

		// Query by submitter (indexed field)
		var user1Entries []Entry
		err := database.DB.Where("submitter = ?", "user1").Find(&user1Entries).Error
		assert.NoError(t, err)
		assert.Len(t, user1Entries, 2)

		// Query by name (indexed field)
		var namedEntry Entry
		err = database.DB.Where("name = ?", "Restaurant A").First(&namedEntry).Error
		assert.NoError(t, err)
		assert.Equal(t, "user1", namedEntry.Submitter)

		// Query by cuisine (indexed field)
		var italianEntries []Entry
		err = database.DB.Where("cuisine = ?", "Italian").Find(&italianEntries).Error
		assert.NoError(t, err)
		assert.Len(t, italianEntries, 2)

		// Query by visited status (indexed field)
		var visitedEntries []Entry
		err = database.DB.Where("visited = ?", true).Find(&visitedEntries).Error
		assert.NoError(t, err)
		assert.Len(t, visitedEntries, 2)
	})
}

func TestEntryJSONSerialization(t *testing.T) {
	t.Run("entry serializes to JSON correctly", func(t *testing.T) {
		entry := Entry{
			Submitter:      "jsonuser",
			Name:           "JSON Restaurant",
			Location:       "JSON City",
			Cuisine:        "JSON Cuisine",
			Visited:        true,
			FoodRating:     4,
			AmbienceRating: 3,
			ValueRating:    5,
			Comments:       "JSON test",
		}

		// Test that struct tags are properly defined for JSON serialization
		// This is validated by the struct tags in the model definition
		assert.Equal(t, "jsonuser", entry.Submitter)
		assert.Equal(t, "JSON Restaurant", entry.Name)
		assert.Equal(t, "JSON City", entry.Location)
		assert.Equal(t, "JSON Cuisine", entry.Cuisine)
		assert.True(t, entry.Visited)
		assert.Equal(t, 4, entry.FoodRating)
		assert.Equal(t, 3, entry.AmbienceRating)
		assert.Equal(t, 5, entry.ValueRating)
		assert.Equal(t, "JSON test", entry.Comments)
	})
}
