package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/helpers"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/search"
	"github.com/toozej/dinnerclub/internal/validation"
	"github.com/toozej/dinnerclub/pkg/database"
	"github.com/toozej/dinnerclub/pkg/pagination"
	"gorm.io/gorm"
)

// SearchController handles search-related operations
type SearchController struct {
	searchService *search.SearchService
}

// NewSearchController creates a new search controller instance with an initialized
// search service for handling search operations.
func NewSearchController() *SearchController {
	return &SearchController{
		searchService: search.NewSearchService(database.DB),
	}
}

// GetSearch handles GET /search requests and performs comprehensive search across
// entries and restaurants with pagination and support for both HTML and JSON responses.
func (sc *SearchController) GetSearch(c *gin.Context) {
	query := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	// Handle empty queries specially - return empty results without validation
	if strings.TrimSpace(query) == "" {
		// Check if request wants JSON response
		if c.GetHeader("Accept") == "application/json" || c.Query("format") == "json" {
			c.JSON(http.StatusOK, gin.H{
				"query":       query,
				"entries":     []interface{}{},
				"restaurants": []interface{}{},
				"total":       0,
				"pagination":  nil,
			})
			return
		}

		// Render HTML template with empty results
		templateData := gin.H{
			"query":        query,
			"entries":      []interface{}{},
			"restaurants":  []interface{}{},
			"total":        0,
			"pagination":   nil,
			"is_logged_in": c.MustGet("is_logged_in").(bool),
			"citycode":     c.MustGet("citycode").(string),
			"messages":     []interface{}{},
			"context":      c,
		}
		templateData = helpers.InitializeTemplateHelpers(templateData)
		c.HTML(http.StatusOK, "search.html", templateData)
		return
	}

	// Validate search parameters using the new validation helper
	validationResult := validation.ValidationHelperInstance.ValidateSearchParameters(query, pageStr, limitStr, "", "")
	if !validationResult.IsValid {
		validation.ValidationHelperInstance.HandleValidationErrors(c, validationResult, "Invalid search parameters")
		return
	}

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// Perform search
	result, err := sc.searchService.Search(query)
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}

	// Apply pagination to results
	totalResults := len(result.Entries) + len(result.Restaurants)
	p, err := pagination.Paginate(pageStr, totalResults, limit)
	if err != nil {
		Response.BadRequest(c, "Invalid pagination parameters", err)
		return
	}

	// Calculate pagination offsets for entries and restaurants
	offset := (page - 1) * limit
	endOffset := offset + limit

	var paginatedEntries []interface{}
	var paginatedRestaurants []interface{}

	// Convert entries to interface{} and apply pagination
	allResults := make([]interface{}, 0, totalResults)
	for _, entry := range result.Entries {
		allResults = append(allResults, entry)
	}
	for _, restaurant := range result.Restaurants {
		allResults = append(allResults, restaurant)
	}

	// Apply pagination to combined results
	if offset < len(allResults) {
		end := endOffset
		if end > len(allResults) {
			end = len(allResults)
		}
		paginatedResults := allResults[offset:end]

		// Separate back into entries and restaurants
		for _, item := range paginatedResults {
			switch v := item.(type) {
			case models.Entry:
				paginatedEntries = append(paginatedEntries, v)
			case models.Restaurant:
				paginatedRestaurants = append(paginatedRestaurants, v)
			}
		}
	}

	// Check if request wants JSON response
	if c.GetHeader("Accept") == "application/json" || c.Query("format") == "json" {
		c.JSON(http.StatusOK, gin.H{
			"query":       query,
			"entries":     paginatedEntries,
			"restaurants": paginatedRestaurants,
			"total":       totalResults,
			"pagination":  p,
		})
		return
	}

	// Render HTML template with helpers
	templateData := gin.H{
		"query":        query,
		"entries":      paginatedEntries,
		"restaurants":  paginatedRestaurants,
		"total":        totalResults,
		"pagination":   p,
		"is_logged_in": c.MustGet("is_logged_in").(bool),
		"citycode":     c.MustGet("citycode").(string),
		"messages":     []interface{}{}, // Empty messages for now
		"context":      c,
	}
	templateData = helpers.InitializeTemplateHelpers(templateData)
	c.HTML(http.StatusOK, "search/results.html", templateData)
}

