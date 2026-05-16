package validation

import (
	"testing"

	"github.com/toozej/dinnerclub/internal/models"
)

func TestValidateEntry(t *testing.T) {
	// Test valid visited entry
	validVisitedEntry := &models.Entry{
		Name:           "Test Restaurant",
		Location:       "N",
		Cuisine:        "Italian",
		Visited:        true,
		FoodRating:     4,
		AmbienceRating: 3,
		ValueRating:    5,
	}

	errors := ValidateEntry(validVisitedEntry)
	if len(errors) != 0 {
		t.Errorf("Expected no validation errors for valid visited entry, got %d errors", len(errors))
	}

	// Test valid unvisited entry
	validUnvisitedEntry := &models.Entry{
		Name:             "Future Restaurant",
		Location:         "SE",
		Cuisine:          "Mexican",
		Visited:          false,
		LookingForwardTo: "Tacos and margaritas",
	}

	errors = ValidateEntry(validUnvisitedEntry)
	if len(errors) != 0 {
		t.Errorf("Expected no validation errors for valid unvisited entry, got %d errors", len(errors))
	}

	// Test entry with missing required fields
	invalidEntry := &models.Entry{
		Name:     "", // Missing name
		Location: "", // Missing location
		Cuisine:  "", // Missing cuisine
		Visited:  false,
		// Missing LookingForwardTo for unvisited restaurant
	}

	errors = ValidateEntry(invalidEntry)
	if len(errors) != 4 { // Should have 4 errors
		t.Errorf("Expected 4 validation errors for invalid entry, got %d errors", len(errors))
	}

	// Test entry with invalid ratings
	invalidRatingEntry := &models.Entry{
		Name:           "Test Restaurant",
		Location:       "N",
		Cuisine:        "Italian",
		Visited:        true,
		FoodRating:     6,  // Invalid rating
		AmbienceRating: -1, // Invalid rating
		ValueRating:    10, // Invalid rating
	}

	errors = ValidateEntry(invalidRatingEntry)
	ratingErrors := 0
	for _, err := range errors {
		if err.Field == "foodrating" || err.Field == "ambiencerating" || err.Field == "valuerating" {
			ratingErrors++
		}
	}
	if ratingErrors != 3 {
		t.Errorf("Expected 3 rating validation errors, got %d", ratingErrors)
	}

	// Test entry with invalid location
	invalidLocationEntry := &models.Entry{
		Name:     "Test Restaurant",
		Location: "INVALID", // Invalid location
		Cuisine:  "Italian",
		Visited:  true,
	}

	errors = ValidateEntry(invalidLocationEntry)
	locationError := false
	for _, err := range errors {
		if err.Field == "location" {
			locationError = true
			break
		}
	}
	if !locationError {
		t.Error("Expected location validation error for invalid location")
	}
}

func TestValidateRatingValue(t *testing.T) {
	tests := []struct {
		rating    int
		fieldName string
		expectErr bool
	}{
		{0, "foodrating", false},
		{1, "foodrating", false},
		{3, "foodrating", false},
		{5, "foodrating", false},
		{-1, "foodrating", true},
		{6, "foodrating", true},
		{10, "ambiencerating", true},
	}

	for _, test := range tests {
		err := ValidateRatingValue(test.rating, test.fieldName)
		if test.expectErr && err == nil {
			t.Errorf("Expected error for rating %d in field %s, but got none", test.rating, test.fieldName)
		}
		if !test.expectErr && err != nil {
			t.Errorf("Expected no error for rating %d in field %s, but got: %v", test.rating, test.fieldName, err)
		}
	}
}

func TestGetValidationErrorMessages(t *testing.T) {
	errors := []EntryValidationError{
		{Field: "name", Message: "Name is required"},
		{Field: "location", Message: "Location is required"},
		{Field: "foodrating", Message: "Rating must be between 1 and 5"},
	}

	messages := GetValidationErrorMessages(errors)

	if len(messages) != 3 {
		t.Errorf("Expected 3 error messages, got %d", len(messages))
	}

	expectedMessages := []string{
		"Name is required",
		"Location is required",
		"Rating must be between 1 and 5",
	}

	for i, expected := range expectedMessages {
		if messages[i] != expected {
			t.Errorf("Message %d: expected %s, got %s", i, expected, messages[i])
		}
	}
}

func TestEntryValidationError(t *testing.T) {
	err := EntryValidationError{
		Field:   "name",
		Message: "Name is required",
	}

	expected := "name: Name is required"
	if err.Error() != expected {
		t.Errorf("EntryValidationError.Error() = %s, expected %s", err.Error(), expected)
	}
}
