// Package helpers provides template helper functions and utilities
// for the dinnerclub web application's HTML templates.
package helpers

import (
	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/pkg/session"
)

// InitializeTemplateHelpers adds template helper functions to gin.H data
// for use in HTML templates with consistent helper function availability.
func InitializeTemplateHelpers(data gin.H) gin.H {
	h := &TemplateHelpers{}

	// Add helper functions to template data
	data["FormatBoolean"] = h.FormatBoolean
	data["FormatRating"] = h.FormatRating
	data["GetRatingDescription"] = h.GetRatingDescription
	data["GetRatingLabel"] = h.GetRatingLabel
	data["UserProfileLink"] = h.UserProfileLink
	data["TooltipText"] = h.TooltipText
	data["GetRatingOptions"] = h.GetRatingOptions
	data["SafeHTML"] = h.SafeHTML
	data["Add"] = h.Add
	data["StringToInt"] = h.StringToInt

	return data
}

// CreateTemplateData creates a gin.H with common template data and helpers
// including user authentication status, city code, flash messages, and additional data.
func CreateTemplateData(c *gin.Context, additionalData gin.H) gin.H {
	data := gin.H{
		"is_logged_in": c.MustGet("is_logged_in").(bool),
		"citycode":     c.MustGet("citycode").(string),
		"messages":     session.GetFlashes(c),
	}

	// Add additional data
	for key, value := range additionalData {
		data[key] = value
	}

	// Add template helpers
	return InitializeTemplateHelpers(data)
}