// GetSearchEntries handles GET /search/entries requests for entry-specific search
// with enhanced filtering, sorting, and pagination capabilities.
func (sc *SearchController) GetSearchEntries(c *gin.Context) {
	query := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	sortBy := c.DefaultQuery("sort", "created_at")
	sortOrder := c.DefaultQuery("order", "desc")

	// Validate search parameters using the new validation helper
	validationResult := validation.ValidationHelperInstance.ValidateSearchParameters(query, pageStr, limitStr, sortBy, sortOrder)
	if !validationResult.IsValid {
		validation.ValidationHelperInstance.HandleValidationErrors(c, validationResult, "Invalid search parameters")
		return
	}

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// Parse search query
	searchQuery := sc.searchService.ParseQuery(query)

	// Search entries with pagination and sorting
	entries, err := sc.searchEntriesWithPagination(searchQuery, page, limit, sortBy, sortOrder)
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}

	// Get total count for pagination
	totalEntries, err := sc.getEntriesCount(searchQuery)
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}

	p, err := pagination.Paginate(pageStr, totalEntries, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Check if request wants JSON response
	if c.GetHeader("Accept") == "application/json" || c.Query("format") == "json" {
		c.JSON(http.StatusOK, gin.H{
			"query":      query,
			"entries":    entries,
			"total":      totalEntries,
			"pagination": p,
			"sort":       sortBy,
			"order":      sortOrder,
		})
		return
	}

	// Render HTML template with helpers
	templateData := gin.H{
		"query":        query,
		"entries":      entries,
		"total":        totalEntries,
		"pagination":   p,
		"sort":         sortBy,
		"order":        sortOrder,
		"is_logged_in": c.MustGet("is_logged_in").(bool),
		"citycode":     c.MustGet("citycode").(string),
		"messages":     []interface{}{}, // Empty messages for now
		"context":      c,
	}
	templateData = helpers.InitializeTemplateHelpers(templateData)
	c.HTML(http.StatusOK, "search/entry_results.html", templateData)
}

// GetSearchRestaurants handles GET /search/restaurants requests for restaurant-specific
// search with enhanced filtering, sorting, and pagination capabilities.
func (sc *SearchController) GetSearchRestaurants(c *gin.Context) {
	query := c.Query("q")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	sortBy := c.DefaultQuery("sort", "name")
	sortOrder := c.DefaultQuery("order", "asc")

	// Validate search parameters using the new validation helper
	validationResult := validation.ValidationHelperInstance.ValidateSearchParameters(query, pageStr, limitStr, sortBy, sortOrder)
	if !validationResult.IsValid {
		validation.ValidationHelperInstance.HandleValidationErrors(c, validationResult, "Invalid search parameters")
		return
	}

	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)

	// Parse search query
	searchQuery := sc.searchService.ParseQuery(query)

	// Search restaurants with pagination and sorting
	restaurants, err := sc.searchRestaurantsWithPagination(searchQuery, page, limit, sortBy, sortOrder)
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}

	// Get total count for pagination
	totalRestaurants, err := sc.getRestaurantsCount(searchQuery)
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}

	p, err := pagination.Paginate(pageStr, totalRestaurants, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
		return
	}

	// Check if request wants JSON response
	if c.GetHeader("Accept") == "application/json" || c.Query("format") == "json" {
		c.JSON(http.StatusOK, gin.H{
			"query":       query,
			"restaurants": restaurants,
			"total":       totalRestaurants,
			"pagination":  p,
			"sort":        sortBy,
			"order":       sortOrder,
		})
		return
	}

	// Render HTML template with helpers
	templateData := gin.H{
		"query":        query,
		"restaurants":  restaurants,
		"total":        totalRestaurants,
		"pagination":   p,
		"sort":         sortBy,
		"order":        sortOrder,
		"is_logged_in": c.MustGet("is_logged_in").(bool),
		"citycode":     c.MustGet("citycode").(string),
		"messages":     []interface{}{}, // Empty messages for now
		"context":      c,
	}
	templateData = helpers.InitializeTemplateHelpers(templateData)
	c.HTML(http.StatusOK, "search/restaurant_results.html", templateData)
}

// GetSuggestions handles GET /search/suggestions requests and returns search
// suggestions for autocomplete functionality based on partial query input.
func (sc *SearchController) GetSuggestions(c *gin.Context) {
	partial := c.Query("q")
	if len(partial) < 2 {
		c.JSON(http.StatusOK, gin.H{"suggestions": []string{}})
		return
	}

	suggestions, err := sc.searchService.GetSearchSuggestions(partial)
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestions": suggestions})
}

