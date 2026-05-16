package validation

import (
	"testing"

	"github.com/toozej/dinnerclub/internal/models"
)

func TestValidationHelper_ComprehensiveEntryValidation(t *testing.T) {
	helper := NewValidationHelper()

	tests := []struct {
		name         string
		entry        models.Entry
		expectValid  bool
		expectFields []string
		description  string
	}{
		{
			name: "Perfect visited entry",
			entry: models.Entry{
				Name:           "Perfect Restaurant",
				Location:       "N",
				Cuisine:        "Italian",
				Visited:        true,
				FoodRating:     5,
				AmbienceRating: 4,
				ValueRating:    3,
				Ordered:        "pasta, pizza",
				OrderNext:      "dessert",
				DoNotOrder:     "fish",
				Comments:       "Excellent food and service",
			},
			expectValid: true,
			description: "All fields valid for visited entry",
		},
		{
			name: "Perfect unvisited entry",
			entry: models.Entry{
				Name:             "Future Restaurant",
				Location:         "SE",
				Cuisine:          "Japanese",
				Visited:          false,
				LookingForwardTo: "Trying their famous sushi",
			},
			expectValid: true,
			description: "All fields valid for unvisited entry",
		},
		{
			name: "Boundary values - minimum valid",
			entry: models.Entry{
				Name:       "A", // Minimum length
				Location:   "N",
				Cuisine:    "X", // Minimum length
				Visited:    true,
				FoodRating: 1, // Minimum rating
			},
			expectValid: true,
			description: "Minimum valid values should pass",
		},
		{
			name: "Boundary values - maximum valid",
			entry: models.Entry{
				Name:       "Restaurant with a very long name that is exactly one hundred characters long for testing purposes",
				Location:   "SW",
				Cuisine:    "Cuisine with a very long name that is exactly 50",
				Visited:    true,
				FoodRating: 5, // Maximum rating
			},
			expectValid: true,
			description: "Maximum valid values should pass",
		},
		{
			name: "Name too long",
			entry: models.Entry{
				Name:     "Restaurant with a name that is way too long and exceeds the maximum allowed length of one hundred characters",
				Location: "N",
				Cuisine:  "Italian",
				Visited:  true,
			},
			expectValid:  false,
			expectFields: []string{"name"},
			description:  "Name exceeding 100 characters should fail",
		},
		{
			name: "Cuisine too long",
			entry: models.Entry{
				Name:     "Valid Restaurant",
				Location: "N",
				Cuisine:  "Cuisine name that is way too long and exceeds fifty characters",
				Visited:  true,
			},
			expectValid:  false,
			expectFields: []string{"cuisine"},
			description:  "Cuisine exceeding 50 characters should fail",
		},
		{
			name: "Invalid location",
			entry: models.Entry{
				Name:     "Valid Restaurant",
				Location: "INVALID",
				Cuisine:  "Italian",
				Visited:  true,
			},
			expectValid:  false,
			expectFields: []string{"location"},
			description:  "Invalid location should fail",
		},
		{
			name: "Multiple validation errors",
			entry: models.Entry{
				Name:       "",        // Empty name
				Location:   "INVALID", // Invalid location
				Cuisine:    "",        // Empty cuisine
				Visited:    true,
				FoodRating: 10, // Invalid rating
			},
			expectValid:  false,
			expectFields: []string{"name", "location", "cuisine", "foodrating"},
			description:  "Multiple errors should be caught",
		},
		{
			name: "Unvisited without looking forward to",
			entry: models.Entry{
				Name:     "Restaurant",
				Location: "N",
				Cuisine:  "Italian",
				Visited:  false,
				// Missing LookingForwardTo
			},
			expectValid:  false,
			expectFields: []string{"lookingforwardto"},
			description:  "Unvisited entry must have looking_forward_to",
		},
		{
			name: "Visited with invalid ratings",
			entry: models.Entry{
				Name:           "Restaurant",
				Location:       "N",
				Cuisine:        "Italian",
				Visited:        true,
				FoodRating:     10, // Too high
				AmbienceRating: 6,  // Too high
				ValueRating:    -1, // Too low
			},
			expectValid:  false,
			expectFields: []string{"foodrating", "ambiencerating", "valuerating"},
			description:  "Invalid ratings should fail",
		},
		{
			name: "Edge case - zero ratings for unvisited",
			entry: models.Entry{
				Name:             "Restaurant",
				Location:         "N",
				Cuisine:          "Italian",
				Visited:          false,
				LookingForwardTo: "Something",
				FoodRating:       0, // OK for unvisited
				AmbienceRating:   0, // OK for unvisited
				ValueRating:      0, // OK for unvisited
			},
			expectValid: true,
			description: "Zero ratings OK for unvisited entries",
		},
		{
			name: "Special characters in text fields",
			entry: models.Entry{
				Name:       "Café & Restaurant (Downtown)",
				Location:   "N",
				Cuisine:    "French/Italian",
				Visited:    true,
				FoodRating: 4,
				Comments:   "Great food! 5/5 stars ⭐⭐⭐⭐⭐",
			},
			expectValid: true,
			description: "Special characters should be allowed",
		},
		{
			name: "Unicode characters",
			entry: models.Entry{
				Name:       "寿司レストラン",
				Location:   "N",
				Cuisine:    "日本料理",
				Visited:    true,
				FoodRating: 5,
			},
			expectValid: true,
			description: "Unicode characters should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.ValidateEntry(&tt.entry)

			if tt.expectValid && !result.IsValid {
				t.Errorf("Expected valid entry but got errors: %v. %s",
					result.Errors, tt.description)
			}

			if !tt.expectValid && result.IsValid {
				t.Errorf("Expected invalid entry but validation passed. %s", tt.description)
			}

			if !tt.expectValid {
				for _, expectedField := range tt.expectFields {
					found := false
					for _, fieldError := range result.Errors {
						if fieldError.Field == expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error for field '%s' but not found. %s",
							expectedField, tt.description)
					}
				}
			}
		})
	}
}

