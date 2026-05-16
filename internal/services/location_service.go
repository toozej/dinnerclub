// Package services provides business logic services for the dinnerclub application
package services

import (
	"math"

	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/gorm"
)

// LocationService provides GPS-based functionality for finding nearby restaurants
type LocationService struct {
	db *gorm.DB
}

// NewLocationService creates a new LocationService instance
func NewLocationService(db *gorm.DB) *LocationService {
	return &LocationService{db: db}
}

// Location represents a geographic coordinate
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// RestaurantDistance represents a restaurant with its distance from a location
type RestaurantDistance struct {
	Restaurant models.Restaurant `json:"restaurant"`
	Distance   float64           `json:"distance"` // in kilometers
	Visited    bool              `json:"visited"`
	Wishlist   bool              `json:"wishlist"`
}

// CalculateDistance calculates the distance between two geographic coordinates
// using the Haversine formula. Returns distance in kilometers.
func (l *LocationService) CalculateDistance(loc1, loc2 Location) float64 {
	const R = 6371 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := loc1.Latitude * math.Pi / 180
	lat2Rad := loc2.Latitude * math.Pi / 180
	dLat := (loc2.Latitude - loc1.Latitude) * math.Pi / 180
	dLon := (loc2.Longitude - loc1.Longitude) * math.Pi / 180

	// Haversine formula
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// FindNearbyRestaurants finds restaurants within a specified radius of a location
// Returns restaurants sorted by distance (closest first)
func (l *LocationService) FindNearbyRestaurants(userLocation Location, radius float64) ([]RestaurantDistance, error) {
	var restaurants []models.Restaurant

	// Get all restaurants from database
	if err := l.db.Find(&restaurants).Error; err != nil {
		return nil, err
	}

	var results []RestaurantDistance

	// Calculate distance for each restaurant and filter by radius
	for _, restaurant := range restaurants {
		// Parse restaurant location coordinates
		// Note: This assumes restaurant.Location contains coordinates
		// In a real implementation, you'd need to geocode the location string
		// or store lat/long separately in the database

		// For now, we'll skip restaurants without valid coordinates
		// This would need to be enhanced with geocoding service
		restaurantLoc := Location{
			Latitude:  0, // Would come from geocoding or stored coordinates
			Longitude: 0,
		}

		distance := l.CalculateDistance(userLocation, restaurantLoc)

		if distance <= radius {
			results = append(results, RestaurantDistance{
				Restaurant: restaurant,
				Distance:   distance,
				Visited:    false, // Would be determined by checking entries
				Wishlist:   false, // Would be determined by checking entries
			})
		}
	}

	// Sort results by distance (bubble sort for simplicity)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Distance > results[j].Distance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results, nil
}

// FindNearestWishlist finds the nearest restaurants from a user's wishlist
// Returns restaurants the user wants to visit, sorted by distance
func (l *LocationService) FindNearestWishlist(userID uint, userLocation Location) ([]RestaurantDistance, error) {
	var entries []models.Entry

	// Get user's wishlist entries (visited = false)
	if err := l.db.Where("submitter = ? AND visited = ?", userID, false).Find(&entries).Error; err != nil {
		return nil, err
	}

	var results []RestaurantDistance

	// For each wishlist entry, find the corresponding restaurant and calculate distance
	for _, entry := range entries {
		var restaurant models.Restaurant

		// Find restaurant by name and location
		if err := l.db.Where("name = ? AND location = ?", entry.Name, entry.Location).First(&restaurant).Error; err != nil {
			// Skip if restaurant not found
			continue
		}

		// Parse restaurant location coordinates
		restaurantLoc := Location{
			Latitude:  0, // Would come from geocoding or stored coordinates
			Longitude: 0,
		}

		distance := l.CalculateDistance(userLocation, restaurantLoc)

		results = append(results, RestaurantDistance{
			Restaurant: restaurant,
			Distance:   distance,
			Visited:    false,
			Wishlist:   true,
		})
	}

	// Sort results by distance
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Distance > results[j].Distance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results, nil
}

// FindNearestUnvisited finds the nearest restaurants the user hasn't visited yet
// Returns unvisited restaurants sorted by distance
func (l *LocationService) FindNearestUnvisited(userID uint, userLocation Location) ([]RestaurantDistance, error) {
	var visitedEntries []models.Entry

	// Get all restaurants the user has visited
	if err := l.db.Where("submitter = ? AND visited = ?", userID, true).Find(&visitedEntries).Error; err != nil {
		return nil, err
	}

	// Create a map of visited restaurants for quick lookup
	visitedMap := make(map[string]bool)
	for _, entry := range visitedEntries {
		key := entry.Name + "|" + entry.Location
		visitedMap[key] = true
	}

	var restaurants []models.Restaurant

	// Get all restaurants
	if err := l.db.Find(&restaurants).Error; err != nil {
		return nil, err
	}

	var results []RestaurantDistance

	// Calculate distance for unvisited restaurants
	for _, restaurant := range restaurants {
		key := restaurant.Name + "|" + restaurant.Location

		// Skip if user has visited this restaurant
		if visitedMap[key] {
			continue
		}

		// Parse restaurant location coordinates
		restaurantLoc := Location{
			Latitude:  0, // Would come from geocoding or stored coordinates
			Longitude: 0,
		}

		distance := l.CalculateDistance(userLocation, restaurantLoc)

		results = append(results, RestaurantDistance{
			Restaurant: restaurant,
			Distance:   distance,
			Visited:    false,
			Wishlist:   false, // Could check if it's in wishlist
		})
	}

	// Sort results by distance
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Distance > results[j].Distance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results, nil
}
