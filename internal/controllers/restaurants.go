package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/aggregation"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
	"github.com/toozej/dinnerclub/pkg/helpers"
	"github.com/toozej/dinnerclub/pkg/pagination"
)

// RestaurantWithEntryCount holds restaurant data along with its entry count for listing
type RestaurantWithEntryCount struct {
	models.Restaurant
	EntryCount int64 `json:"entry_count"`
}

// GetRestaurants handles GET /restaurants requests and returns a paginated list
// of all restaurants sorted by entry count in ascending order.
func GetRestaurants(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")

	// Get all restaurants first
	var allRestaurants []models.Restaurant
	if err := database.DB.Find(&allRestaurants).Error; err != nil {
		Response.DatabaseError(c, "find", err)
		return
	}

	// Create aggregation service
	aggregationService := aggregation.NewRestaurantAggregationService(database.DB)

	// Calculate entry count for each restaurant and create RestaurantWithEntryCount slice
	var restaurantsWithCounts []RestaurantWithEntryCount
	for _, restaurant := range allRestaurants {
		entryCount, err := aggregationService.GetEntryCountByRestaurant(restaurant.Name)
		if err != nil {
			Response.DatabaseError(c, "query", err)
			return
		}

		restaurantsWithCounts = append(restaurantsWithCounts, RestaurantWithEntryCount{
			Restaurant: restaurant,
			EntryCount: entryCount,
		})
	}

	// Sort by entry count (ascending - restaurants with fewer entries first)
	sort.Slice(restaurantsWithCounts, func(i, j int) bool {
		return restaurantsWithCounts[i].EntryCount < restaurantsWithCounts[j].EntryCount
	})

	// Apply pagination
	const restaurantsPerPage = 10
	totalCount := len(restaurantsWithCounts)
	p, err := pagination.Paginate(pageStr, totalCount, restaurantsPerPage)
	if err != nil {
		Response.BadRequest(c, "Invalid pagination parameters", err)
		return
	}

	// Get the restaurants for the current page
	start := p.Offset
	end := start + restaurantsPerPage
	if end > totalCount {
		end = totalCount
	}

	var paginatedRestaurants []RestaurantWithEntryCount
	if start < totalCount {
		paginatedRestaurants = restaurantsWithCounts[start:end]
	}

	c.HTML(http.StatusOK, "restaurants/index.html",
		gin.H{"restaurants": paginatedRestaurants, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "p": p, "context": c})
}

// getPopularItems retrieves and processes popular ordered items for a restaurant
// by analyzing all entries and returning deduplicated items sorted by frequency.
func getPopularItems(restaurant models.Restaurant) ([]string, error) {
	var entry models.Entry
	var rawpopularitems []string
	if err := database.DB.Model(&entry).Select("Ordered").Where("Name = ? AND Ordered <> '' AND Ordered IS NOT NULL", restaurant.Name).Find(&rawpopularitems).Error; err != nil {
		return nil, err
	}

	var popularitems []string

	// remove leading and trailing whitespace on ordered items
	// then split comma-separated items into the popularitems slice
	for _, i := range rawpopularitems {
		trimmeditem := strings.TrimSpace(i)
		splititem := strings.Split(trimmeditem, ",")
		popularitems = append(popularitems, splititem...)
	}

	// sort and then deduplicate popularitems, most to least popular
	popularitems = helpers.RemoveDuplicates(helpers.SortByFrequency(popularitems))

	return popularitems, nil
}

