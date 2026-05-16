// Package validation provides comprehensive validation functionality for the dinnerclub application.
// It includes validation helpers, custom validators, and consistent error handling patterns
// for user input validation across controllers and forms.
package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/toozej/dinnerclub/internal/models"
)

// ValidationHelper provides consistent validation patterns across controllers
type ValidationHelper struct {
	validator       *validator.Validate
	responseHandler func(*gin.Context, string, []ValidationErrorResponse)
}

// NewValidationHelper creates a new validation helper instance with custom validators
// registered for location, rating, and conditional field validation.
func NewValidationHelper() *ValidationHelper {
	v := validator.New()

	// Register custom validators
	if err := v.RegisterValidation("location", validateLocation); err != nil {
		panic(fmt.Sprintf("Failed to register location validator: %v", err))
	}
	if err := v.RegisterValidation("rating", validateRating); err != nil {
		panic(fmt.Sprintf("Failed to register rating validator: %v", err))
	}
	if err := v.RegisterValidation("conditional_required", validateConditionalRequired); err != nil {
		panic(fmt.Sprintf("Failed to register conditional_required validator: %v", err))
	}

	return &ValidationHelper{
		validator: v,
	}
}

// ValidationResult holds validation results
type ValidationResult struct {
	IsValid bool                    `json:"is_valid"`
	Errors  []ValidationErrorDetail `json:"errors,omitempty"`
}

// ValidationErrorDetail provides detailed validation error information
type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
	Message string `json:"message"`
}

// ValidateStruct validates any struct using the validator instance and returns
// a ValidationResult with detailed error information if validation fails.
func (vh *ValidationHelper) ValidateStruct(data interface{}) ValidationResult {
	err := vh.validator.Struct(data)
	if err == nil {
		return ValidationResult{IsValid: true}
	}

	var errors []ValidationErrorDetail
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			errors = append(errors, ValidationErrorDetail{
				Field:   fieldError.Field(),
				Tag:     fieldError.Tag(),
				Value:   fmt.Sprintf("%v", fieldError.Value()),
				Message: vh.getFieldErrorMessage(fieldError),
			})
		}
	}

	return ValidationResult{
		IsValid: false,
		Errors:  errors,
	}
}

