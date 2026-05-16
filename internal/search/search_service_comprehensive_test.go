package search

import (
	"testing"

	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupComprehensiveTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err := db.AutoMigrate(&models.Entry{}, &models.Restaurant{}); err != nil {
		panic("Failed to migrate comprehensive test database: " + err.Error())
	}
	return db
}

func TestSearchService_ComplexQueryCombinations(t *testing.T) {
	db := setupComprehensiveTestDB()
	service := NewSearchService(db)

	// Create comprehensive test data
	entries := []models.Entry{
		{
			Name:           "Italian Bistro",
			Submitter:      "alice",
			Cuisine:        "Italian",
			Location:       "N",
			Visited:        true,
			FoodRating:     5,
			AmbienceRating: 4,
			ValueRating:    4,
			Comments:       "Excellent pasta and pizza",
			Ordered:        "pasta, pizza, wine",
		},
		{
			Name:             "Sushi Express",
			Submitter:        "bob",
			Cuisine:          "Japanese",
			Location:         "SE",
			Visited:          false,
			LookingForwardTo: "Fresh sashimi and rolls",
		},
		{
			Name:           "Taco Fiesta",
			Submitter:      "alice",
			Cuisine:        "Mexican",
			Location:       "SW",
			Visited:        true,
			FoodRating:     4,
			AmbienceRating: 3,
			ValueRating:    5,
			Comments:       "Great tacos and margaritas",
			Ordered:        "tacos, guacamole, margaritas",
		},
		{
			Name:           "Pizza Corner",
			Submitter:      "charlie",
			Cuisine:        "Italian",
			Location:       "NE",
			Visited:        true,
			FoodRating:     3,
			AmbienceRating: 2,
			ValueRating:    4,
			Comments:       "Decent pizza, good value",
			Ordered:        "pizza, breadsticks",
		},
		{
			Name:             "Thai Garden",
			Submitter:        "bob",
			Cuisine:          "Thai",
			Location:         "S",
			Visited:          false,
			LookingForwardTo: "Pad thai and curry",
		},
	}

	for _, entry := range entries {
		db.Create(&entry)
	}

	restaurants := []models.Restaurant{
		{Name: "Italian Bistro", Cuisine: "Italian", Location: "N"},
		{Name: "Sushi Express", Cuisine: "Japanese", Location: "SE"},
		{Name: "Taco Fiesta", Cuisine: "Mexican", Location: "SW"},
		{Name: "Pizza Corner", Cuisine: "Italian", Location: "NE"},
		{Name: "Thai Garden", Cuisine: "Thai", Location: "S"},
	}

	for _, restaurant := range restaurants {
		db.Create(&restaurant)
	}

	tests := []struct {
		name            string
		queryString     string
		expectedEntries int
		expectedRests   int
		description     string
	}{
		{
			name:            "Author and cuisine combination",
			queryString:     "author:alice italian",
			expectedEntries: 1, // Only Italian Bistro by alice
			expectedRests:   0, // No restaurants for author queries
			description:     "Should find entries by alice with italian cuisine",
		},
		{
			name:            "Wishlist with location",
			queryString:     "wishlist location:SE",
			expectedEntries: 1, // Sushi Express (unvisited in SE)
			expectedRests:   0, // No restaurants for wishlist queries
			description:     "Should find unvisited entries in SE location",
		},
		{
			name:            "Rating with cuisine",
			queryString:     "rating:4 cuisine:italian",
			expectedEntries: 2, // Italian Bistro (5 rating) and Pizza Corner (4 rating in value)
			expectedRests:   0, // No restaurants for rating queries
			description:     "Should find Italian entries with rating >= 4",
		},
		{
			name:            "Pizza text search",
			queryString:     "pizza",
			expectedEntries: 2, // Italian Bistro (has pizza), Pizza Corner (has pizza)
			expectedRests:   1, // Only Pizza Corner restaurant has "pizza" in name
			description:     "Should find entries and restaurants matching pizza",
		},
		{
			name:            "Location with cuisine",
			queryString:     "location:N cuisine:italian",
			expectedEntries: 1, // Italian Bistro
			expectedRests:   1, // Italian Bistro restaurant
			description:     "Should find Italian restaurants in North location",
		},
		{
			name:            "Multiple author search",
			queryString:     "author:bob",
			expectedEntries: 2, // Sushi Express, Thai Garden
			expectedRests:   0, // No restaurants for author queries
			description:     "Should find all entries by bob",
		},
		{
			name:            "Wishlist with text",
			queryString:     "want to try sushi",
			expectedEntries: 1, // Sushi Express (wishlist + text match)
			expectedRests:   0, // No restaurants for wishlist queries
			description:     "Should find wishlist entries matching text",
		},
		{
			name:            "High rating filter",
			queryString:     "rating:5",
			expectedEntries: 2, // Italian Bistro (5 food rating) and Taco Fiesta (5 value rating)
			expectedRests:   0, // No restaurants for rating queries
			description:     "Should find entries with any rating >= 5",
		},
		{
			name:            "Empty result combination",
			queryString:     "author:alice location:SE",
			expectedEntries: 0, // Alice has no entries in SE
			expectedRests:   0, // No restaurants for author queries
			description:     "Should return empty results for impossible combinations",
		},
		{
			name:            "Case insensitive search",
			queryString:     "ITALIAN",
			expectedEntries: 2, // Italian Bistro, Pizza Corner (both Italian cuisine)
			expectedRests:   2, // Italian Bistro, Pizza Corner
			description:     "Should handle case insensitive searches",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Search(tt.queryString)
			if err != nil {
				t.Fatalf("Search failed: %v", err)
			}

			if len(result.Entries) != tt.expectedEntries {
				t.Errorf("Expected %d entries, got %d. %s",
					tt.expectedEntries, len(result.Entries), tt.description)
			}

			if len(result.Restaurants) != tt.expectedRests {
				t.Errorf("Expected %d restaurants, got %d. %s",
					tt.expectedRests, len(result.Restaurants), tt.description)
			}
		})
	}
}

