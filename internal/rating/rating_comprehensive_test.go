package rating

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestRatingSystem_ComprehensiveValidation(t *testing.T) {
	tests := []struct {
		name        string
		rating      int
		expectValid bool
		description string
	}{
		{
			name:        "Minimum valid rating",
			rating:      1,
			expectValid: true,
			description: "Rating of 1 should be valid",
		},
		{
			name:        "Maximum valid rating",
			rating:      5,
			expectValid: true,
			description: "Rating of 5 should be valid",
		},
		{
			name:        "Mid-range rating",
			rating:      3,
			expectValid: true,
			description: "Rating of 3 should be valid",
		},
		{
			name:        "Zero rating",
			rating:      0,
			expectValid: true,
			description: "Rating of 0 should be valid (Not Rated)",
		},
		{
			name:        "Negative rating",
			rating:      -1,
			expectValid: false,
			description: "Negative rating should be invalid",
		},
		{
			name:        "Rating too high",
			rating:      6,
			expectValid: false,
			description: "Rating above 5 should be invalid",
		},
		{
			name:        "Very high rating",
			rating:      100,
			expectValid: false,
			description: "Very high rating should be invalid",
		},
		{
			name:        "Very low rating",
			rating:      -100,
			expectValid: false,
			description: "Very low rating should be invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := ValidateRating(tt.rating)

			if tt.expectValid && !isValid {
				t.Errorf("Expected rating %d to be valid but it was invalid. %s",
					tt.rating, tt.description)
			}

			if !tt.expectValid && isValid {
				t.Errorf("Expected rating %d to be invalid but it was valid. %s",
					tt.rating, tt.description)
			}
		})
	}
}

func TestRatingSystem_DescriptionRetrieval(t *testing.T) {
	tests := []struct {
		name          string
		rating        int
		expectedLabel string
		expectedDesc  string
		expectError   bool
		description   string
	}{
		{
			name:          "Rating 1 description",
			rating:        1,
			expectedLabel: "Poor",
			expectedDesc:  "Disappointing experience, would not recommend",
			expectError:   false,
			description:   "Should return correct description for rating 1",
		},
		{
			name:          "Rating 2 description",
			rating:        2,
			expectedLabel: "Fair",
			expectedDesc:  "Below average, has significant issues",
			expectError:   false,
			description:   "Should return correct description for rating 2",
		},
		{
			name:          "Rating 3 description",
			rating:        3,
			expectedLabel: "Good",
			expectedDesc:  "Solid choice, meets expectations",
			expectError:   false,
			description:   "Should return correct description for rating 3",
		},
		{
			name:          "Rating 4 description",
			rating:        4,
			expectedLabel: "Very Good",
			expectedDesc:  "Excellent quality, highly recommended",
			expectError:   false,
			description:   "Should return correct description for rating 4",
		},
		{
			name:          "Rating 5 description",
			rating:        5,
			expectedLabel: "Outstanding",
			expectedDesc:  "Exceptional experience, destination worthy",
			expectError:   false,
			description:   "Should return correct description for rating 5",
		},
		{
			name:          "Rating 0 description",
			rating:        0,
			expectedLabel: "Not Rated",
			expectedDesc:  "No rating provided",
			expectError:   false,
			description:   "Should return correct description for rating 0",
		},
		{
			name:        "Invalid rating 6",
			rating:      6,
			expectError: true,
			description: "Should return error for rating 6",
		},
		{
			name:        "Invalid negative rating",
			rating:      -1,
			expectError: true,
			description: "Should return error for negative rating",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test GetRatingDescription
			desc := GetRatingDescription(tt.rating)

			if tt.expectError {
				if desc.Label != "Invalid" {
					t.Errorf("Expected invalid rating for rating %d but got %s", tt.rating, desc.Label)
				}
			} else {
				if desc.Value != tt.rating {
					t.Errorf("Expected value %d, got %d", tt.rating, desc.Value)
				}

				if desc.Label != tt.expectedLabel {
					t.Errorf("Expected label '%s', got '%s'", tt.expectedLabel, desc.Label)
				}

				if desc.Description != tt.expectedDesc {
					t.Errorf("Expected description '%s', got '%s'", tt.expectedDesc, desc.Description)
				}
			}

			// Test GetRatingLabel
			label := GetRatingLabel(tt.rating)
			if !tt.expectError && label != tt.expectedLabel {
				t.Errorf("GetRatingLabel: Expected '%s', got '%s'", tt.expectedLabel, label)
			}
			if tt.expectError && label != "Invalid" {
				t.Errorf("GetRatingLabel: Expected 'Invalid' for invalid rating, got '%s'", label)
			}

			// Test GetRatingFullText
			fullText := GetRatingFullText(tt.rating)
			if !tt.expectError {
				var expectedFullText string
				if tt.rating == 0 {
					expectedFullText = tt.expectedLabel
				} else {
					expectedFullText = fmt.Sprintf("%d - %s: %s", tt.rating, tt.expectedLabel, tt.expectedDesc)
				}
				if fullText != expectedFullText {
					t.Errorf("GetRatingFullText: Expected '%s', got '%s'", expectedFullText, fullText)
				}
			} else if !strings.Contains(fullText, "Invalid") {
				t.Errorf("GetRatingFullText: Expected text containing 'Invalid' for invalid rating, got '%s'", fullText)
			}
		})
	}
}

