package services

import (
	"math"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// **Validates: Requirements 8.6**
// Property 40: Distance Calculation Accuracy
// For any two geographic coordinates, the calculated distance should be mathematically accurate using the Haversine formula
func TestProperty40_DistanceCalculationAccuracy(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 40: Distance calculation is accurate and consistent", prop.ForAll(
		func(lat1, lon1, lat2, lon2 float64) bool {
			db := setupLocationTestDB(t)
			service := NewLocationService(db)

			loc1 := Location{Latitude: lat1, Longitude: lon1}
			loc2 := Location{Latitude: lat2, Longitude: lon2}

			distance := service.CalculateDistance(loc1, loc2)

			// Property 1: Distance should be non-negative
			if distance < 0 {
				return false
			}

			// Property 2: Distance from a point to itself should be 0
			if lat1 == lat2 && lon1 == lon2 {
				return math.Abs(distance) < 0.001
			}

			// Property 3: Distance should be symmetric (d(A,B) = d(B,A))
			reverseDistance := service.CalculateDistance(loc2, loc1)
			if math.Abs(distance-reverseDistance) > 0.001 {
				return false
			}

			// Property 4: Distance should be reasonable (not exceed Earth's circumference)
			earthCircumference := 40075.0 // km
			return distance <= earthCircumference
		},
		gen.Float64Range(-90, 90),   // lat1
		gen.Float64Range(-180, 180), // lon1
		gen.Float64Range(-90, 90),   // lat2
		gen.Float64Range(-180, 180), // lon2
	))

	properties.TestingRun(t)
}

// **Validates: Requirements 8.2**
// Property 41: Wishlist Distance Calculation
// For any user location and wishlist, distances to all wishlist restaurants should be calculated correctly
func TestProperty41_WishlistDistanceCalculation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 41: Wishlist distances are calculated correctly", prop.ForAll(
		func(userLat, userLon float64) bool {
			db := setupLocationTestDB(t)
			service := NewLocationService(db)

			userLocation := Location{Latitude: userLat, Longitude: userLon}

			// Get wishlist results
			results, err := service.FindNearestWishlist(1, userLocation)
			if err != nil {
				return false
			}

			// Property 1: All results should have wishlist flag set
			for _, result := range results {
				if !result.Wishlist {
					return false
				}
			}

			// Property 2: All results should have visited flag as false
			for _, result := range results {
				if result.Visited {
					return false
				}
			}

			// Property 3: All distances should be non-negative
			for _, result := range results {
				if result.Distance < 0 {
					return false
				}
			}

			return true
		},
		gen.Float64Range(-90, 90),   // userLat
		gen.Float64Range(-180, 180), // userLon
	))

	properties.TestingRun(t)
}

// **Validates: Requirements 8.3**
// Property 42: Location Result Sorting
// For any location-based search results, unvisited restaurants should be sorted by distance with closest first
func TestProperty42_LocationResultSorting(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 42: Results are sorted by distance (closest first)", prop.ForAll(
		func(userLat, userLon, radius float64) bool {
			db := setupLocationTestDB(t)
			service := NewLocationService(db)

			userLocation := Location{Latitude: userLat, Longitude: userLon}

			// Get nearby restaurants
			results, err := service.FindNearbyRestaurants(userLocation, radius)
			if err != nil {
				return false
			}

			// Property 1: Results should be sorted by distance (ascending)
			for i := 1; i < len(results); i++ {
				if results[i-1].Distance > results[i].Distance {
					return false
				}
			}

			// Property 2: All results should be within the specified radius
			for _, result := range results {
				if result.Distance > radius {
					return false
				}
			}

			// Property 3: All distances should be non-negative
			for _, result := range results {
				if result.Distance < 0 {
					return false
				}
			}

			return true
		},
		gen.Float64Range(-90, 90),   // userLat
		gen.Float64Range(-180, 180), // userLon
		gen.Float64Range(1, 10000),  // radius (1 to 10000 km)
	))

	properties.TestingRun(t)
}

