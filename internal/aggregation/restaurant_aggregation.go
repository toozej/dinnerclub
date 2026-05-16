// Package aggregation provides data aggregation services for the dinnerclub application.
// It includes restaurant data aggregation, popular item calculation, rating averages,
// and comprehensive restaurant analytics functionality.
package aggregation

import (
	"sort"
	"strings"

	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// RestaurantAggregateData contains comprehensive restaurant data with aggregated information
type RestaurantAggregateData struct {
	Restaurant     models.Restaurant `json:"restaurant"`
	EntryCount     int               `json:"entry_count"`
	PopularItems   []PopularItem     `json:"popular_items"`
	AverageRatings AverageRatings    `json:"average_ratings"`
	RecentEntries  []models.Entry    `json:"recent_entries"`
}

// PopularItem represents a frequently ordered item with its frequency count
type PopularItem struct {
	Name      string `json:"name"`
	Frequency int    `json:"frequency"`
}

// AverageRatings contains averaged ratings across all restaurant entries
type AverageRatings struct {
	Food     float64 `json:"food"`
	Ambience float64 `json:"ambience"`
	Value    float64 `json:"value"`
}

// RestaurantAggregationService provides methods for aggregating restaurant data
type RestaurantAggregationService struct {
	db *gorm.DB
}

// NewRestaurantAggregationService creates a new restaurant aggregation service
// with the provided database connection for performing data aggregation operations.
func NewRestaurantAggregationService(db *gorm.DB) *RestaurantAggregationService {
	return &RestaurantAggregationService{db: db}
}

// GetRestaurantAggregateData retrieves comprehensive aggregated data for a restaurant
// including popular items, average ratings, entry count, and recent entries.
func (s *RestaurantAggregationService) GetRestaurantAggregateData(restaurantName string) (*RestaurantAggregateData, error) {
	// Find the restaurant
	var restaurant models.Restaurant
	if err := s.db.Where("name = ?", restaurantName).First(&restaurant).Error; err != nil {
		return nil, err
	}

	// Get all entries for this restaurant
	entries, err := s.GetRecentEntriesByRestaurant(restaurantName, 0) // 0 means no limit
	if err != nil {
		return nil, err
	}

	// Calculate aggregated data
	popularItems := s.CalculatePopularItems(entries)
	averageRatings := s.CalculateAverageRatings(entries)

	// Get recent entries (limited to 10 for display)
	recentEntries, err := s.GetRecentEntriesByRestaurant(restaurantName, 10)
	if err != nil {
		return nil, err
	}

	return &RestaurantAggregateData{
		Restaurant:     restaurant,
		EntryCount:     len(entries),
		PopularItems:   popularItems,
		AverageRatings: averageRatings,
		RecentEntries:  recentEntries,
	}, nil
}

// CalculatePopularItems analyzes ordered items across entries and returns popular items
// with frequency counts, sorted by popularity and then alphabetically.
func (s *RestaurantAggregationService) CalculatePopularItems(entries []models.Entry) []PopularItem {
	itemFrequency := make(map[string]int)

	// Track original case for each item
	itemCases := make(map[string]string)

	// Process each entry's ordered items
	for _, entry := range entries {
		if entry.Ordered == "" {
			continue
		}

		// Split by various separators (comma, semicolon, pipe)
		separators := []string{",", ";", "|"}
		items := []string{entry.Ordered}

		for _, sep := range separators {
			var newItems []string
			for _, item := range items {
				newItems = append(newItems, strings.Split(item, sep)...)
			}
			items = newItems
		}

		for _, item := range items {
			item = strings.TrimSpace(item)
			if item != "" {
				// Use lowercase for counting but preserve original case
				lowerItem := strings.ToLower(item)
				if _, exists := itemCases[lowerItem]; !exists {
					itemCases[lowerItem] = item // Store first occurrence case
				}
				itemFrequency[lowerItem]++
			}
		}
	}

	// Convert map to slice and sort by frequency
	var popularItems []PopularItem
	for lowerItem, frequency := range itemFrequency {
		originalCase := itemCases[lowerItem]
		popularItems = append(popularItems, PopularItem{
			Name:      originalCase,
			Frequency: frequency,
		})
	}

	// Sort by frequency (descending) and then by name (ascending) for consistency
	sort.Slice(popularItems, func(i, j int) bool {
		if popularItems[i].Frequency == popularItems[j].Frequency {
			return popularItems[i].Name < popularItems[j].Name
		}
		return popularItems[i].Frequency > popularItems[j].Frequency
	})

	return popularItems
}

// CalculateAverageRatings computes average ratings across all restaurant entries
// for food, ambience, and value, excluding zero ratings from calculations.
func (s *RestaurantAggregationService) CalculateAverageRatings(entries []models.Entry) AverageRatings {
	var foodSum, ambienceSum, valueSum float64
	var foodCount, ambienceCount, valueCount int

	for _, entry := range entries {
		// Only include non-zero ratings in averages
		if entry.FoodRating > 0 {
			foodSum += float64(entry.FoodRating)
			foodCount++
		}
		if entry.AmbienceRating > 0 {
			ambienceSum += float64(entry.AmbienceRating)
			ambienceCount++
		}
		if entry.ValueRating > 0 {
			valueSum += float64(entry.ValueRating)
			valueCount++
		}
	}

	var avgFood, avgAmbience, avgValue float64
	if foodCount > 0 {
		avgFood = foodSum / float64(foodCount)
	}
	if ambienceCount > 0 {
		avgAmbience = ambienceSum / float64(ambienceCount)
	}
	if valueCount > 0 {
		avgValue = valueSum / float64(valueCount)
	}

	return AverageRatings{
		Food:     avgFood,
		Ambience: avgAmbience,
		Value:    avgValue,
	}
}

// GetRecentEntriesByRestaurant retrieves entries for a restaurant sorted by date
// (most recent first) with optional limit for pagination.
func (s *RestaurantAggregationService) GetRecentEntriesByRestaurant(restaurantName string, limit int) ([]models.Entry, error) {
	var entries []models.Entry

	query := s.db.Where("name = ?", restaurantName).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&entries).Error; err != nil {
		return nil, err
	}

	return entries, nil
}

// GetEntryCountByRestaurant returns the total number of entries for a specific
// restaurant for statistical and sorting purposes.
func (s *RestaurantAggregationService) GetEntryCountByRestaurant(restaurantName string) (int64, error) {
	var count int64
	if err := s.db.Model(&models.Entry{}).Where("name = ?", restaurantName).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