func TestSearchService_EdgeCases(t *testing.T) {
	db := setupComprehensiveTestDB()
	service := NewSearchService(db)

	tests := []struct {
		name        string
		queryString string
		expectError bool
		description string
	}{
		{
			name:        "Empty query",
			queryString: "",
			expectError: false,
			description: "Should handle empty queries gracefully",
		},
		{
			name:        "Only whitespace",
			queryString: "   \t\n   ",
			expectError: false,
			description: "Should handle whitespace-only queries",
		},
		{
			name:        "Special characters",
			queryString: "pizza & pasta!",
			expectError: false,
			description: "Should handle special characters in search",
		},
		{
			name:        "Very long query",
			queryString: "this is a very long search query that contains many words and should still work properly without causing any issues",
			expectError: false,
			description: "Should handle long queries",
		},
		{
			name:        "Invalid location",
			queryString: "location:INVALID",
			expectError: false,
			description: "Should handle invalid location gracefully",
		},
		{
			name:        "Invalid rating",
			queryString: "rating:10",
			expectError: false,
			description: "Should handle invalid rating values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.Search(tt.queryString)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none. %s", tt.description)
			}

			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v. %s", err, tt.description)
			}

			if !tt.expectError {
				// Should return valid result structure even for edge cases
				if result.Entries == nil {
					t.Error("Entries slice should not be nil")
				}
				if result.Restaurants == nil {
					t.Error("Restaurants slice should not be nil")
				}
			}
		})
	}
}

