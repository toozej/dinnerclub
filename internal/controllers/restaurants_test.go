package controllers

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFindRestaurants_SortingLogic(t *testing.T) {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("failed to connect database")
	}

	// Migrate the schema
	if err := db.AutoMigrate(&models.Entry{}, &models.Restaurant{}); err != nil {
		t.Fatal("failed to migrate database:", err)
	}

	// Set the global database
	database.DB = db

	// Create test restaurants
	restaurant1 := models.Restaurant{
		Name:     "Test Restaurant 1",
		Location: "N",
		Cuisine:  "Italian",
		Closed:   false,
	}
	restaurant2 := models.Restaurant{
		Name:     "Test Restaurant 2",
		Location: "S",
		Cuisine:  "Mexican",
		Closed:   false,
	}
	restaurant3 := models.Restaurant{
		Name:     "Test Restaurant 3",
		Location: "NE",
		Cuisine:  "Thai",
		Closed:   false,
	}

	database.DB.Create(&restaurant1)
	database.DB.Create(&restaurant2)
	database.DB.Create(&restaurant3)

	// Create test entries - restaurant1 gets 2 entries, restaurant2 gets 1, restaurant3 gets 0
	entry1 := models.Entry{
		Submitter: "testuser1",
		Name:      "Test Restaurant 1",
		Location:  "N",
		Cuisine:   "Italian",
		Visited:   true,
	}
	entry2 := models.Entry{
		Submitter: "testuser2",
		Name:      "Test Restaurant 1",
		Location:  "N",
		Cuisine:   "Italian",
		Visited:   true,
	}
	entry3 := models.Entry{
		Submitter: "testuser3",
		Name:      "Test Restaurant 2",
		Location:  "S",
		Cuisine:   "Mexican",
		Visited:   true,
	}

	database.DB.Create(&entry1)
	database.DB.Create(&entry2)
	database.DB.Create(&entry3)

	// Test the sorting logic directly by simulating what FindRestaurants does
	var allRestaurants []models.Restaurant
	database.DB.Find(&allRestaurants)

	// Calculate entry counts and create RestaurantWithEntryCount slice
	var restaurantsWithCounts []RestaurantWithEntryCount
	for _, restaurant := range allRestaurants {
		var entryCount int64
		database.DB.Model(&models.Entry{}).Where("name = ?", restaurant.Name).Count(&entryCount)

		restaurantsWithCounts = append(restaurantsWithCounts, RestaurantWithEntryCount{
			Restaurant: restaurant,
			EntryCount: entryCount,
		})
	}

	// Sort by entry count (ascending)
	sort.Slice(restaurantsWithCounts, func(i, j int) bool {
		return restaurantsWithCounts[i].EntryCount < restaurantsWithCounts[j].EntryCount
	})

	// Verify sorting: restaurant3 (0 entries) should be first, restaurant2 (1 entry) second, restaurant1 (2 entries) last
	assert.Equal(t, 3, len(restaurantsWithCounts))
	assert.Equal(t, "Test Restaurant 3", restaurantsWithCounts[0].Name)
	assert.Equal(t, int64(0), restaurantsWithCounts[0].EntryCount)
	assert.Equal(t, "Test Restaurant 2", restaurantsWithCounts[1].Name)
	assert.Equal(t, int64(1), restaurantsWithCounts[1].EntryCount)
	assert.Equal(t, "Test Restaurant 1", restaurantsWithCounts[2].Name)
	assert.Equal(t, int64(2), restaurantsWithCounts[2].EntryCount)
}

func TestRestaurantWithEntryCountStruct(t *testing.T) {
	// Test that the struct works correctly
	restaurant := models.Restaurant{
		Name:     "Test Restaurant",
		Location: "N",
		Cuisine:  "Italian",
		Closed:   false,
	}

	restaurantWithCount := RestaurantWithEntryCount{
		Restaurant: restaurant,
		EntryCount: 5,
	}

	assert.Equal(t, "Test Restaurant", restaurantWithCount.Name)
	assert.Equal(t, "N", restaurantWithCount.Location)
	assert.Equal(t, "Italian", restaurantWithCount.Cuisine)
	assert.Equal(t, false, restaurantWithCount.Closed)
	assert.Equal(t, int64(5), restaurantWithCount.EntryCount)
}