func TestValidationHelper_LoginCredentialsEdgeCases(t *testing.T) {
	helper := NewValidationHelper()

	tests := []struct {
		name         string
		username     string
		password     string
		expectValid  bool
		expectFields []string
		description  string
	}{
		{
			name:        "Minimum valid credentials",
			username:    "ab",
			password:    "1234567890",
			expectValid: true,
			description: "Minimum length credentials should pass",
		},
		{
			name:        "Maximum valid username",
			username:    "user123456789012345678901234567890", // 32 chars
			password:    "validpassword123",
			expectValid: true,
			description: "Maximum length username should pass",
		},
		{
			name:        "Username too short",
			username:    "a",
			password:    "validpassword123",
			expectValid: true,
			description: "Single character username should pass (no length validation)",
		},
		{
			name:        "Username too long",
			username:    "user1234567890123456789012345678901", // 33 chars
			password:    "validpassword123",
			expectValid: true,
			description: "Long username should pass (no length validation)",
		},
		{
			name:         "Password too short",
			username:     "validuser",
			password:     "123456789",
			expectValid:  false,
			expectFields: []string{"password"},
			description:  "Password under 8 characters should fail",
		},
		{
			name:         "Username with special characters",
			username:     "user@domain.com",
			password:     "validpassword123",
			expectValid:  false,
			expectFields: []string{"username"},
			description:  "Email-like username should fail",
		},
		{
			name:         "Username with underscores and numbers",
			username:     "user_123_test",
			password:     "validpassword123",
			expectValid:  false,
			expectFields: []string{"username"},
			description:  "Username with underscores should fail",
		},
		{
			name:         "Username with invalid characters",
			username:     "user name", // Space not allowed
			password:     "validpassword",
			expectValid:  false,
			expectFields: []string{"username"},
			description:  "Username with spaces should fail",
		},
		{
			name:        "Password with special characters",
			username:    "validuser",
			password:    "P@ssw0rd!123",
			expectValid: true,
			description: "Password with special characters should pass",
		},
		{
			name:        "Very long password",
			username:    "validuser",
			password:    "verylongpasswordthatexceedsnormallengthbutshouldbefine123456789",
			expectValid: true,
			description: "Long passwords should be allowed",
		},
		{
			name:         "Both fields invalid",
			username:     "user@invalid", // Invalid format
			password:     "123",          // Too short
			expectValid:  false,
			expectFields: []string{"username", "password"},
			description:  "Both invalid fields should be reported",
		},
		{
			name:         "Empty credentials",
			username:     "",
			password:     "",
			expectValid:  false,
			expectFields: []string{"username", "password"},
			description:  "Empty credentials should fail",
		},
		{
			name:         "Whitespace only credentials",
			username:     "   ",
			password:     "   ",
			expectValid:  false,
			expectFields: []string{"username", "password"},
			description:  "Whitespace-only credentials should fail",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.ValidateLoginCredentials(tt.username, tt.password)

			if tt.expectValid && !result.IsValid {
				t.Errorf("Expected valid credentials but got errors: %v. %s",
					result.Errors, tt.description)
			}

			if !tt.expectValid && result.IsValid {
				t.Errorf("Expected invalid credentials but validation passed. %s", tt.description)
			}

			if !tt.expectValid {
				for _, expectedField := range tt.expectFields {
					found := false
					for _, fieldError := range result.Errors {
						if fieldError.Field == expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error for field '%s' but not found. %s",
							expectedField, tt.description)
					}
				}
			}
		})
	}
}

