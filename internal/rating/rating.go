// Package rating provides a Michelin-inspired rating system for restaurant reviews
// in the dinnerclub application, with ratings from 1-5 and descriptive labels.
package rating

import (
	"fmt"
)

// RatingDescription represents a rating level with its numeric value, label, and description.
type RatingDescription struct {
	Value       int    `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// Michelin-inspired rating system (1-5 scale)
var RatingDescriptions = []RatingDescription{
	{0, "Not Rated", "No rating provided"},
	{1, "Poor", "Disappointing experience, would not recommend"},
	{2, "Fair", "Below average, has significant issues"},
	{3, "Good", "Solid choice, meets expectations"},
	{4, "Very Good", "Excellent quality, highly recommended"},
	{5, "Outstanding", "Exceptional experience, destination worthy"},
}

// GetRatingDescription returns the complete rating description for a given rating value
// including the numeric value, label, and detailed description text.
func GetRatingDescription(rating int) RatingDescription {
	if rating < 0 || rating > 5 {
		return RatingDescription{0, "Invalid", "Invalid rating value"}
	}
	return RatingDescriptions[rating]
}

// GetRatingLabel returns just the label text for a given rating value
// (e.g., "Good", "Excellent") without numeric value or description.
func GetRatingLabel(rating int) string {
	desc := GetRatingDescription(rating)
	return desc.Label
}

// GetRatingFullText returns formatted text with value, label and description
// in a complete format suitable for detailed display purposes.
func GetRatingFullText(rating int) string {
	desc := GetRatingDescription(rating)
	if rating == 0 {
		return desc.Label
	}
	return fmt.Sprintf("%d - %s: %s", desc.Value, desc.Label, desc.Description)
}

// ValidateRating checks if a rating value is valid (0-5) according to
// the application's rating scale constraints.
func ValidateRating(rating int) bool {
	return rating >= 0 && rating <= 5
}

// GetAllRatingOptions returns all rating options for form dropdowns
// excluding the "Not Rated" option which is for display purposes only.
func GetAllRatingOptions() []RatingDescription {
	return RatingDescriptions[1:] // Exclude "Not Rated" for form options
}

// FormatRatingForDisplay formats a rating for display in templates
// with both numeric value and label for user-friendly presentation.
func FormatRatingForDisplay(rating int) string {
	if rating == 0 {
		return "Not Rated"
	}
	desc := GetRatingDescription(rating)
	return fmt.Sprintf("%d - %s", rating, desc.Label)
}
