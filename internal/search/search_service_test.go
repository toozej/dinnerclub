package search

import (
	"testing"

	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err := db.AutoMigrate(&models.Entry{}, &models.Restaurant{}); err != nil {
		panic("Failed to migrate test database: " + err.Error())
	}
	return db
}

func TestParseQuery(t *testing.T) {
	service := NewSearchService(nil)

	tests := []struct {
		input    string
		expected SearchQuery
	}{
		{
			input: "pizza",
			expected: SearchQuery{
				Text: "pizza",
			},
		},
		{
			input: "wishlist italian",
			expected: SearchQuery{
				Text:     "italian",
				Wishlist: true,
				Visited:  &[]bool{false}[0],
			},
		},
		{
			input: "by:john pizza",
			expected: SearchQuery{
				Text:   "pizza",
				Author: "john",
			},
		},
		{
			input: "location:N sushi",
			expected: SearchQuery{
				Text:     "sushi",
				Location: "N",
			},
		},
		{
			input: "cuisine:mexican rating:4",
			expected: SearchQuery{
				Text:    "",
				Cuisine: "mexican",
				Rating:  4,
			},
		},
		{
			input: "want to go to thai restaurant",
			expected: SearchQuery{
				Text:     "thai restaurant",
				Wishlist: true,
				Visited:  &[]bool{false}[0],
			},
		},
		{
			input: "author:alice location:SE cuisine:italian rating:5",
			expected: SearchQuery{
				Text:     "",
				Author:   "alice",
				Location: "SE",
				Cuisine:  "italian",
				Rating:   5,
			},
		},
	}

	for _, test := range tests {
		result := service.ParseQuery(test.input)

		if result.Text != test.expected.Text {
			t.Errorf("ParseQuery(%s) Text = %s, expected %s", test.input, result.Text, test.expected.Text)
		}

		if result.Author != test.expected.Author {
			t.Errorf("ParseQuery(%s) Author = %s, expected %s", test.input, result.Author, test.expected.Author)
		}

		if result.Wishlist != test.expected.Wishlist {
			t.Errorf("ParseQuery(%s) Wishlist = %v, expected %v", test.input, result.Wishlist, test.expected.Wishlist)
		}

		if result.Location != test.expected.Location {
			t.Errorf("ParseQuery(%s) Location = %s, expected %s", test.input, result.Location, test.expected.Location)
		}

		if result.Cuisine != test.expected.Cuisine {
			t.Errorf("ParseQuery(%s) Cuisine = %s, expected %s", test.input, result.Cuisine, test.expected.Cuisine)
		}

		if result.Rating != test.expected.Rating {
			t.Errorf("ParseQuery(%s) Rating = %d, expected %d", test.input, result.Rating, test.expected.Rating)
		}

		if test.expected.Visited != nil {
			if result.Visited == nil || *result.Visited != *test.expected.Visited {
				t.Errorf("ParseQuery(%s) Visited = %v, expected %v", test.input, result.Visited, test.expected.Visited)
			}
		} else if result.Visited != nil {
			t.Errorf("ParseQuery(%s) Visited = %v, expected nil", test.input, result.Visited)
		}
	}
}

func TestSearchEntries(t *testing.T) {
	db := setupTestDB()
	service := NewSearchService(db)

	// Create test data
	entries := []models.Entry{
		{
			Name:       "Pizza Palace",
			Submitter:  "john",
			Cuisine:    "Italian",
			Location:   "N",
			Visited:    true,
			FoodRating: 4,
			Comments:   "Great pizza",
		},
		{
			Name:             "Sushi Bar",
			Submitter:        "alice",
			Cuisine:          "Japanese",
			Location:         "SE",
			Visited:          false,
			LookingForwardTo: "Fresh sashimi",
		},
		{
			Name:       "Taco Truck",
			Submitter:  "bob",
			Cuisine:    "Mexican",
			Location:   "SW",
			Visited:    true,
			FoodRating: 5,
			Comments:   "Amazing tacos",
		},
	}

	for _, entry := range entries {
		db.Create(&entry)
	}

	tests := []struct {
		query    SearchQuery
		expected int
	}{
		{
			query:    SearchQuery{Text: "pizza"},
			expected: 1,
		},
		{
			query:    SearchQuery{Author: "alice"},
			expected: 1,
		},
		{
			query:    SearchQuery{Visited: &[]bool{false}[0]},
			expected: 1,
		},
		{
			query:    SearchQuery{Location: "N"},
			expected: 1,
		},
		{
			query:    SearchQuery{Cuisine: "italian"},
			expected: 1,
		},
		{
			query:    SearchQuery{Rating: 4},
			expected: 2, // Both entries with rating >= 4
		},
		{
			query:    SearchQuery{Text: "nonexistent"},
			expected: 0,
		},
	}

	for _, test := range tests {
		results, err := service.SearchEntries(test.query)
		if err != nil {
			t.Errorf("SearchEntries failed: %v", err)
			continue
		}

		if len(results) != test.expected {
			t.Errorf("SearchEntries returned %d results, expected %d", len(results), test.expected)
		}
	}
}

