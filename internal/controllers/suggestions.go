package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/pkg/database"
)

// SuggestionResponse represents a single suggestion item
type SuggestionResponse struct {
	Value string `json:"value"`
	Count int    `json:"count,omitempty"` // How many times this value appears
}

// SuggestionsResponse represents the response for suggestions API
type SuggestionsResponse struct {
	Suggestions []SuggestionResponse `json:"suggestions"`
	Field       string               `json:"field"`
	Query       string               `json:"query"`
}

// GetCuisineSuggestions handles GET /api/suggestions/cuisine requests
// Returns a list of cuisine types based on existing entries, optionally filtered by query
func GetCuisineSuggestions(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	limit := 10 // Default limit for suggestions

	// Create timeout context for database operations
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Build the SQL query to get distinct cuisine values with counts
	sqlQuery := `
		SELECT cuisine, COUNT(*) as count 
		FROM entries 
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}

	// Add search filter if query provided
	if query != "" {
		sqlQuery += " AND LOWER(cuisine) LIKE LOWER(?)"
		args = append(args, "%"+query+"%")
	}

	sqlQuery += " GROUP BY LOWER(cuisine) ORDER BY count DESC, cuisine ASC LIMIT ?"
	args = append(args, limit)

	// Execute the query
	rows, err := database.WithContext(ctx).Raw(sqlQuery, args...).Rows()
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}
	defer rows.Close()

	var suggestions []SuggestionResponse
	for rows.Next() {
		var cuisine string
		var count int
		if err := rows.Scan(&cuisine, &count); err != nil {
			continue // Skip invalid rows
		}
		suggestions = append(suggestions, SuggestionResponse{
			Value: cuisine,
			Count: count,
		})
	}

	response := SuggestionsResponse{
		Suggestions: suggestions,
		Field:       "cuisine",
		Query:       query,
	}

	Response.Success(c, response, "")
}

// GetRestaurantNameSuggestions handles GET /api/suggestions/restaurant-names requests
// Returns a list of restaurant names based on existing entries
func GetRestaurantNameSuggestions(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	limit := 10

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	sqlQuery := `
		SELECT name, COUNT(*) as count 
		FROM entries 
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}

	if query != "" {
		sqlQuery += " AND LOWER(name) LIKE LOWER(?)"
		args = append(args, "%"+query+"%")
	}

	sqlQuery += " GROUP BY LOWER(name) ORDER BY count DESC, name ASC LIMIT ?"
	args = append(args, limit)

	rows, err := database.WithContext(ctx).Raw(sqlQuery, args...).Rows()
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}
	defer rows.Close()

	var suggestions []SuggestionResponse
	for rows.Next() {
		var name string
		var count int
		if err := rows.Scan(&name, &count); err != nil {
			continue
		}
		suggestions = append(suggestions, SuggestionResponse{
			Value: name,
			Count: count,
		})
	}

	response := SuggestionsResponse{
		Suggestions: suggestions,
		Field:       "restaurant-name",
		Query:       query,
	}

	Response.Success(c, response, "")
}

// GetMealServiceSuggestions handles GET /api/suggestions/meal-service requests
// Returns a list of meal service types based on existing entries
func GetMealServiceSuggestions(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	limit := 10

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	sqlQuery := `
		SELECT meal_service, COUNT(*) as count 
		FROM entries 
		WHERE deleted_at IS NULL AND meal_service != ''
	`
	args := []interface{}{}

	if query != "" {
		sqlQuery += " AND LOWER(meal_service) LIKE LOWER(?)"
		args = append(args, "%"+query+"%")
	}

	sqlQuery += " GROUP BY LOWER(meal_service) ORDER BY count DESC, meal_service ASC LIMIT ?"
	args = append(args, limit)

	rows, err := database.WithContext(ctx).Raw(sqlQuery, args...).Rows()
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}
	defer rows.Close()

	var suggestions []SuggestionResponse
	for rows.Next() {
		var mealService string
		var count int
		if err := rows.Scan(&mealService, &count); err != nil {
			continue
		}
		suggestions = append(suggestions, SuggestionResponse{
			Value: mealService,
			Count: count,
		})
	}

	response := SuggestionsResponse{
		Suggestions: suggestions,
		Field:       "meal-service",
		Query:       query,
	}

	Response.Success(c, response, "")
}