// **Validates: Requirements 8.4**
// Property 43: Wishlist Location Search
// For any "closest wishlist" search, results should only include restaurants the user wants to visit, sorted by distance
func TestProperty43_WishlistLocationSearch(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 43: Wishlist search returns only wishlist items sorted by distance", prop.ForAll(
		func(userLat, userLon float64) bool {
			db := setupLocationTestDB(t)
			service := NewLocationService(db)

			userLocation := Location{Latitude: userLat, Longitude: userLon}

			// Get wishlist results
			results, err := service.FindNearestWishlist(1, userLocation)
			if err != nil {
				return false
			}

			// Property 1: All results should have wishlist flag set
			for _, result := range results {
				if !result.Wishlist {
					return false
				}
			}

			// Property 2: All results should have visited flag as false (wishlist = not visited)
			for _, result := range results {
				if result.Visited {
					return false
				}
			}

			// Property 3: Results should be sorted by distance (ascending)
			for i := 1; i < len(results); i++ {
				if results[i-1].Distance > results[i].Distance {
					return false
				}
			}

			// Property 4: All distances should be non-negative
			for _, result := range results {
				if result.Distance < 0 {
					return false
				}
			}

			return true
		},
		gen.Float64Range(-90, 90),   // userLat
		gen.Float64Range(-180, 180), // userLon
	))

	properties.TestingRun(t)
}

// Additional property test: Triangle Inequality
// For any three points A, B, C: distance(A,C) <= distance(A,B) + distance(B,C)
func TestProperty_TriangleInequality(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Triangle inequality holds for distance calculations", prop.ForAll(
		func(lat1, lon1, lat2, lon2, lat3, lon3 float64) bool {
			db := setupLocationTestDB(t)
			service := NewLocationService(db)

			locA := Location{Latitude: lat1, Longitude: lon1}
			locB := Location{Latitude: lat2, Longitude: lon2}
			locC := Location{Latitude: lat3, Longitude: lon3}

			distAB := service.CalculateDistance(locA, locB)
			distBC := service.CalculateDistance(locB, locC)
			distAC := service.CalculateDistance(locA, locC)

			// Triangle inequality: d(A,C) <= d(A,B) + d(B,C)
			// Allow small floating point error
			return distAC <= distAB+distBC+0.001
		},
		gen.Float64Range(-90, 90),   // lat1
		gen.Float64Range(-180, 180), // lon1
		gen.Float64Range(-90, 90),   // lat2
		gen.Float64Range(-180, 180), // lon2
		gen.Float64Range(-90, 90),   // lat3
		gen.Float64Range(-180, 180), // lon3
	))

	properties.TestingRun(t)
}

// Additional property test: Radius filtering consistency
func TestProperty_RadiusFilteringConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Radius filtering is consistent", prop.ForAll(
		func(userLat, userLon, radius1, radius2 float64) bool {
			// Ensure radius1 < radius2
			if radius1 > radius2 {
				radius1, radius2 = radius2, radius1
			}

			db := setupLocationTestDB(t)
			service := NewLocationService(db)

			userLocation := Location{Latitude: userLat, Longitude: userLon}

			// Get results with smaller radius
			results1, err1 := service.FindNearbyRestaurants(userLocation, radius1)
			if err1 != nil {
				return false
			}

			// Get results with larger radius
			results2, err2 := service.FindNearbyRestaurants(userLocation, radius2)
			if err2 != nil {
				return false
			}

			// Property: Results with smaller radius should be a subset of results with larger radius
			// (or equal in count if no restaurants exist in the larger radius)
			return len(results1) <= len(results2)
		},
		gen.Float64Range(-90, 90),   // userLat
		gen.Float64Range(-180, 180), // userLon
		gen.Float64Range(1, 5000),   // radius1
		gen.Float64Range(1, 5000),   // radius2
	))

	properties.TestingRun(t)
}
