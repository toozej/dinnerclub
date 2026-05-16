package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/services"
	"gorm.io/gorm"
)

// LocationController handles location-based search requests
type LocationController struct {
	db              *gorm.DB
	locationService *services.LocationService
}

// NewLocationController creates a new LocationController instance
func NewLocationController(db *gorm.DB) *LocationController {
	return &LocationController{
		db:              db,
		locationService: services.NewLocationService(db),
	}
}

// LocationError represents a location-specific error with user guidance
type LocationError struct {
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions"`
	ErrorCode   string   `json:"error_code"`
}

// getLocationErrorSuggestions provides user-friendly suggestions based on error type
func (lc *LocationController) getLocationErrorSuggestions(errorType string) []string {
	switch errorType {
	case "INVALID_COORDINATES":
		return []string{
			"Check that latitude is between -90 and 90",
			"Check that longitude is between -180 and 180",
			"Use decimal degrees format (e.g., 40.7128, -74.0060)",
			"Try using the 'Use My Current Location' button instead",
		}
	case "INVALID_RADIUS":
		return []string{
			"Enter a radius between 1 and 100 kilometers",
			"Try starting with 10km and expand if needed",
			"Consider that larger radius may return many results",
		}
	case "NO_RESULTS":
		return []string{
			"Try expanding your search radius",
			"Check if your location is correct",
			"Try a different search type (nearby, wishlist, unvisited)",
			"Search in a nearby city or popular area",
		}
	case "AUTH_REQUIRED":
		return []string{
			"Log in to search your personal wishlist",
			"Log in to find restaurants you haven't visited",
			"Use 'All Nearby Restaurants' for public search",
		}
	default:
		return []string{
			"Try refreshing the page",
			"Check your internet connection",
			"Try entering your location manually",
		}
	}
}

// ShowLocationSearch displays the location search page
func (lc *LocationController) ShowLocationSearch(c *gin.Context) {
	c.HTML(http.StatusOK, "search/location.html", gin.H{
		"context": c,
		"results": nil,
	})
}

// SearchNearby handles location-based search requests
func (lc *LocationController) SearchNearby(c *gin.Context) {
	// Parse form data
	latStr := c.PostForm("latitude")
	lonStr := c.PostForm("longitude")
	address := c.PostForm("address")
	radiusStr := c.PostForm("radius")
	searchType := c.PostForm("search-type")

	var lat, lon float64
	var err error

	// Handle address input vs coordinate input
	if address != "" && (latStr == "" || lonStr == "") {
		// TODO: Implement geocoding service to convert address to coordinates
		// For now, return an error asking for coordinates
		c.HTML(http.StatusBadRequest, "search/location.html", gin.H{
			"context": c,
			"error":   "Address geocoding is not yet implemented. Please use coordinates for now.",
			"results": nil,
		})
		return
	}

	// Validate coordinate inputs
	if latStr == "" || lonStr == "" {
		c.HTML(http.StatusBadRequest, "search/location.html", gin.H{
			"context": c,
			"error":   "Please provide both latitude and longitude coordinates",
			"results": nil,
		})
		return
	}

	lat, err = strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "search/location.html", gin.H{
			"context": c,
			"error":   "Invalid latitude value. Please enter a number between -90 and 90.",
			"results": nil,
		})
		return
	}

	lon, err = strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.HTML(http.StatusBadRequest, "search/location.html", gin.H{
			"context": c,
			"error":   "Invalid longitude value. Please enter a number between -180 and 180.",
			"results": nil,
		})
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 || radius > 100 {
		c.HTML(http.StatusBadRequest, "search/location.html", gin.H{
			"context": c,
			"error":   "Invalid radius value. Please enter a number between 1 and 100 km.",
			"results": nil,
		})
		return
	}

	// Validate latitude and longitude ranges
	if lat < -90 || lat > 90 {
		c.HTML(http.StatusBadRequest, "search/location.html", gin.H{
			"context": c,
			"error":   "Latitude must be between -90 and 90 degrees.",
			"results": nil,
		})
		return
	}

	if lon < -180 || lon > 180 {
		c.HTML(http.StatusBadRequest, "search/location.html", gin.H{
			"context": c,
			"error":   "Longitude must be between -180 and 180 degrees.",
			"results": nil,
		})
		return
	}

	userLocation := services.Location{
		Latitude:  lat,
		Longitude: lon,
	}

	var results []services.RestaurantDistance
	var searchErr error

	// Perform search based on type
	switch searchType {
	case "wishlist":
		// Get user ID from context (assuming authentication middleware sets this)
		userID, exists := c.Get("userID")
		if !exists {
			c.HTML(http.StatusUnauthorized, "search/location.html", gin.H{
				"context": c,
				"error":   "You must be logged in to search your wishlist.",
				"results": nil,
			})
			return
		}

		results, searchErr = lc.locationService.FindNearestWishlist(userID.(uint), userLocation)

	case "unvisited":
		// Get user ID from context
		userID, exists := c.Get("userID")
		if !exists {
			c.HTML(http.StatusUnauthorized, "search/location.html", gin.H{
				"context": c,
				"error":   "You must be logged in to search unvisited restaurants.",
				"results": nil,
			})
			return
		}

		results, searchErr = lc.locationService.FindNearestUnvisited(userID.(uint), userLocation)

	default: // "nearby"
		results, searchErr = lc.locationService.FindNearbyRestaurants(userLocation, radius)
	}

	if searchErr != nil {
		c.HTML(http.StatusInternalServerError, "search/location.html", gin.H{
			"context": c,
			"error":   "Error performing location search. Please try again.",
			"results": nil,
		})
		return
	}

	// Provide helpful message if no results found
	var searchMessage string
	if len(results) == 0 {
		switch searchType {
		case "wishlist":
			searchMessage = "No wishlist restaurants found within " + radiusStr + "km. Try expanding your search radius or adding more restaurants to your wishlist."
		case "unvisited":
			searchMessage = "No unvisited restaurants found within " + radiusStr + "km. Try expanding your search radius or exploring a different area."
		default:
			searchMessage = "No restaurants found within " + radiusStr + "km. Try expanding your search radius or checking a different location."
		}
	}

	// Render results
	c.HTML(http.StatusOK, "search/location.html", gin.H{
		"context":       c,
		"results":       results,
		"lat":           lat,
		"lon":           lon,
		"radius":        radius,
		"searchType":    searchType,
		"searchMessage": searchMessage,
	})
}

// GetNearbyRestaurantsAPI returns nearby restaurants as JSON (API endpoint)
func (lc *LocationController) GetNearbyRestaurantsAPI(c *gin.Context) {
	// Parse query parameters
	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")
	radiusStr := c.DefaultQuery("radius", "10")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude value"})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude value"})
		return
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil || radius <= 0 || radius > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid radius value"})
		return
	}

	userLocation := services.Location{
		Latitude:  lat,
		Longitude: lon,
	}

	results, searchErr := lc.locationService.FindNearbyRestaurants(userLocation, radius)
	if searchErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error performing location search"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}

// GetNearestWishlistAPI returns nearest wishlist restaurants as JSON (API endpoint)
func (lc *LocationController) GetNearestWishlistAPI(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	// Parse query parameters
	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude value"})
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude value"})
		return
	}

	userLocation := services.Location{
		Latitude:  lat,
		Longitude: lon,
	}

	results, searchErr := lc.locationService.FindNearestWishlist(userID.(uint), userLocation)
	if searchErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error performing wishlist search"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}
