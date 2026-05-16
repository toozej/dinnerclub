package rating

import (
	"testing"
)

func TestGetRatingDescription(t *testing.T) {
	tests := []struct {
		rating   int
		expected RatingDescription
	}{
		{0, RatingDescription{0, "Not Rated", "No rating provided"}},
		{1, RatingDescription{1, "Poor", "Disappointing experience, would not recommend"}},
		{2, RatingDescription{2, "Fair", "Below average, has significant issues"}},
		{3, RatingDescription{3, "Good", "Solid choice, meets expectations"}},
		{4, RatingDescription{4, "Very Good", "Excellent quality, highly recommended"}},
		{5, RatingDescription{5, "Outstanding", "Exceptional experience, destination worthy"}},
		{-1, RatingDescription{0, "Invalid", "Invalid rating value"}},
		{6, RatingDescription{0, "Invalid", "Invalid rating value"}},
	}

	for _, test := range tests {
		result := GetRatingDescription(test.rating)
		if result != test.expected {
			t.Errorf("GetRatingDescription(%d) = %v, expected %v", test.rating, result, test.expected)
		}
	}
}

func TestGetRatingLabel(t *testing.T) {
	tests := []struct {
		rating   int
		expected string
	}{
		{0, "Not Rated"},
		{1, "Poor"},
		{3, "Good"},
		{5, "Outstanding"},
		{-1, "Invalid"},
		{6, "Invalid"},
	}

	for _, test := range tests {
		result := GetRatingLabel(test.rating)
		if result != test.expected {
			t.Errorf("GetRatingLabel(%d) = %s, expected %s", test.rating, result, test.expected)
		}
	}
}

func TestGetRatingFullText(t *testing.T) {
	tests := []struct {
		rating   int
		expected string
	}{
		{0, "Not Rated"},
		{1, "1 - Poor: Disappointing experience, would not recommend"},
		{3, "3 - Good: Solid choice, meets expectations"},
		{5, "5 - Outstanding: Exceptional experience, destination worthy"},
	}

	for _, test := range tests {
		result := GetRatingFullText(test.rating)
		if result != test.expected {
			t.Errorf("GetRatingFullText(%d) = %s, expected %s", test.rating, result, test.expected)
		}
	}
}

func TestValidateRating(t *testing.T) {
	tests := []struct {
		rating   int
		expected bool
	}{
		{0, true},
		{1, true},
		{3, true},
		{5, true},
		{-1, false},
		{6, false},
		{10, false},
	}

	for _, test := range tests {
		result := ValidateRating(test.rating)
		if result != test.expected {
			t.Errorf("ValidateRating(%d) = %v, expected %v", test.rating, result, test.expected)
		}
	}
}

func TestGetAllRatingOptions(t *testing.T) {
	options := GetAllRatingOptions()

	// Should return 5 options (excluding "Not Rated")
	if len(options) != 5 {
		t.Errorf("GetAllRatingOptions() returned %d options, expected 5", len(options))
	}

	// First option should be rating 1
	if options[0].Value != 1 {
		t.Errorf("First rating option value = %d, expected 1", options[0].Value)
	}

	// Last option should be rating 5
	if options[4].Value != 5 {
		t.Errorf("Last rating option value = %d, expected 5", options[4].Value)
	}
}

func TestFormatRatingForDisplay(t *testing.T) {
	tests := []struct {
		rating   int
		expected string
	}{
		{0, "Not Rated"},
		{1, "1 - Poor"},
		{3, "3 - Good"},
		{5, "5 - Outstanding"},
	}

	for _, test := range tests {
		result := FormatRatingForDisplay(test.rating)
		if result != test.expected {
			t.Errorf("FormatRatingForDisplay(%d) = %s, expected %s", test.rating, result, test.expected)
		}
	}
}