func TestSearchService_QueryParsing(t *testing.T) {
	service := NewSearchService(nil) // No DB needed for parsing tests

	tests := []struct {
		name     string
		input    string
		expected SearchQuery
	}{
		{
			name:  "Multiple filters",
			input: "author:alice location:N cuisine:italian rating:4 great food",
			expected: SearchQuery{
				Text:     "great food",
				Author:   "alice",
				Location: "N",
				Cuisine:  "italian",
				Rating:   4,
			},
		},
		{
			name:  "Wishlist with multiple keywords",
			input: "i want to go to thai restaurant in location:SE",
			expected: SearchQuery{
				Text:     "thai restaurant in",
				Location: "SE",
				Wishlist: true,
				Visited:  &[]bool{false}[0],
			},
		},
		{
			name:  "Case insensitive filters",
			input: "AUTHOR:ALICE LOCATION:N CUISINE:ITALIAN",
			expected: SearchQuery{
				Text:     "",
				Author:   "alice",
				Location: "N",
				Cuisine:  "ITALIAN",
			},
		},
		{
			name:  "Multiple wishlist patterns",
			input: "want to try some sushi",
			expected: SearchQuery{
				Text:     "some sushi",
				Wishlist: true,
				Visited:  &[]bool{false}[0],
			},
		},
		{
			name:  "Filter without value",
			input: "author: location:N",
			expected: SearchQuery{
				Text:     "author:",
				Location: "N",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ParseQuery(tt.input)

			if result.Text != tt.expected.Text {
				t.Errorf("Expected text '%s', got '%s'", tt.expected.Text, result.Text)
			}

			if result.Author != tt.expected.Author {
				t.Errorf("Expected author '%s', got '%s'", tt.expected.Author, result.Author)
			}

			if result.Location != tt.expected.Location {
				t.Errorf("Expected location '%s', got '%s'", tt.expected.Location, result.Location)
			}

			if result.Cuisine != tt.expected.Cuisine {
				t.Errorf("Expected cuisine '%s', got '%s'", tt.expected.Cuisine, result.Cuisine)
			}

			if result.Rating != tt.expected.Rating {
				t.Errorf("Expected rating %d, got %d", tt.expected.Rating, result.Rating)
			}

			if result.Wishlist != tt.expected.Wishlist {
				t.Errorf("Expected wishlist %v, got %v", tt.expected.Wishlist, result.Wishlist)
			}

			if tt.expected.Visited != nil {
				if result.Visited == nil || *result.Visited != *tt.expected.Visited {
					t.Errorf("Expected visited %v, got %v", tt.expected.Visited, result.Visited)
				}
			} else if result.Visited != nil {
				t.Errorf("Expected visited nil, got %v", result.Visited)
			}
		})
	}
}

func TestSearchService_SearchSuggestionsComprehensive(t *testing.T) {
	db := setupComprehensiveTestDB()
	service := NewSearchService(db)

	// Create test data
	restaurants := []models.Restaurant{
		{Name: "Pizza Palace", Cuisine: "Italian"},
		{Name: "Pizza Express", Cuisine: "Italian"},
		{Name: "Pasta House", Cuisine: "Italian"},
		{Name: "Sushi Bar", Cuisine: "Japanese"},
	}

	for _, restaurant := range restaurants {
		db.Create(&restaurant)
	}

	entries := []models.Entry{
		{Name: "Pizza Palace", Cuisine: "Italian"},
		{Name: "Burger Joint", Cuisine: "American"},
	}

	for _, entry := range entries {
		db.Create(&entry)
	}

	tests := []struct {
		name             string
		partial          string
		expectedMin      int
		expectedMax      int
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:          "Pizza suggestions",
			partial:       "piz",
			expectedMin:   2,
			expectedMax:   5,
			shouldContain: []string{"Pizza Palace", "Pizza Express"},
		},
		{
			name:          "Short query",
			partial:       "p",
			expectedMin:   0,
			expectedMax:   5,
			shouldContain: []string{},
		},
		{
			name:             "No matches",
			partial:          "xyz",
			expectedMin:      0,
			expectedMax:      0,
			shouldNotContain: []string{"Pizza Palace"},
		},
		{
			name:          "Case insensitive",
			partial:       "PIZZA",
			expectedMin:   2,
			expectedMax:   5,
			shouldContain: []string{"Pizza Palace", "Pizza Express"},
		},
		{
			name:          "Cuisine suggestions",
			partial:       "ital",
			expectedMin:   1,
			expectedMax:   5,
			shouldContain: []string{"cuisine:Italian"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions, err := service.GetSearchSuggestions(tt.partial)
			if err != nil {
				t.Fatalf("GetSearchSuggestions failed: %v", err)
			}

			if len(suggestions) < tt.expectedMin {
				t.Errorf("Expected at least %d suggestions, got %d", tt.expectedMin, len(suggestions))
			}

			if len(suggestions) > tt.expectedMax {
				t.Errorf("Expected at most %d suggestions, got %d", tt.expectedMax, len(suggestions))
			}

			for _, expected := range tt.shouldContain {
				found := false
				for _, suggestion := range suggestions {
					if suggestion == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected suggestion '%s' not found", expected)
				}
			}

			for _, notExpected := range tt.shouldNotContain {
				for _, suggestion := range suggestions {
					if suggestion == notExpected {
						t.Errorf("Unexpected suggestion '%s' found", notExpected)
					}
				}
			}
		})
	}
}
