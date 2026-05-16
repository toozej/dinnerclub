package aggregation

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.Restaurant{}, &models.Entry{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestData(t *testing.T, db *gorm.DB) {
	// Create test restaurant
	restaurant := models.Restaurant{
		Name:     "Test Restaurant",
		Location: "N",
		Cuisine:  "Italian",
		Closed:   false,
	}
	if err := db.Create(&restaurant).Error; err != nil {
		t.Fatalf("Failed to create test restaurant: %v", err)
	}

	// Create test entries with different timestamps
	entries := []models.Entry{
		{
			Model:          gorm.Model{CreatedAt: time.Now().Add(-3 * time.Hour)},
			Submitter:      "user1",
			Name:           "Test Restaurant",
			Location:       "N",
			Cuisine:        "Italian",
			Visited:        true,
			Ordered:        "pasta, pizza, salad",
			FoodRating:     4,
			AmbienceRating: 3,
			ValueRating:    5,
		},
		{
			Model:          gorm.Model{CreatedAt: time.Now().Add(-2 * time.Hour)},
			Submitter:      "user2",
			Name:           "Test Restaurant",
			Location:       "N",
			Cuisine:        "Italian",
			Visited:        true,
			Ordered:        "pizza, bread, pasta",
			FoodRating:     5,
			AmbienceRating: 4,
			ValueRating:    4,
		},
		{
			Model:          gorm.Model{CreatedAt: time.Now().Add(-1 * time.Hour)},
			Submitter:      "user3",
			Name:           "Test Restaurant",
			Location:       "N",
			Cuisine:        "Italian",
			Visited:        true,
			Ordered:        "pizza, wine",
			FoodRating:     3,
			AmbienceRating: 5,
			ValueRating:    3,
		},
		{
			Model:     gorm.Model{CreatedAt: time.Now()},
			Submitter: "user4",
			Name:      "Test Restaurant",
			Location:  "N",
			Cuisine:   "Italian",
			Visited:   false,
			// No ratings for unvisited entry
		},
	}

	for _, entry := range entries {
		if err := db.Create(&entry).Error; err != nil {
			t.Fatalf("Failed to create test entry: %v", err)
		}
	}
}

func TestNewRestaurantAggregationService(t *testing.T) {
	db := setupTestDB(t)
	service := NewRestaurantAggregationService(db)

	if service == nil {
		t.Error("Expected service to be created, got nil")
		return
	}

	if service.db != db {
		t.Error("Expected service to have correct database reference")
	}
}

func TestCalculatePopularItems(t *testing.T) {
	db := setupTestDB(t)
	service := NewRestaurantAggregationService(db)

	entries := []models.Entry{
		{Ordered: "pasta, pizza, salad"},
		{Ordered: "pizza, bread, pasta"},
		{Ordered: "pizza, wine"},
		{Ordered: ""}, // Empty ordered field
	}

	popularItems := service.CalculatePopularItems(entries)

	// Expected: pizza (3), pasta (2), bread (1), salad (1), wine (1)
	expectedItems := map[string]int{
		"pizza": 3,
		"pasta": 2,
		"bread": 1,
		"salad": 1,
		"wine":  1,
	}

	if len(popularItems) != len(expectedItems) {
		t.Errorf("Expected %d popular items, got %d", len(expectedItems), len(popularItems))
	}

	// Check that pizza is the most popular (should be first)
	if len(popularItems) > 0 && popularItems[0].Name != "pizza" {
		t.Errorf("Expected pizza to be most popular, got %s", popularItems[0].Name)
	}

	if len(popularItems) > 0 && popularItems[0].Frequency != 3 {
		t.Errorf("Expected pizza frequency to be 3, got %d", popularItems[0].Frequency)
	}

	// Verify all expected items are present
	itemMap := make(map[string]int)
	for _, item := range popularItems {
		itemMap[item.Name] = item.Frequency
	}

	for expectedName, expectedFreq := range expectedItems {
		if freq, exists := itemMap[expectedName]; !exists {
			t.Errorf("Expected item %s not found in popular items", expectedName)
		} else if freq != expectedFreq {
			t.Errorf("Expected frequency %d for %s, got %d", expectedFreq, expectedName, freq)
		}
	}
}

func TestCalculateAverageRatings(t *testing.T) {
	db := setupTestDB(t)
	service := NewRestaurantAggregationService(db)

	entries := []models.Entry{
		{FoodRating: 4, AmbienceRating: 3, ValueRating: 5},
		{FoodRating: 5, AmbienceRating: 4, ValueRating: 4},
		{FoodRating: 3, AmbienceRating: 5, ValueRating: 3},
		{FoodRating: 0, AmbienceRating: 0, ValueRating: 0}, // Zero ratings should be excluded
	}

	averages := service.CalculateAverageRatings(entries)

	// Expected averages: Food (4+5+3)/3 = 4.0, Ambience (3+4+5)/3 = 4.0, Value (5+4+3)/3 = 4.0
	expectedFood := 4.0
	expectedAmbience := 4.0
	expectedValue := 4.0

	if averages.Food != expectedFood {
		t.Errorf("Expected food average %.2f, got %.2f", expectedFood, averages.Food)
	}

	if averages.Ambience != expectedAmbience {
		t.Errorf("Expected ambience average %.2f, got %.2f", expectedAmbience, averages.Ambience)
	}

	if averages.Value != expectedValue {
		t.Errorf("Expected value average %.2f, got %.2f", expectedValue, averages.Value)
	}
}

func TestCalculateAverageRatingsWithZeros(t *testing.T) {
	db := setupTestDB(t)
	service := NewRestaurantAggregationService(db)

	// Test with all zero ratings
	entries := []models.Entry{
		{FoodRating: 0, AmbienceRating: 0, ValueRating: 0},
		{FoodRating: 0, AmbienceRating: 0, ValueRating: 0},
	}

	averages := service.CalculateAverageRatings(entries)

	if averages.Food != 0.0 {
		t.Errorf("Expected food average 0.0 for all zero ratings, got %.2f", averages.Food)
	}

	if averages.Ambience != 0.0 {
		t.Errorf("Expected ambience average 0.0 for all zero ratings, got %.2f", averages.Ambience)
	}

	if averages.Value != 0.0 {
		t.Errorf("Expected value average 0.0 for all zero ratings, got %.2f", averages.Value)
	}
}

func TestGetRecentEntriesByRestaurant(t *testing.T) {
	db := setupTestDB(t)
	createTestData(t, db)
	service := NewRestaurantAggregationService(db)

	// Test without limit
	entries, err := service.GetRecentEntriesByRestaurant("Test Restaurant", 0)
	if err != nil {
		t.Fatalf("Failed to get recent entries: %v", err)
	}

	if len(entries) != 4 {
		t.Errorf("Expected 4 entries, got %d", len(entries))
	}

	// Verify entries are sorted by created_at DESC (most recent first)
	for i := 1; i < len(entries); i++ {
		if entries[i-1].CreatedAt.Before(entries[i].CreatedAt) {
			t.Error("Entries are not sorted by created_at DESC")
		}
	}

	// Test with limit
	limitedEntries, err := service.GetRecentEntriesByRestaurant("Test Restaurant", 2)
	if err != nil {
		t.Fatalf("Failed to get limited recent entries: %v", err)
	}

	if len(limitedEntries) != 2 {
		t.Errorf("Expected 2 entries with limit, got %d", len(limitedEntries))
	}
}

func TestGetEntryCountByRestaurant(t *testing.T) {
	db := setupTestDB(t)
	createTestData(t, db)
	service := NewRestaurantAggregationService(db)

	count, err := service.GetEntryCountByRestaurant("Test Restaurant")
	if err != nil {
		t.Fatalf("Failed to get entry count: %v", err)
	}

	if count != 4 {
		t.Errorf("Expected entry count 4, got %d", count)
	}

	// Test with non-existent restaurant
	count, err = service.GetEntryCountByRestaurant("Non-existent Restaurant")
	if err != nil {
		t.Fatalf("Failed to get entry count for non-existent restaurant: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected entry count 0 for non-existent restaurant, got %d", count)
	}
}

func TestGetRestaurantAggregateData(t *testing.T) {
	db := setupTestDB(t)
	createTestData(t, db)
	service := NewRestaurantAggregationService(db)

	aggregateData, err := service.GetRestaurantAggregateData("Test Restaurant")
	if err != nil {
		t.Fatalf("Failed to get restaurant aggregate data: %v", err)
	}

	// Verify restaurant data
	if aggregateData.Restaurant.Name != "Test Restaurant" {
		t.Errorf("Expected restaurant name 'Test Restaurant', got '%s'", aggregateData.Restaurant.Name)
	}

	// Verify entry count
	if aggregateData.EntryCount != 4 {
		t.Errorf("Expected entry count 4, got %d", aggregateData.EntryCount)
	}

	// Verify popular items (pizza should be most popular)
	if len(aggregateData.PopularItems) == 0 {
		t.Error("Expected popular items, got none")
	} else if aggregateData.PopularItems[0].Name != "pizza" {
		t.Errorf("Expected most popular item to be 'pizza', got '%s'", aggregateData.PopularItems[0].Name)
	}

	// Verify average ratings are calculated
	if aggregateData.AverageRatings.Food == 0.0 {
		t.Error("Expected non-zero average food rating")
	}

	// Verify recent entries (should be limited to 10)
	if len(aggregateData.RecentEntries) > 10 {
		t.Errorf("Expected at most 10 recent entries, got %d", len(aggregateData.RecentEntries))
	}
}

func TestGetRestaurantAggregateDataNonExistent(t *testing.T) {
	db := setupTestDB(t)
	service := NewRestaurantAggregationService(db)

	_, err := service.GetRestaurantAggregateData("Non-existent Restaurant")
	if err == nil {
		t.Error("Expected error for non-existent restaurant, got nil")
	}
}
