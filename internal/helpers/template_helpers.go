// Package helpers provides template helper functions and utilities
// for the dinnerclub web application's HTML templates.
package helpers

import (
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/rating"
	"github.com/toozej/dinnerclub/pkg/csrf"
)

// TemplateHelpers contains helper functions for HTML templates.
type TemplateHelpers struct{}

// FormatBoolean converts boolean values to user-friendly yes/no strings.
func (h *TemplateHelpers) FormatBoolean(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

// FormatRating formats a rating value for display in templates.
func (h *TemplateHelpers) FormatRating(ratingValue int) string {
	return rating.FormatRatingForDisplay(ratingValue)
}

// GetRatingDescription returns the full description text for a rating value
// including the numeric value, label, and detailed description.
func (h *TemplateHelpers) GetRatingDescription(ratingValue int) string {
	return rating.GetRatingFullText(ratingValue)
}

// GetRatingLabel returns just the label text for a rating value
// (e.g., "Good", "Excellent") without the numeric value or description.
func (h *TemplateHelpers) GetRatingLabel(ratingValue int) string {
	return rating.GetRatingLabel(ratingValue)
}

// UserProfileLink creates a hyperlink to a user's submitted entries page
// with proper HTML escaping to prevent XSS attacks.
func (h *TemplateHelpers) UserProfileLink(username string) template.HTML {
	if username == "" {
		return template.HTML("Anonymous")
	}
	// #nosec G203 - This is safe because we're escaping the username content
	return template.HTML(`<a href="/entries/submittedby/` + template.HTMLEscapeString(username) + `">` + template.HTMLEscapeString(username) + `</a>`)
}

// TooltipText returns helpful tooltip text for form fields based on field name
// to provide users with guidance on how to fill out form fields correctly.
func (h *TemplateHelpers) TooltipText(fieldName string) string {
	tooltips := map[string]string{
		"name":             "Enter the restaurant name",
		"location":         "Select the general area/direction of the restaurant",
		"cuisine":          "Type of cuisine (e.g., Italian, Mexican, Thai)",
		"visited":          "Check if you have visited this restaurant",
		"lookingforwardto": "What dishes or experiences are you excited to try?",
		"closed":           "Check if the restaurant is permanently closed",
		"mealservice":      "What type of meal did you have?",
		"ordered":          "List items you ordered (separate multiple items with commas)",
		"ordernext":        "Items you want to try on your next visit (separate with commas)",
		"donotorder":       "Items you wouldn't order again (separate with commas)",
		"foodrating":       "Rate the food quality (1=Poor, 2=Fair, 3=Good, 4=Very Good, 5=Outstanding)",
		"ambiencerating":   "Rate the atmosphere and ambience (1=Poor, 2=Fair, 3=Good, 4=Very Good, 5=Outstanding)",
		"valuerating":      "Rate the value for money (1=Poor, 2=Fair, 3=Good, 4=Very Good, 5=Outstanding)",
		"comments":         "Additional thoughts or comments about your experience",
	}

	if tooltip, exists := tooltips[fieldName]; exists {
		return tooltip
	}
	return ""
}

// GetRatingOptions returns all available rating options for dropdown menus
// excluding the "Not Rated" option which is used for display purposes only.
func (h *TemplateHelpers) GetRatingOptions() []rating.RatingDescription {
	return rating.GetAllRatingOptions()
}

// SafeHTML marks a string as safe HTML content that bypasses escaping.
// WARNING: This function bypasses HTML escaping and should only be used with trusted content
// to prevent XSS vulnerabilities. Consider using EscapeHTML for user-generated content.
func (h *TemplateHelpers) SafeHTML(s string) template.HTML {
	// #nosec G203 - This is intentionally designed to bypass HTML escaping for trusted content
	return template.HTML(s)
}

// EscapeHTML safely escapes HTML content to prevent XSS attacks
// by converting special characters to their HTML entity equivalents.
func (h *TemplateHelpers) EscapeHTML(s string) string {
	return template.HTMLEscapeString(s)
}

// Add performs integer addition for use in templates where arithmetic
// operations are needed for pagination or counting purposes.
func (h *TemplateHelpers) Add(a, b int) int {
	return a + b
}

// Sub performs integer subtraction for use in templates.
func (h *TemplateHelpers) Sub(a, b int) int {
	return a - b
}

// Mul performs integer multiplication for use in templates.
func (h *TemplateHelpers) Mul(a, b int) int {
	return a * b
}

// Default returns the value if non-zero, otherwise returns the default value.
func (h *TemplateHelpers) Default(defaultVal, val interface{}) interface{} {
	switch v := val.(type) {
	case int:
		if v == 0 {
			return defaultVal
		}
	case float64:
		if v == 0 {
			return defaultVal
		}
	case string:
		if v == "" {
			return defaultVal
		}
	case bool:
		if !v {
			return defaultVal
		}
	case nil:
		return defaultVal
	}
	return val
}

// StringToInt converts a string to an integer, returning 0 if conversion fails
// to provide safe numeric conversion in templates without error handling.
func (h *TemplateHelpers) StringToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// IsEmpty checks if a string is empty or contains only whitespace characters
// for conditional rendering in templates based on content availability.
func (h *TemplateHelpers) IsEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

// FormatDate formats a time.Time as a date string for display
// in a user-friendly format (e.g., "Jan 02 2006").
func (h *TemplateHelpers) FormatDate(t time.Time) string {
	return t.Format("Jan 02 2006")
}

// FormatDateTime formats a time.Time as a date and time string for display
// in a user-friendly format (e.g., "Jan 02 2006 3:04 PM").
func (h *TemplateHelpers) FormatDateTime(t time.Time) string {
	return t.Format("Jan 02 2006 3:04 PM")
}

// CSRFToken returns the CSRF token for the current request context
// to be used in forms for CSRF protection validation.
func (h *TemplateHelpers) CSRFToken(c *gin.Context) string {
	return csrf.Token(c)
}

// CSRFField returns the CSRF form field name used by the CSRF middleware
// for proper form field naming in HTML templates.
func (h *TemplateHelpers) CSRFField() string {
	return csrf.Field()
}

// CSRFInput generates a hidden HTML input field containing the CSRF token
// for easy inclusion in forms without manual token handling.
func (h *TemplateHelpers) CSRFInput(c *gin.Context) template.HTML {
	token := csrf.Token(c)
	fieldName := csrf.Field()
	// #nosec G203 - This is safe because we're using CSRF library functions
	return template.HTML(`<input type="hidden" name="` + fieldName + `" value="` + template.HTMLEscapeString(token) + `">`)
}