// GetRestaurant handles GET /restaurants/:id requests and returns a single restaurant
// by ID with popular items and associated entries.
func GetRestaurant(c *gin.Context) {
	var restaurant models.Restaurant

	if err := database.DB.Where("id = ?", c.Param("id")).First(&restaurant).Error; err != nil {
		Response.NotFound(c, "Restaurant")
		return
	}

	popularitems, err := getPopularItems(restaurant)
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}

	var entries []models.Entry
	if err := database.DB.Model(&entries).Select("Submitter", "MealService", "FoodRating", "AmbienceRating", "ValueRating", "Comments").Where("Name = ?", restaurant.Name).Find(&entries).Error; err != nil {
		Response.DatabaseError(c, "find", err)
		return
	}

	c.HTML(http.StatusOK, "restaurants/restaurant.html",
		gin.H{"restaurant": restaurant, "popularitems": popularitems, "entries": entries, "is_logged_in": c.MustGet("is_logged_in").(bool), "citycode": c.MustGet("citycode").(string), "messages": Flashes(c), "context": c})
}

// GetRestaurantByName handles GET /restaurants/name/:name requests and returns
// comprehensive aggregated restaurant data including popular items and ratings.
func GetRestaurantByName(c *gin.Context) {
	restaurantName := c.Param("name")
	if restaurantName == "" {
		Response.BadRequest(c, "Restaurant name is required", nil)
		return
	}

	// Create aggregation service
	aggregationService := aggregation.NewRestaurantAggregationService(database.DB)

	// Get comprehensive restaurant data
	restaurantData, err := aggregationService.GetRestaurantAggregateData(restaurantName)
	if err != nil {
		Response.NotFound(c, "Restaurant")
		return
	}

	c.HTML(http.StatusOK, "restaurants/restaurant.html", gin.H{
		"restaurantData": restaurantData,
		"restaurant":     restaurantData.Restaurant,
		"popularitems":   restaurantData.PopularItems,
		"entries":        restaurantData.RecentEntries,
		"averageRatings": restaurantData.AverageRatings,
		"entryCount":     restaurantData.EntryCount,
		"is_logged_in":   c.MustGet("is_logged_in").(bool),
		"citycode":       c.MustGet("citycode").(string),
		"messages":       Flashes(c),
		"context":        c,
	})
}

// PostRestaurant handles POST /restaurants requests and creates new restaurant
// records if they don't already exist in the database.
func PostRestaurant(c *gin.Context) {
	// Validate input
	closed, _ := strconv.ParseBool(c.PostForm("closed"))
	restaurant := models.Restaurant{
		Name:     c.PostForm("name"),
		Location: c.PostForm("location"),
		Cuisine:  c.PostForm("cuisine"),
		Closed:   closed,
	}

	// Check if restaurant already exists
	var existingRestaurant models.Restaurant
	err := database.DB.Where("name = ?", restaurant.Name).First(&existingRestaurant).Error
	if err == nil {
		// Restaurant already exists, no need to create
		return
	}

	// Create the new restaurant
	if err := database.DB.Create(&restaurant).Error; err != nil {
		Response.DatabaseError(c, "create", err)
		return
	}

	// Flash a message that the new restaurant entry was saved successfully
	flashMessage(c, fmt.Sprintf("New restaurant '%s' saved successfully.", restaurant.Name))
}

// DeleteRestaurantIfUnused deletes a restaurant record if no entries reference it,
// maintaining data integrity by cleaning up unused restaurant records.
func DeleteRestaurantIfUnused(c *gin.Context, restaurantName string) {
	var entryCount int64
	// Check if any entries still reference this restaurant
	if err := database.DB.Model(&models.Entry{}).Where("name = ?", restaurantName).Count(&entryCount).Error; err != nil {
		Response.InternalError(c, "Failed to check restaurant references", err)
		return
	}

	// If no entries reference this restaurant, delete it
	if entryCount == 0 {
		var restaurant models.Restaurant
		if err := database.DB.Where("name = ?", restaurantName).First(&restaurant).Error; err != nil {
			// Restaurant not found, nothing to delete
			return
		}

		if err := database.DB.Delete(&restaurant).Error; err != nil {
			Response.InternalError(c, "Failed to delete restaurant", err)
			return
		}

		flashMessage(c, fmt.Sprintf("Restaurant '%s' deleted successfully.", restaurant.Name))
	}
}
