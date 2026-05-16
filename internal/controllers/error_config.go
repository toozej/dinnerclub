package controllers

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ErrorConfig holds configuration for error handling behavior
type ErrorConfig struct {
	// LogSensitiveData determines if sensitive data should be logged (for debugging)
	LogSensitiveData bool
	// IncludeStackTrace determines if stack traces should be included in error responses
	IncludeStackTrace bool
	// DetailedAPIErrors determines if API responses should include detailed error information
	DetailedAPIErrors bool
}

// DefaultErrorConfig returns the default error handling configuration
// suitable for production environments with security-focused settings.
func DefaultErrorConfig() *ErrorConfig {
	return &ErrorConfig{
		LogSensitiveData:  false, // Never log sensitive data in production
		IncludeStackTrace: false, // Don't include stack traces in production
		DetailedAPIErrors: false, // Don't expose detailed errors in production
	}
}

// DevelopmentErrorConfig returns error configuration suitable for development
// environments with detailed debugging information and verbose logging.
func DevelopmentErrorConfig() *ErrorConfig {
	return &ErrorConfig{
		LogSensitiveData:  true, // Log sensitive data for debugging
		IncludeStackTrace: true, // Include stack traces for debugging
		DetailedAPIErrors: true, // Include detailed error information
	}
}

// SetupErrorHandling configures error handling middleware for the application
// with the specified error configuration and registers global error handlers.
func SetupErrorHandling(r *gin.Engine, config *ErrorConfig) {
	// Set the global error config
	errorConfig = config

	// Add error handling middleware
	r.Use(ErrorHandlingMiddleware())
	r.Use(RecoveryMiddleware())

	log.Info("Error handling middleware configured")
}

// Global error configuration
var errorConfig *ErrorConfig = DefaultErrorConfig()

// GetErrorConfig returns the current error configuration
// or the default configuration if none has been set.
func GetErrorConfig() *ErrorConfig {
	if errorConfig == nil {
		return DefaultErrorConfig()
	}
	return errorConfig
}

// Common error messages for consistent user experience
const (
	ErrMsgGenericInternal     = "An unexpected error occurred. Please try again later."
	ErrMsgDatabaseUnavailable = "Service temporarily unavailable. Please try again later."
	ErrMsgInvalidInput        = "The information you provided is invalid. Please check and try again."
	ErrMsgUnauthorized        = "You are not authorized to perform this action."
	ErrMsgNotFound            = "The requested resource was not found."
	ErrMsgValidationFailed    = "Please check the information you provided and try again."
	ErrMsgRateLimited         = "Too many requests. Please wait a moment and try again."
)

// Error codes for consistent error identification
const (
	ErrCodeInternal       = "INTERNAL_ERROR"
	ErrCodeDatabase       = "DATABASE_ERROR"
	ErrCodeValidation     = "VALIDATION_ERROR"
	ErrCodeAuthentication = "AUTHENTICATION_ERROR"
	ErrCodeAuthorization  = "AUTHORIZATION_ERROR"
	ErrCodeNotFound       = "NOT_FOUND"
	ErrCodeBadRequest     = "BAD_REQUEST"
	ErrCodeConflict       = "CONFLICT"
	ErrCodeRateLimit      = "RATE_LIMIT_EXCEEDED"
)