func TestValidationHelper_SearchParametersComprehensive(t *testing.T) {
	helper := NewValidationHelper()

	tests := []struct {
		name         string
		query        string
		page         string
		limit        string
		sortField    string
		sortOrder    string
		expectValid  bool
		expectFields []string
		description  string
	}{
		{
			name:        "All valid parameters",
			query:       "pizza",
			page:        "1",
			limit:       "20",
			sortField:   "name",
			sortOrder:   "asc",
			expectValid: true,
			description: "All valid parameters should pass",
		},
		{
			name:        "Minimum valid query",
			query:       "a",
			page:        "1",
			limit:       "1",
			sortField:   "",
			sortOrder:   "",
			expectValid: true,
			description: "Minimum valid parameters should pass",
		},
		{
			name:        "Maximum valid parameters",
			query:       "very long search query that contains many words and phrases to test the maximum length validation",
			page:        "999",
			limit:       "100",
			sortField:   "created_at",
			sortOrder:   "desc",
			expectValid: true,
			description: "Maximum valid parameters should pass",
		},
		{
			name:         "Empty query",
			query:        "",
			page:         "1",
			limit:        "20",
			expectValid:  false,
			expectFields: []string{"query"},
			description:  "Empty query should fail",
		},
		{
			name:         "Query too long",
			query:        "extremely long search query that exceeds the maximum allowed length and should cause validation to fail because it is way too long for practical use and could cause performance issues in the database and application layer when processing such lengthy queries that users might accidentally or maliciously submit to the system which needs to be prevented by proper validation rules",
			page:         "1",
			limit:        "20",
			expectValid:  false,
			expectFields: []string{"query"},
			description:  "Query over 200 characters should fail",
		},
		{
			name:         "Invalid page number",
			query:        "pizza",
			page:         "0",
			limit:        "20",
			expectValid:  false,
			expectFields: []string{"page"},
			description:  "Page number 0 should fail",
		},
		{
			name:         "Negative page number",
			query:        "pizza",
			page:         "-1",
			limit:        "20",
			expectValid:  false,
			expectFields: []string{"page"},
			description:  "Negative page number should fail",
		},
		{
			name:         "Non-numeric page",
			query:        "pizza",
			page:         "abc",
			limit:        "20",
			expectValid:  false,
			expectFields: []string{"page"},
			description:  "Non-numeric page should fail",
		},
		{
			name:         "Invalid limit - zero",
			query:        "pizza",
			page:         "1",
			limit:        "0",
			expectValid:  false,
			expectFields: []string{"limit"},
			description:  "Limit of 0 should fail",
		},
		{
			name:         "Invalid limit - too high",
			query:        "pizza",
			page:         "1",
			limit:        "101",
			expectValid:  false,
			expectFields: []string{"limit"},
			description:  "Limit over 100 should fail",
		},
		{
			name:         "Invalid sort field",
			query:        "pizza",
			page:         "1",
			limit:        "20",
			sortField:    "invalid_field",
			sortOrder:    "asc",
			expectValid:  false,
			expectFields: []string{"sort_field"},
			description:  "Invalid sort field should fail",
		},
		{
			name:         "Invalid sort order",
			query:        "pizza",
			page:         "1",
			limit:        "20",
			sortField:    "name",
			sortOrder:    "invalid",
			expectValid:  false,
			expectFields: []string{"sort_order"},
			description:  "Invalid sort order should fail",
		},
		{
			name:         "Multiple invalid parameters",
			query:        "",          // Empty
			page:         "0",         // Invalid
			limit:        "200",       // Too high
			sortField:    "bad_field", // Invalid
			sortOrder:    "bad_order", // Invalid
			expectValid:  false,
			expectFields: []string{"query", "page", "limit", "sort_field", "sort_order"},
			description:  "Multiple invalid parameters should all be reported",
		},
		{
			name:        "Special characters in query",
			query:       "pizza & pasta!",
			page:        "1",
			limit:       "20",
			expectValid: true,
			description: "Special characters in query should be allowed",
		},
		{
			name:        "Unicode in query",
			query:       "寿司 pizza",
			page:        "1",
			limit:       "20",
			expectValid: true,
			description: "Unicode characters in query should be allowed",
		},
		{
			name:        "Whitespace in query",
			query:       "  pizza   pasta  ",
			page:        "1",
			limit:       "20",
			expectValid: true,
			description: "Whitespace in query should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.ValidateSearchParameters(tt.query, tt.page, tt.limit, tt.sortField, tt.sortOrder)

			if tt.expectValid && !result.IsValid {
				t.Errorf("Expected valid parameters but got errors: %v. %s",
					result.Errors, tt.description)
			}

			if !tt.expectValid && result.IsValid {
				t.Errorf("Expected invalid parameters but validation passed. %s", tt.description)
			}

			if !tt.expectValid {
				for _, expectedField := range tt.expectFields {
					found := false
					for _, fieldError := range result.Errors {
						if fieldError.Field == expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error for field '%s' but not found. %s",
							expectedField, tt.description)
					}
				}
			}
		})
	}
}

