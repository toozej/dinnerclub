// Package validation provides comprehensive validation functionality for the dinnerclub application.
// It includes validation helpers, custom validators, and consistent error handling patterns
// for user input validation across controllers and forms.
package validation

import (
	"fmt"

	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/rating"
)

// EntryValidationError represents validation errors for entries (legacy compatibility)
type EntryValidationError struct {
	Field   string
	Message string
}

func (e EntryValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateEntry performs comprehensive validation on an entry (legacy compatibility)
// Use ValidationHelperInstance.ValidateEntry for new code
func ValidateEntry(entry *models.Entry) []EntryValidationError {
	result := ValidationHelperInstance.ValidateEntry(entry)

	var legacyErrors []EntryValidationError
	for _, err := range result.Errors {
		legacyErrors = append(legacyErrors, EntryValidationError{
			Field:   err.Field,
			Message: err.Message,
		})
	}

	return legacyErrors
}

// ValidateRatingValue validates a single rating value
func ValidateRatingValue(ratingValue int, fieldName string) error {
	if !rating.ValidateRating(ratingValue) {
		return fmt.Errorf("%s must be between 0 and 5", fieldName)
	}
	return nil
}

// GetValidationErrorMessages converts validation errors to user-friendly messages
func GetValidationErrorMessages(errors []EntryValidationError) []string {
	messages := make([]string, len(errors))
	for i, err := range errors {
		messages[i] = err.Message
	}
	return messages
}