// ValidateEntry performs comprehensive entry validation with conditional logic
// based on visited status, including required field validation and rating constraints.
func (vh *ValidationHelper) ValidateEntry(entry *models.Entry) ValidationResult {
	var errors []ValidationErrorDetail

	// Basic required field validation
	if strings.TrimSpace(entry.Name) == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "name",
			Tag:     "required",
			Message: "Restaurant name is required",
		})
	}

	if strings.TrimSpace(entry.Location) == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "location",
			Tag:     "required",
			Message: "Location is required",
		})
	}

	if strings.TrimSpace(entry.Cuisine) == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "cuisine",
			Tag:     "required",
			Message: "Cuisine type is required",
		})
	}

	// Conditional validation based on visited status
	if !entry.Visited {
		if strings.TrimSpace(entry.LookingForwardTo) == "" {
			errors = append(errors, ValidationErrorDetail{
				Field:   "lookingforwardto",
				Tag:     "conditional_required",
				Message: "Please describe what you're looking forward to trying",
			})
		}
	} else {
		// For visited restaurants, validate ratings if provided
		if entry.FoodRating != 0 && !vh.isValidRating(entry.FoodRating) {
			errors = append(errors, ValidationErrorDetail{
				Field:   "foodrating",
				Tag:     "rating",
				Value:   fmt.Sprintf("%d", entry.FoodRating),
				Message: "Food rating must be between 1 and 5",
			})
		}

		if entry.AmbienceRating != 0 && !vh.isValidRating(entry.AmbienceRating) {
			errors = append(errors, ValidationErrorDetail{
				Field:   "ambiencerating",
				Tag:     "rating",
				Value:   fmt.Sprintf("%d", entry.AmbienceRating),
				Message: "Ambience rating must be between 1 and 5",
			})
		}

		if entry.ValueRating != 0 && !vh.isValidRating(entry.ValueRating) {
			errors = append(errors, ValidationErrorDetail{
				Field:   "valuerating",
				Tag:     "rating",
				Value:   fmt.Sprintf("%d", entry.ValueRating),
				Message: "Value rating must be between 1 and 5",
			})
		}
	}

	// Location validation
	if entry.Location != "" && !vh.isValidLocation(entry.Location) {
		errors = append(errors, ValidationErrorDetail{
			Field:   "location",
			Tag:     "location",
			Value:   entry.Location,
			Message: "Location must be one of: N, NE, NW, SW, SE, S",
		})
	}

	// Length validations
	if len(entry.Name) > 100 {
		errors = append(errors, ValidationErrorDetail{
			Field:   "name",
			Tag:     "max",
			Value:   entry.Name,
			Message: "Restaurant name must be 100 characters or less",
		})
	}

	if len(entry.Cuisine) > 50 {
		errors = append(errors, ValidationErrorDetail{
			Field:   "cuisine",
			Tag:     "max",
			Value:   entry.Cuisine,
			Message: "Cuisine type must be 50 characters or less",
		})
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// ValidateLoginCredentials validates login form data including username format
// and password length requirements.
func (vh *ValidationHelper) ValidateLoginCredentials(username, password string) ValidationResult {
	var errors []ValidationErrorDetail

	if strings.TrimSpace(username) == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "username",
			Tag:     "required",
			Message: "Username is required",
		})
	} else if !vh.isValidUsername(username) {
		errors = append(errors, ValidationErrorDetail{
			Field:   "username",
			Tag:     "format",
			Value:   username,
			Message: "Username must contain only letters and numbers",
		})
	}

	if strings.TrimSpace(password) == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "password",
			Tag:     "required",
			Message: "Password is required",
		})
	} else if len(password) < 10 {
		errors = append(errors, ValidationErrorDetail{
			Field:   "password",
			Tag:     "min",
			Message: "Password must be at least 10 characters long",
		})
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// ValidateRegistrationData validates user registration form data including
// username, password, and referral code requirements.
func (vh *ValidationHelper) ValidateRegistrationData(user *models.User) ValidationResult {
	var errors []ValidationErrorDetail

	if strings.TrimSpace(user.Username) == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "username",
			Tag:     "required",
			Message: "Username is required",
		})
	} else if !vh.isValidUsername(user.Username) {
		errors = append(errors, ValidationErrorDetail{
			Field:   "username",
			Tag:     "format",
			Value:   user.Username,
			Message: "Username must contain only letters and numbers",
		})
	}

	if strings.TrimSpace(user.Password) == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "password",
			Tag:     "required",
			Message: "Password is required",
		})
	} else if len(user.Password) < 10 {
		errors = append(errors, ValidationErrorDetail{
			Field:   "password",
			Tag:     "min",
			Message: "Password must be at least 10 characters long",
		})
	}

	if strings.TrimSpace(user.ReferralCode) == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "referralcode",
			Tag:     "required",
			Message: "Referral code is required",
		})
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// ValidateUserUpdate validates user profile update data with optional field
// validation for username, password, and email changes.
func (vh *ValidationHelper) ValidateUserUpdate(username, password, email string) ValidationResult {
	var errors []ValidationErrorDetail

	// Username validation (if provided)
	if username != "" {
		if !vh.isValidUsername(username) {
			errors = append(errors, ValidationErrorDetail{
				Field:   "username",
				Tag:     "format",
				Value:   username,
				Message: "Username must contain only letters and numbers and be lowercase",
			})
		}
	}

	// Password validation (if provided)
	if password != "" && len(password) < 10 {
		errors = append(errors, ValidationErrorDetail{
			Field:   "password",
			Tag:     "min",
			Message: "Password must be at least 10 characters long",
		})
	}

	// Email validation (if provided)
	if email != "" && !vh.isValidEmail(email) {
		errors = append(errors, ValidationErrorDetail{
			Field:   "email",
			Tag:     "email",
			Value:   email,
			Message: "Please enter a valid email address",
		})
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// ValidateSearchParameters validates search query parameters including query text,
// pagination parameters, and sorting options with appropriate constraints.
func (vh *ValidationHelper) ValidateSearchParameters(query, pageStr, limitStr, sortBy, sortOrder string) ValidationResult {
	var errors []ValidationErrorDetail

	// Query validation - require non-empty query
	if strings.TrimSpace(query) == "" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "query",
			Tag:     "required",
			Value:   query,
			Message: "Search query is required",
		})
	} else if len(query) > 200 {
		errors = append(errors, ValidationErrorDetail{
			Field:   "query",
			Tag:     "max",
			Value:   query,
			Message: "Search query must be 200 characters or less",
		})
	}

	// Page validation
	if pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err != nil || page < 1 {
			errors = append(errors, ValidationErrorDetail{
				Field:   "page",
				Tag:     "format",
				Value:   pageStr,
				Message: "Page parameter must be a positive integer",
			})
		}
	}

	// Limit validation
	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err != nil || limit < 1 || limit > 100 {
			errors = append(errors, ValidationErrorDetail{
				Field:   "limit",
				Tag:     "range",
				Value:   limitStr,
				Message: "Limit parameter must be between 1 and 100",
			})
		}
	}

	// Sort field validation (if provided)
	if sortBy != "" {
		validSortFields := map[string]bool{
			"created_at":  true,
			"name":        true,
			"submitter":   true,
			"location":    true,
			"cuisine":     true,
			"entry_count": true,
		}
		if !validSortFields[sortBy] {
			errors = append(errors, ValidationErrorDetail{
				Field:   "sort_field",
				Tag:     "enum",
				Value:   sortBy,
				Message: "Invalid sort field. Valid options: created_at, name, submitter, location, cuisine, entry_count",
			})
		}
	}

	// Sort order validation (if provided)
	if sortOrder != "" && sortOrder != "asc" && sortOrder != "desc" {
		errors = append(errors, ValidationErrorDetail{
			Field:   "sort_order",
			Tag:     "enum",
			Value:   sortOrder,
			Message: "Sort order must be 'asc' or 'desc'",
		})
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// HandleValidationErrors converts validation results to controller response format
// and sends appropriate error responses to the client with user-friendly messages.
func (vh *ValidationHelper) HandleValidationErrors(c *gin.Context, result ValidationResult, userMessage string) {
	if result.IsValid {
		return
	}

	var validationErrors []ValidationErrorResponse
	for _, err := range result.Errors {
		validationErrors = append(validationErrors, ValidationErrorResponse{
			Field:   err.Field,
			Message: err.Message,
			Value:   err.Value,
		})
	}

	// Create a response helper interface to handle validation errors
	// This will be called from controllers that have access to the Response helper
	vh.sendValidationError(c, userMessage, validationErrors)
}

// sendValidationError handles the actual response sending
func (vh *ValidationHelper) sendValidationError(c *gin.Context, userMessage string, validationErrors []ValidationErrorResponse) {
	if vh.responseHandler != nil {
		vh.responseHandler(c, userMessage, validationErrors)
		return
	}

	// Use the global Response helper for consistent error handling
	// This will be available once we import the controllers package
	// For now, use the default implementation
	c.AbortWithStatusJSON(400, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": userMessage,
		},
		"data": validationErrors,
	})
}

