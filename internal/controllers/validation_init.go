package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/validation"
)

// InitializeValidation sets up the validation system with the response helper
// to ensure consistent error handling across all controllers.
func InitializeValidation() {
	// Set the response handler for validation errors
	validation.ValidationHelperInstance.SetResponseHandler(handleValidationResponse)
}

// handleValidationResponse handles validation error responses using the controllers' Response helper
// and converts validation errors to the appropriate controller response format.
func handleValidationResponse(c *gin.Context, userMessage string, validationErrors []validation.ValidationErrorResponse) {
	// Convert validation errors to controller format
	var controllerErrors []ValidationErrorResponse
	for _, err := range validationErrors {
		controllerErrors = append(controllerErrors, ValidationErrorResponse{
			Field:   err.Field,
			Message: err.Message,
			Value:   err.Value,
		})
	}
	Response.ValidationError(c, userMessage, controllerErrors)
}