func TestSearchRestaurants(t *testing.T) {
	db := setupTestDB()
	service := NewSearchService(db)

	// Create test data
	restaurants := []models.Restaurant{
		{
			Name:     "Pizza Palace",
			Cuisine:  "Italian",
			Location: "N",
		},
		{
			Name:     "Sushi Bar",
			Cuisine:  "Japanese",
			Location: "SE",
		},
		{
			Name:     "Taco Truck",
			Cuisine:  "Mexican",
			Location: "SW",
		},
	}

	for _, restaurant := range restaurants {
		db.Create(&restaurant)
	}

	tests := []struct {
		query    SearchQuery
		expected int
	}{
		{
			query:    SearchQuery{Text: "pizza"},
			expected: 1,
		},
		{
			query:    SearchQuery{Location: "N"},
			expected: 1,
		},
		{
			query:    SearchQuery{Cuisine: "italian"},
			expected: 1,
		},
		{
			query:    SearchQuery{Text: "nonexistent"},
			expected: 0,
		},
	}

	for _, test := range tests {
		results, err := service.SearchRestaurants(test.query)
		if err != nil {
			t.Errorf("SearchRestaurants failed: %v", err)
			continue
		}

		if len(results) != test.expected {
			t.Errorf("SearchRestaurants returned %d results, expected %d", len(results), test.expected)
		}
	}
}

func TestSearch(t *testing.T) {
	db := setupTestDB()
	service := NewSearchService(db)

	// Create test data
	entry := models.Entry{
		Name:      "Pizza Palace",
		Submitter: "john",
		Cuisine:   "Italian",
		Location:  "N",
		Visited:   true,
	}
	db.Create(&entry)

	restaurant := models.Restaurant{
		Name:     "Pizza Palace",
		Cuisine:  "Italian",
		Location: "N",
	}
	db.Create(&restaurant)

	result, err := service.Search("pizza")
	if err != nil {
		t.Errorf("Search failed: %v", err)
	}

	if result.TotalCount != 2 {
		t.Errorf("Search returned %d total results, expected 2", result.TotalCount)
	}

	if len(result.Entries) != 1 {
		t.Errorf("Search returned %d entries, expected 1", len(result.Entries))
	}

	if len(result.Restaurants) != 1 {
		t.Errorf("Search returned %d restaurants, expected 1", len(result.Restaurants))
	}
}

func TestGetSearchSuggestions(t *testing.T) {
	db := setupTestDB()
	service := NewSearchService(db)

	// Create test data
	restaurant := models.Restaurant{
		Name:     "Pizza Palace",
		Cuisine:  "Italian",
		Location: "N",
	}
	db.Create(&restaurant)

	entry := models.Entry{
		Name:    "Pizza Palace",
		Cuisine: "Italian",
	}
	db.Create(&entry)

	suggestions, err := service.GetSearchSuggestions("piz")
	if err != nil {
		t.Errorf("GetSearchSuggestions failed: %v", err)
	}

	if len(suggestions) == 0 {
		t.Error("GetSearchSuggestions returned no suggestions")
	}

	// Should contain "Pizza Palace"
	found := false
	for _, suggestion := range suggestions {
		if suggestion == "Pizza Palace" {
			found = true
			break
		}
	}

	if !found {
		t.Error("GetSearchSuggestions did not return expected restaurant name")
	}
}
