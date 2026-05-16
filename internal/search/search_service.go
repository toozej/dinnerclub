// Package search provides advanced search functionality for the dinnerclub application.
// It includes query parsing, filtering capabilities, and search operations across
// entries and restaurants with support for complex search patterns.
package search

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/gorm"
)

// SearchService handles search operations
type SearchService struct {
	db *gorm.DB
}

// SearchQuery represents a parsed search query
type SearchQuery struct {
	Text     string
	Author   string
	Wishlist bool
	Visited  *bool
	Location string
	Cuisine  string
	Rating   int
}

// SearchResult represents search results
type SearchResult struct {
	Entries     []models.Entry
	Restaurants []models.Restaurant
	TotalCount  int
}

// NewSearchService creates a new search service instance with the provided
// database connection for performing search operations.
func NewSearchService(db *gorm.DB) *SearchService {
	return &SearchService{db: db}
}

// ParseQuery parses a search query string into structured SearchQuery with support
// for special keywords like wishlist, author filters, location, cuisine, and rating filters.
func (s *SearchService) ParseQuery(query string) SearchQuery {
	sq := SearchQuery{}

	// Convert to lowercase for case-insensitive matching
	lowerQuery := strings.ToLower(strings.TrimSpace(query))

	// Parse wishlist/want to go keywords
	wishlistPatterns := []string{
		"i want to go to", "want to go to", "planning to visit",
		"want to go", "want to try", "looking forward", "wishlist", "to try",
	}

	for _, pattern := range wishlistPatterns {
		if strings.Contains(lowerQuery, pattern) {
			sq.Wishlist = true
			visited := false
			sq.Visited = &visited
			// Remove the wishlist keywords from the text
			lowerQuery = strings.ReplaceAll(lowerQuery, pattern, "")
			lowerQuery = strings.TrimSpace(lowerQuery)
			break
		}
	}

	// Parse author/submitter patterns
	authorPatterns := []string{
		`by:(\w+)`,
		`author:(\w+)`,
		`submitter:(\w+)`,
		`from:(\w+)`,
	}

	for _, pattern := range authorPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(lowerQuery)
		if len(matches) > 1 {
			sq.Author = matches[1]
			// Remove the author pattern from the text
			lowerQuery = re.ReplaceAllString(lowerQuery, "")
			break
		}
	}

	// Parse location patterns
	locationPatterns := []string{
		`location:([a-zA-Z]+)`,
		`area:([a-zA-Z]+)`,
		`in:([a-zA-Z]+)`,
	}

	validLocations := map[string]bool{
		"n": true, "ne": true, "nw": true,
		"s": true, "se": true, "sw": true,
	}

	for _, pattern := range locationPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(lowerQuery)
		if len(matches) > 1 {
			location := strings.ToUpper(matches[1])
			if validLocations[strings.ToLower(location)] {
				sq.Location = location
				// Remove the location pattern from the text
				lowerQuery = re.ReplaceAllString(lowerQuery, "")
				break
			}
		}
	}

	// Parse cuisine patterns (use original query to preserve case)
	cuisinePatterns := []string{
		`(?i)cuisine:(\w+)`,
		`(?i)type:(\w+)`,
		`(?i)food:(\w+)`,
	}

	originalQuery := strings.TrimSpace(query)
	for _, pattern := range cuisinePatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(originalQuery)
		if len(matches) > 1 {
			sq.Cuisine = matches[1]
			// Remove the cuisine pattern from the text (case insensitive)
			lowerRe := regexp.MustCompile(strings.ToLower(pattern))
			lowerQuery = lowerRe.ReplaceAllString(lowerQuery, "")
			break
		}
	}

	// Parse rating patterns
	ratingPatterns := []string{
		`rating:([1-5])`,
		`rated:([1-5])`,
		`stars:([1-5])`,
	}

	for _, pattern := range ratingPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(lowerQuery)
		if len(matches) > 1 {
			if rating := matches[1]; rating >= "1" && rating <= "5" {
				sq.Rating = int(rating[0] - '0')
				// Remove the rating pattern from the text
				lowerQuery = re.ReplaceAllString(lowerQuery, "")
				break
			}
		}
	}

	// Clean up remaining text
	sq.Text = strings.TrimSpace(lowerQuery)

	return sq
}