// SetResponseHandler allows controllers to set their response handler function
// for consistent error response formatting across the application.
func (vh *ValidationHelper) SetResponseHandler(handler func(*gin.Context, string, []ValidationErrorResponse)) {
	vh.responseHandler = handler
}

// InitializeWithResponseHandler sets up the validation helper with the response handler
// to avoid circular import dependencies between validation and controllers packages.
func (vh *ValidationHelper) InitializeWithResponseHandler() {
	// This will be called from the controllers package to set up the response handler
	// The actual implementation will be in the controllers package to avoid circular imports
}

// Helper methods for validation logic

func (vh *ValidationHelper) isValidLocation(location string) bool {
	validLocations := []string{"N", "NE", "NW", "SW", "SE", "S"}
	for _, loc := range validLocations {
		if location == loc {
			return true
		}
	}
	return false
}

func (vh *ValidationHelper) isValidRating(rating int) bool {
	return rating >= 1 && rating <= 5
}

func (vh *ValidationHelper) isValidUsername(username string) bool {
	// Username should contain only alphanumeric characters and be lowercase
	matched, _ := regexp.MatchString("^[a-z0-9]+$", username)
	return matched
}

func (vh *ValidationHelper) isValidEmail(email string) bool {
	// Basic email validation regex
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	return matched
}

func (vh *ValidationHelper) getFieldErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldError.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", fieldError.Field(), fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", fieldError.Field(), fieldError.Param())
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", fieldError.Field())
	case "lowercase":
		return fmt.Sprintf("%s must be lowercase", fieldError.Field())
	case "location":
		return "Location must be one of: N, NE, NW, SW, SE, S"
	case "rating":
		return "Rating must be between 1 and 5"
	case "conditional_required":
		return fmt.Sprintf("%s is required for this type of entry", fieldError.Field())
	default:
		return fmt.Sprintf("%s is invalid", fieldError.Field())
	}
}

// Custom validator functions

func validateLocation(fl validator.FieldLevel) bool {
	location := fl.Field().String()
	validLocations := []string{"N", "NE", "NW", "SW", "SE", "S"}
	for _, loc := range validLocations {
		if location == loc {
			return true
		}
	}
	return false
}

func validateRating(fl validator.FieldLevel) bool {
	rating := fl.Field().Int()
	return rating >= 1 && rating <= 5
}

func validateConditionalRequired(fl validator.FieldLevel) bool {
	// This is a placeholder - actual conditional logic should be handled in ValidateEntry
	return fl.Field().String() != ""
}

// Global validation helper instance
var ValidationHelperInstance = NewValidationHelper()

// ValidationErrorResponse represents validation error details for controller responses
type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ResponseHelperInterface defines the interface for response handling
type ResponseHelperInterface interface {
	ValidationError(c *gin.Context, userMessage string, validationErrors []ValidationErrorResponse)
}