// Helper method to search entries with pagination and sorting
func (sc *SearchController) searchEntriesWithPagination(query search.SearchQuery, page, limit int, sortBy, sortOrder string) ([]models.Entry, error) {
	var entries []models.Entry

	db := database.DB.Model(&models.Entry{})

	// Apply search filters
	db = sc.applyEntryFilters(db, query)

	// Apply sorting
	orderClause := sortBy + " " + sortOrder
	db = db.Order(orderClause)

	// Apply pagination
	offset := (page - 1) * limit
	db = db.Limit(limit).Offset(offset)

	err := db.Find(&entries).Error
	return entries, err
}

// Helper method to search restaurants with pagination and sorting
func (sc *SearchController) searchRestaurantsWithPagination(query search.SearchQuery, page, limit int, sortBy, sortOrder string) ([]models.Restaurant, error) {
	var restaurants []models.Restaurant

	db := database.DB.Model(&models.Restaurant{})

	// Apply search filters
	db = sc.applyRestaurantFilters(db, query)

	// Handle entry_count sorting specially
	if sortBy == "entry_count" {
		// Join with entries table and count
		db = db.Select("restaurants.*, COUNT(entries.id) as entry_count").
			Joins("LEFT JOIN entries ON entries.name = restaurants.name").
			Group("restaurants.id").
			Order("entry_count " + sortOrder)
	} else {
		// Apply regular sorting
		orderClause := sortBy + " " + sortOrder
		db = db.Order(orderClause)
	}

	// Apply pagination
	offset := (page - 1) * limit
	db = db.Limit(limit).Offset(offset)

	err := db.Find(&restaurants).Error
	return restaurants, err
}

// Helper method to get total count of entries matching search query
func (sc *SearchController) getEntriesCount(query search.SearchQuery) (int, error) {
	var count int64

	db := database.DB.Model(&models.Entry{})
	db = sc.applyEntryFilters(db, query)

	err := db.Count(&count).Error
	return int(count), err
}

// Helper method to get total count of restaurants matching search query
func (sc *SearchController) getRestaurantsCount(query search.SearchQuery) (int, error) {
	var count int64

	db := database.DB.Model(&models.Restaurant{})
	db = sc.applyRestaurantFilters(db, query)

	err := db.Count(&count).Error
	return int(count), err
}

// Helper method to apply entry search filters
func (sc *SearchController) applyEntryFilters(db *gorm.DB, query search.SearchQuery) *gorm.DB {
	// Apply text search on name, cuisine, and comments
	if query.Text != "" {
		searchText := "%" + query.Text + "%"
		db = db.Where(
			"LOWER(name) LIKE ? OR LOWER(cuisine) LIKE ? OR LOWER(comments) LIKE ? OR LOWER(ordered) LIKE ?",
			searchText, searchText, searchText, searchText,
		)
	}

	// Apply author filter
	if query.Author != "" {
		db = db.Where("LOWER(submitter) = ?", strings.ToLower(query.Author))
	}

	// Apply visited filter
	if query.Visited != nil {
		db = db.Where("visited = ?", *query.Visited)
	}

	// Apply location filter
	if query.Location != "" {
		db = db.Where("location = ?", query.Location)
	}

	// Apply cuisine filter
	if query.Cuisine != "" {
		db = db.Where("LOWER(cuisine) LIKE ?", "%"+strings.ToLower(query.Cuisine)+"%")
	}

	// Apply rating filter (search for entries with any rating >= specified rating)
	if query.Rating > 0 {
		db = db.Where(
			"food_rating >= ? OR ambience_rating >= ? OR value_rating >= ?",
			query.Rating, query.Rating, query.Rating,
		)
	}

	return db
}

// Helper method to apply restaurant search filters
func (sc *SearchController) applyRestaurantFilters(db *gorm.DB, query search.SearchQuery) *gorm.DB {
	// Apply text search on name and cuisine
	if query.Text != "" {
		searchText := "%" + query.Text + "%"
		db = db.Where(
			"LOWER(restaurants.name) LIKE ? OR LOWER(restaurants.cuisine) LIKE ?",
			searchText, searchText,
		)
	}

	// Apply location filter
	if query.Location != "" {
		db = db.Where("restaurants.location = ?", query.Location)
	}

	// Apply cuisine filter
	if query.Cuisine != "" {
		db = db.Where("LOWER(restaurants.cuisine) LIKE ?", "%"+strings.ToLower(query.Cuisine)+"%")
	}

	return db
}