// SearchEntries searches for entries based on the parsed query with support for
// text search, author filtering, visited status, location, cuisine, and rating filters.
func (s *SearchService) SearchEntries(query SearchQuery) ([]models.Entry, error) {
	entries := make([]models.Entry, 0)

	// If query is completely empty (no filters at all), return empty results
	if query.Text == "" && query.Author == "" && query.Visited == nil &&
		query.Location == "" && query.Cuisine == "" && query.Rating == 0 && !query.Wishlist {
		return entries, nil
	}

	db := s.db.Model(&models.Entry{})

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

	// Order by most recent first
	db = db.Order("created_at DESC")

	err := db.Find(&entries).Error
	return entries, err
}

// SearchRestaurants searches for restaurants based on the parsed query with support
// for text search on name and cuisine, location filtering, and cuisine filtering.
func (s *SearchService) SearchRestaurants(query SearchQuery) ([]models.Restaurant, error) {
	restaurants := make([]models.Restaurant, 0)

	// If query has entry-specific filters (author, wishlist, visited, rating), return empty results
	// since restaurants don't have these attributes
	if query.Author != "" || query.Wishlist || query.Visited != nil || query.Rating > 0 {
		return restaurants, nil
	}

	// If query is completely empty (no filters at all), return empty results
	if query.Text == "" && query.Location == "" && query.Cuisine == "" {
		return restaurants, nil
	}

	db := s.db.Model(&models.Restaurant{})

	// Apply text search on name and cuisine
	if query.Text != "" {
		searchText := "%" + query.Text + "%"
		db = db.Where(
			"LOWER(name) LIKE ? OR LOWER(cuisine) LIKE ?",
			searchText, searchText,
		)
	}

	// Apply location filter
	if query.Location != "" {
		db = db.Where("location = ?", query.Location)
	}

	// Apply cuisine filter
	if query.Cuisine != "" {
		db = db.Where("LOWER(cuisine) LIKE ?", "%"+strings.ToLower(query.Cuisine)+"%")
	}

	// Order by name
	db = db.Order("name ASC")

	err := db.Find(&restaurants).Error
	return restaurants, err
}

// Search performs a comprehensive search across entries and restaurants using
// the provided query string and returns combined results with total count.
func (s *SearchService) Search(queryString string) (SearchResult, error) {
	query := s.ParseQuery(queryString)

	// Initialize empty slices to ensure they're never nil
	entries := make([]models.Entry, 0)
	restaurants := make([]models.Restaurant, 0)

	searchEntries, err := s.SearchEntries(query)
	if err != nil {
		return SearchResult{
			Entries:     entries,
			Restaurants: restaurants,
			TotalCount:  0,
		}, fmt.Errorf("error searching entries: %w", err)
	}
	entries = searchEntries

	searchRestaurants, err := s.SearchRestaurants(query)
	if err != nil {
		return SearchResult{
			Entries:     entries,
			Restaurants: restaurants,
			TotalCount:  len(entries),
		}, fmt.Errorf("error searching restaurants: %w", err)
	}
	restaurants = searchRestaurants

	return SearchResult{
		Entries:     entries,
		Restaurants: restaurants,
		TotalCount:  len(entries) + len(restaurants),
	}, nil
}

// GetSearchSuggestions returns search suggestions based on partial query input
// including restaurant names and cuisine types for autocomplete functionality.
func (s *SearchService) GetSearchSuggestions(partial string) ([]string, error) {
	var suggestions []string

	if len(partial) < 2 {
		return suggestions, nil
	}

	// Get restaurant name suggestions
	var restaurantNames []string
	err := s.db.Model(&models.Restaurant{}).
		Select("DISTINCT name").
		Where("LOWER(name) LIKE ?", "%"+strings.ToLower(partial)+"%").
		Limit(5).
		Pluck("name", &restaurantNames).Error

	if err != nil {
		return suggestions, err
	}

	suggestions = append(suggestions, restaurantNames...)

	// Get cuisine suggestions
	var cuisines []string
	err = s.db.Model(&models.Entry{}).
		Select("DISTINCT cuisine").
		Where("LOWER(cuisine) LIKE ?", "%"+strings.ToLower(partial)+"%").
		Limit(3).
		Pluck("cuisine", &cuisines).Error

	if err != nil {
		return suggestions, err
	}

	for _, cuisine := range cuisines {
		suggestions = append(suggestions, "cuisine:"+cuisine)
	}

	return suggestions, nil
}