func TestRatingSystem_AllRatingOptions(t *testing.T) {
	options := GetAllRatingOptions()

	// Should have exactly 5 options
	if len(options) != 5 {
		t.Errorf("Expected 5 rating options, got %d", len(options))
	}

	// Verify all ratings 1-5 are present
	expectedRatings := []int{1, 2, 3, 4, 5}
	for _, expectedRating := range expectedRatings {
		found := false
		for _, option := range options {
			if option.Value == expectedRating {
				found = true

				// Verify the option has all required fields
				if option.Label == "" {
					t.Errorf("Rating %d has empty label", expectedRating)
				}
				if option.Description == "" {
					t.Errorf("Rating %d has empty description", expectedRating)
				}
				break
			}
		}
		if !found {
			t.Errorf("Rating %d not found in options", expectedRating)
		}
	}

	// Verify options are in ascending order
	for i := 1; i < len(options); i++ {
		if options[i-1].Value >= options[i].Value {
			t.Error("Rating options are not in ascending order")
		}
	}
}

func TestRatingSystem_DisplayFormatting(t *testing.T) {
	tests := []struct {
		name     string
		rating   int
		expected string
	}{
		{
			name:     "Format rating 1",
			rating:   1,
			expected: "1 - Poor",
		},
		{
			name:     "Format rating 3",
			rating:   3,
			expected: "3 - Good",
		},
		{
			name:     "Format rating 5",
			rating:   5,
			expected: "5 - Outstanding",
		},
		{
			name:     "Format zero rating",
			rating:   0,
			expected: "Not Rated",
		},
		{
			name:     "Format negative rating",
			rating:   -1,
			expected: "-1 - Invalid",
		},
		{
			name:     "Format high rating",
			rating:   10,
			expected: "10 - Invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRatingForDisplay(tt.rating)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRatingSystem_EdgeCases(t *testing.T) {
	// Test with extreme values
	extremeValues := []int{
		-999999,
		-1,
		0,
		6,
		999999,
	}

	for _, value := range extremeValues {
		t.Run("Extreme value "+string(rune(value+'0')), func(t *testing.T) {
			// Should not panic
			isValid := ValidateRating(value)
			if value == 0 {
				if !isValid {
					t.Errorf("Expected rating %d to be valid (Not Rated)", value)
				}
				label := GetRatingLabel(value)
				if label != "Not Rated" {
					t.Errorf("Expected 'Not Rated' for rating 0, got '%s'", label)
				}
				fullText := GetRatingFullText(value)
				if fullText != "Not Rated" {
					t.Errorf("Expected 'Not Rated' for rating 0, got '%s'", fullText)
				}
			} else {
				if isValid {
					t.Errorf("Expected rating %d to be invalid", value)
				}
				label := GetRatingLabel(value)
				if label != "Invalid" {
					t.Errorf("Expected 'Invalid' for invalid rating %d, got '%s'", value, label)
				}
				fullText := GetRatingFullText(value)
				if !strings.Contains(fullText, "Invalid") {
					t.Errorf("Expected text containing 'Invalid' for invalid rating %d, got '%s'", value, fullText)
				}
			}
		})
	}
}

func TestRatingSystem_Consistency(t *testing.T) {
	// Test that all rating functions return consistent results
	for rating := 1; rating <= 5; rating++ {
		t.Run("Consistency check for rating "+string(rune(rating+'0')), func(t *testing.T) {
			// Get description
			desc := GetRatingDescription(rating)

			// Get label
			label := GetRatingLabel(rating)

			// Get full text
			fullText := GetRatingFullText(rating)

			// Verify consistency
			if desc.Label != label {
				t.Errorf("Inconsistent label: GetRatingDescription returned '%s', GetRatingLabel returned '%s'",
					desc.Label, label)
			}

			expectedFullText := fmt.Sprintf("%d - %s: %s", rating, label, desc.Description)
			if fullText != expectedFullText {
				t.Errorf("Inconsistent full text: expected '%s', got '%s'", expectedFullText, fullText)
			}

			// Verify validation
			if !ValidateRating(rating) {
				t.Errorf("ValidateRating returned false for valid rating %d", rating)
			}

			// Verify in all options
			options := GetAllRatingOptions()
			found := false
			for _, option := range options {
				if option.Value == rating {
					found = true
					if option.Label != label {
						t.Errorf("Inconsistent label in options: expected '%s', got '%s'", label, option.Label)
					}
					if option.Description != desc.Description {
						t.Errorf("Inconsistent description in options: expected '%s', got '%s'",
							desc.Description, option.Description)
					}
					break
				}
			}
			if !found {
				t.Errorf("Rating %d not found in GetAllRatingOptions", rating)
			}
		})
	}
}

func TestRatingSystem_ConcurrentAccess(t *testing.T) {
	// Test concurrent access to rating functions
	const numGoroutines = 10
	const operationsPerGoroutine = 100

	results := make(chan error, numGoroutines*operationsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < operationsPerGoroutine; j++ {
				rating := (j % 5) + 1 // Ratings 1-5

				// Test all functions concurrently
				if !ValidateRating(rating) {
					results <- assert.AnError
					continue
				}

				label := GetRatingLabel(rating)
				if label == "Unknown" {
					results <- assert.AnError
					continue
				}

				desc := GetRatingDescription(rating)
				if desc.Label == "Invalid" {
					results <- assert.AnError
					continue
				}

				fullText := GetRatingFullText(rating)
				if fullText == "Unknown rating" {
					results <- assert.AnError
					continue
				}

				results <- nil
			}
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines*operationsPerGoroutine; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Concurrent access error: %v", err)
		}
	}
}

func TestRatingSystem_MemoryEfficiency(t *testing.T) {
	// Test that repeated calls don't cause memory issues
	for i := 0; i < 1000; i++ {
		rating := (i % 5) + 1

		_ = ValidateRating(rating)
		_ = GetRatingLabel(rating)
		_ = GetRatingFullText(rating)
		_ = GetRatingDescription(rating)
		_ = GetAllRatingOptions()
		_ = FormatRatingForDisplay(rating)
	}

	// If we get here without panicking or running out of memory, test passes
}

// Helper function for concurrent tests
var assert = struct {
	AnError error
}{
	AnError: errors.New("assertion error"),
}