func TestValidationHelper_UserUpdateValidation(t *testing.T) {
	helper := NewValidationHelper()

	tests := []struct {
		name         string
		username     string
		password     string
		email        string
		expectValid  bool
		expectFields []string
		description  string
	}{
		{
			name:        "All valid fields",
			username:    "newuser",
			password:    "newpassword123",
			email:       "user@example.com",
			expectValid: true,
			description: "All valid update fields should pass",
		},
		{
			name:        "Empty fields (valid for update)",
			username:    "",
			password:    "",
			email:       "",
			expectValid: true,
			description: "Empty fields should be allowed for updates",
		},
		{
			name:        "Only username update",
			username:    "newusername",
			password:    "",
			email:       "",
			expectValid: true,
			description: "Partial updates should be allowed",
		},
		{
			name:         "Invalid username format",
			username:     "user name", // Space not allowed
			password:     "",
			email:        "",
			expectValid:  false,
			expectFields: []string{"username"},
			description:  "Invalid username format should fail",
		},
		{
			name:         "Password too short",
			username:     "",
			password:     "123", // Too short
			email:        "",
			expectValid:  false,
			expectFields: []string{"password"},
			description:  "Short password should fail",
		},
		{
			name:         "Invalid email format",
			username:     "",
			password:     "",
			email:        "invalid-email",
			expectValid:  false,
			expectFields: []string{"email"},
			description:  "Invalid email format should fail",
		},
		{
			name:         "Multiple invalid fields",
			username:     "user name", // Invalid
			password:     "123",       // Too short
			email:        "bad-email", // Invalid
			expectValid:  false,
			expectFields: []string{"username", "password", "email"},
			description:  "Multiple invalid fields should all be reported",
		},
		{
			name:        "Valid email formats",
			username:    "",
			password:    "",
			email:       "test+tag@example.co.uk",
			expectValid: true,
			description: "Complex valid email should pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := helper.ValidateUserUpdate(tt.username, tt.password, tt.email)

			if tt.expectValid && !result.IsValid {
				t.Errorf("Expected valid update but got errors: %v. %s",
					result.Errors, tt.description)
			}

			if !tt.expectValid && result.IsValid {
				t.Errorf("Expected invalid update but validation passed. %s", tt.description)
			}

			if !tt.expectValid {
				for _, expectedField := range tt.expectFields {
					found := false
					for _, fieldError := range result.Errors {
						if fieldError.Field == expectedField {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected error for field '%s' but not found. %s",
							expectedField, tt.description)
					}
				}
			}
		})
	}
}

func TestValidationHelper_ConcurrentValidation(t *testing.T) {
	helper := NewValidationHelper()

	// Test concurrent validation calls
	const numGoroutines = 10
	const validationsPerGoroutine = 5

	results := make(chan bool, numGoroutines*validationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < validationsPerGoroutine; j++ {
				entry := models.Entry{
					Name:       "Test Restaurant",
					Location:   "N",
					Cuisine:    "Italian",
					Visited:    true,
					FoodRating: 4,
				}

				result := helper.ValidateEntry(&entry)
				results <- result.IsValid
			}
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines*validationsPerGoroutine; i++ {
		isValid := <-results
		if !isValid {
			t.Error("Concurrent validation failed")
		}
	}
}