// GetDishSuggestions handles GET /api/suggestions/dishes requests
// Returns a list of dishes based on existing entries (from ordered, ordernext, donotorder fields)
func GetDishSuggestions(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	fieldType := c.Query("type") // "ordered", "ordernext", or "donotorder"
	limit := 15

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Determine which field to search based on type
	var field string
	switch fieldType {
	case "ordernext":
		field = "order_next"
	case "donotorder":
		field = "do_not_order"
	default:
		field = "ordered"
	}

	// Get all non-empty values from the specified field
	sqlQuery := `
		SELECT ` + field + ` 
		FROM entries 
		WHERE deleted_at IS NULL AND ` + field + ` != ''
	`

	rows, err := database.WithContext(ctx).Raw(sqlQuery).Rows()
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}
	defer rows.Close()

	// Parse comma-separated dishes and count occurrences
	dishCounts := make(map[string]int)
	for rows.Next() {
		var dishList string
		if err := rows.Scan(&dishList); err != nil {
			continue
		}

		// Split by comma and clean up each dish
		dishes := strings.Split(dishList, ",")
		for _, dish := range dishes {
			dish = strings.TrimSpace(dish)
			if dish != "" {
				// Convert to lowercase for case-insensitive matching
				lowerDish := strings.ToLower(dish)
				if query == "" || strings.Contains(lowerDish, strings.ToLower(query)) {
					dishCounts[dish]++
				}
			}
		}
	}

	// Convert map to sorted slice
	var suggestions []SuggestionResponse
	for dish, count := range dishCounts {
		suggestions = append(suggestions, SuggestionResponse{
			Value: dish,
			Count: count,
		})
	}

	// Sort by count (descending) then by name
	for i := 0; i < len(suggestions)-1; i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[i].Count < suggestions[j].Count ||
				(suggestions[i].Count == suggestions[j].Count && suggestions[i].Value > suggestions[j].Value) {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	// Limit results
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	response := SuggestionsResponse{
		Suggestions: suggestions,
		Field:       "dishes",
		Query:       query,
	}

	Response.Success(c, response, "")
}

// GetLocationSuggestions handles GET /api/suggestions/locations requests
// Returns a list of location values based on existing entries
func GetLocationSuggestions(c *gin.Context) {
	query := strings.TrimSpace(c.Query("q"))
	limit := 10

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	sqlQuery := `
		SELECT location, COUNT(*) as count 
		FROM entries 
		WHERE deleted_at IS NULL
	`
	args := []interface{}{}

	if query != "" {
		sqlQuery += " AND LOWER(location) LIKE LOWER(?)"
		args = append(args, "%"+query+"%")
	}

	sqlQuery += " GROUP BY LOWER(location) ORDER BY count DESC, location ASC LIMIT ?"
	args = append(args, limit)

	rows, err := database.WithContext(ctx).Raw(sqlQuery, args...).Rows()
	if err != nil {
		Response.DatabaseError(c, "query", err)
		return
	}
	defer rows.Close()

	var suggestions []SuggestionResponse
	for rows.Next() {
		var location string
		var count int
		if err := rows.Scan(&location, &count); err != nil {
			continue
		}
		suggestions = append(suggestions, SuggestionResponse{
			Value: location,
			Count: count,
		})
	}

	response := SuggestionsResponse{
		Suggestions: suggestions,
		Field:       "location",
		Query:       query,
	}

	Response.Success(c, response, "")
}
