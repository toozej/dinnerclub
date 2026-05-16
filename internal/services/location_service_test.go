package services

import (
	"math"
	"testing"

	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLocationTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate the schema
	if err := db.AutoMigrate(&models.Restaurant{}, &models.Entry{}, &models.User{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestCalculateDistance(t *testing.T) {
	db := setupLocationTestDB(t)
	service := NewLocationService(db)

	tests := []struct {
		name     string
		loc1     Location
		loc2     Location
		expected float64
		delta    float64 // acceptable error margin
	}{
		{
			name:     "Same location",
			loc1:     Location{Latitude: 40.7128, Longitude: -74.0060},
			loc2:     Location{Latitude: 40.7128, Longitude: -74.0060},
			expected: 0,
			delta:    0.01,
		},
		{
			name:     "New York to Los Angeles",
			loc1:     Location{Latitude: 40.7128, Longitude: -74.0060},
			loc2:     Location{Latitude: 34.0522, Longitude: -118.2437},
			expected: 3936, // approximately 3936 km
			delta:    50,   // 50 km margin
		},
		{
			name:     "London to Paris",
			loc1:     Location{Latitude: 51.5074, Longitude: -0.1278},
			loc2:     Location{Latitude: 48.8566, Longitude: 2.3522},
			expected: 344, // approximately 344 km
			delta:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := service.CalculateDistance(tt.loc1, tt.loc2)

			if math.Abs(distance-tt.expected) > tt.delta {
				t.Errorf("CalculateDistance() = %v, want %v (±%v)", distance, tt.expected, tt.delta)
			}
		})
	}
}

func TestCalculateDistanceSymmetry(t *testing.T) {
	db := setupLocationTestDB(t)
	service := NewLocationService(db)

	loc1 := Location{Latitude: 40.7128, Longitude: -74.0060}
	loc2 := Location{Latitude: 34.0522, Longitude: -118.2437}

	distance1 := service.CalculateDistance(loc1, loc2)
	distance2 := service.CalculateDistance(loc2, loc1)

	if math.Abs(distance1-distance2) > 0.001 {
		t.Errorf("Distance calculation not symmetric: %v vs %v", distance1, distance2)
	}
}

func TestFindNearbyRestaurants(t *testing.T) {
	db := setupLocationTestDB(t)
	service := NewLocationService(db)

	// Create test restaurants
	restaurants := []models.Restaurant{
		{Name: "Restaurant A", Location: "New York", Cuisine: "Italian"},
		{Name: "Restaurant B", Location: "Brooklyn", Cuisine: "Mexican"},
		{Name: "Restaurant C", Location: "Queens", Cuisine: "Chinese"},
	}

	for _, r := range restaurants {
		if err := db.Create(&r).Error; err != nil {
			t.Fatalf("Failed to create test restaurant: %v", err)
		}
	}

	userLocation := Location{Latitude: 40.7128, Longitude: -74.0060}
	radius := 50.0 // 50 km

	results, err := service.FindNearbyRestaurants(userLocation, radius)
	if err != nil {
		t.Fatalf("FindNearbyRestaurants() error = %v", err)
	}

	// Note: Results will be empty because we don't have actual coordinates
	// This test validates the function doesn't error
	if results == nil {
		t.Error("FindNearbyRestaurants() returned nil")
	}
}

func TestFindNearestWishlist(t *testing.T) {
	db := setupLocationTestDB(t)
	service := NewLocationService(db)

	// Create test user
	user := models.User{
		Username:  "testuser",
		Password:  "password123",
		Email:     "test@example.com",
		Firstname: "Test",
		Lastname:  "User",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test restaurants
	restaurant := models.Restaurant{
		Name:     "Wishlist Restaurant",
		Location: "New York",
		Cuisine:  "Italian",
	}
	if err := db.Create(&restaurant).Error; err != nil {
		t.Fatalf("Failed to create test restaurant: %v", err)
	}

	// Create wishlist entry (visited = false)
	entry := models.Entry{
		Submitter: "testuser",
		Name:      "Wishlist Restaurant",
		Location:  "New York",
		Cuisine:   "Italian",
		Visited:   false,
	}
	if err := db.Create(&entry).Error; err != nil {
		t.Fatalf("Failed to create test entry: %v", err)
	}

	userLocation := Location{Latitude: 40.7128, Longitude: -74.0060}

	results, err := service.FindNearestWishlist(user.ID, userLocation)
	if err != nil {
		t.Fatalf("FindNearestWishlist() error = %v", err)
	}

	if results == nil {
		t.Error("FindNearestWishlist() returned nil")
	}

	// Verify wishlist flag is set
	for _, result := range results {
		if !result.Wishlist {
			t.Error("Expected Wishlist flag to be true")
		}
		if result.Visited {
			t.Error("Expected Visited flag to be false for wishlist items")
		}
	}
}

func TestFindNearestUnvisited(t *testing.T) {
	db := setupLocationTestDB(t)
	service := NewLocationService(db)

	// Create test user
	user := models.User{
		Username:  "testuser",
		Password:  "password123",
		Email:     "test@example.com",
		Firstname: "Test",
		Lastname:  "User",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test restaurants
	visitedRestaurant := models.Restaurant{
		Name:     "Visited Restaurant",
		Location: "New York",
		Cuisine:  "Italian",
	}
	if err := db.Create(&visitedRestaurant).Error; err != nil {
		t.Fatalf("Failed to create visited restaurant: %v", err)
	}

	unvisitedRestaurant := models.Restaurant{
		Name:     "Unvisited Restaurant",
		Location: "Brooklyn",
		Cuisine:  "Mexican",
	}
	if err := db.Create(&unvisitedRestaurant).Error; err != nil {
		t.Fatalf("Failed to create unvisited restaurant: %v", err)
	}

	// Create visited entry
	visitedEntry := models.Entry{
		Submitter: "testuser",
		Name:      "Visited Restaurant",
		Location:  "New York",
		Cuisine:   "Italian",
		Visited:   true,
	}
	if err := db.Create(&visitedEntry).Error; err != nil {
		t.Fatalf("Failed to create visited entry: %v", err)
	}

	userLocation := Location{Latitude: 40.7128, Longitude: -74.0060}

	results, err := service.FindNearestUnvisited(user.ID, userLocation)
	if err != nil {
		t.Fatalf("FindNearestUnvisited() error = %v", err)
	}

	if results == nil {
		t.Error("FindNearestUnvisited() returned nil")
	}

	// Verify visited restaurants are excluded
	for _, result := range results {
		if result.Restaurant.Name == "Visited Restaurant" {
			t.Error("Visited restaurant should not be in unvisited results")
		}
	}
}

func TestFindNearbyRestaurantsRadius(t *testing.T) {
	db := setupLocationTestDB(t)
	service := NewLocationService(db)

	userLocation := Location{Latitude: 40.7128, Longitude: -74.0060}

	// Test with different radii
	radii := []float64{10, 50, 100, 500}

	for _, radius := range radii {
		results, err := service.FindNearbyRestaurants(userLocation, radius)
		if err != nil {
			t.Errorf("FindNearbyRestaurants(radius=%v) error = %v", radius, err)
		}

		// Verify all results are within radius
		for _, result := range results {
			if result.Distance > radius {
				t.Errorf("Restaurant distance %v exceeds radius %v", result.Distance, radius)
			}
		}
	}
}

func TestFindNearbyRestaurantsSorting(t *testing.T) {
	db := setupLocationTestDB(t)
	service := NewLocationService(db)

	userLocation := Location{Latitude: 40.7128, Longitude: -74.0060}
	radius := 1000.0

	results, err := service.FindNearbyRestaurants(userLocation, radius)
	if err != nil {
		t.Fatalf("FindNearbyRestaurants() error = %v", err)
	}

	// Verify results are sorted by distance
	for i := 1; i < len(results); i++ {
		if results[i-1].Distance > results[i].Distance {
			t.Errorf("Results not sorted: distance[%d]=%v > distance[%d]=%v",
				i-1, results[i-1].Distance, i, results[i].Distance)
		}
	}
}
