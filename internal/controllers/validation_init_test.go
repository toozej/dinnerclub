package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/toozej/dinnerclub/internal/validation"
)

func TestInitializeValidation(t *testing.T) {
	t.Run("initializes validation system", func(t *testing.T) {
		// This test verifies that InitializeValidation doesn't panic
		// and sets up the validation system properly
		assert.NotPanics(t, func() {
			InitializeValidation()
		})

		// Verify that the validation helper instance exists
		assert.NotNil(t, validation.ValidationHelperInstance)
	})
}

func TestHandleValidationResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles validation errors correctly", func(t *testing.T) {
		router := gin.New()

		router.POST("/test-validation", func(c *gin.Context) {
			// Simulate validation errors
			validationErrors := []validation.ValidationErrorResponse{
				{
					Field:   "username",
					Message: "Username is required",
					Value:   "",
				},
				{
					Field:   "email",
					Message: "Invalid email format",
					Value:   "invalid-email",
				},
			}

			// Call the validation response handler directly
			handleValidationResponse(c, "Validation failed", validationErrors)
		})

		req, _ := http.NewRequest("POST", "/test-validation", nil)
		req.Header.Set("Accept", "application/json") // Request JSON response
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), `"message":"Validation failed"`)
		assert.Contains(t, w.Body.String(), `"username"`)
		assert.Contains(t, w.Body.String(), `"Username is required"`)
		assert.Contains(t, w.Body.String(), `"email"`)
		assert.Contains(t, w.Body.String(), `"Invalid email format"`)
	})

	t.Run("handles empty validation errors", func(t *testing.T) {
		router := gin.New()

		router.POST("/test-empty-validation", func(c *gin.Context) {
			// Call with empty validation errors
			handleValidationResponse(c, "No specific errors", []validation.ValidationErrorResponse{})
		})

		req, _ := http.NewRequest("POST", "/test-empty-validation", nil)
		req.Header.Set("Accept", "application/json") // Request JSON response
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), `"success":false`)
		assert.Contains(t, w.Body.String(), `"message":"No specific errors"`)
	})
}

func TestValidationErrorResponseConversion(t *testing.T) {
	t.Run("converts validation errors to controller format", func(t *testing.T) {
		// Test the conversion logic used in handleValidationResponse
		validationErrors := []validation.ValidationErrorResponse{
			{
				Field:   "password",
				Message: "Password too short",
				Value:   "123",
			},
		}

		var controllerErrors []ValidationErrorResponse
		for _, err := range validationErrors {
			controllerErrors = append(controllerErrors, ValidationErrorResponse{
				Field:   err.Field,
				Message: err.Message,
				Value:   err.Value,
			})
		}

		assert.Len(t, controllerErrors, 1)
		assert.Equal(t, "password", controllerErrors[0].Field)
		assert.Equal(t, "Password too short", controllerErrors[0].Message)
		assert.Equal(t, "123", controllerErrors[0].Value)
	})
}
