package controllers

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// StandardResponse represents a consistent response structure for all API endpoints
type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorInfo provides detailed error information
type ErrorInfo struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// ValidationErrorResponse represents validation error details
type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value,omitempty"`
}

// ResponseHelper provides consistent response handling across controllers
type ResponseHelper struct{}

// NewResponseHelper creates a new response helper instance for consistent
// response handling across all controllers in the application.
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// Success sends a successful response with optional data and message,
// handling both HTML and JSON responses appropriately.
func (r *ResponseHelper) Success(c *gin.Context, data interface{}, message string) {
	// For HTML requests, handle differently based on context
	if r.IsHTMLRequest(c) {
		r.handleHTMLSuccess(c, data, message)
		return
	}

	// For API requests, return JSON
	response := StandardResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	c.JSON(http.StatusOK, response)
}

// Created sends a successful creation response (HTTP 201) with optional data
// and message, handling both HTML and JSON responses appropriately.
func (r *ResponseHelper) Created(c *gin.Context, data interface{}, message string) {
	// For HTML requests, handle differently based on context
	if r.IsHTMLRequest(c) {
		r.handleHTMLSuccess(c, data, message)
		return
	}

	// For API requests, return JSON
	response := StandardResponse{
		Success: true,
		Data:    data,
		Message: message,
	}
	c.JSON(http.StatusCreated, response)
}

// Error sends an error response with appropriate HTTP status code, logging,
// and user-friendly messages for both HTML and JSON responses.
func (r *ResponseHelper) Error(c *gin.Context, statusCode int, errorCode, userMessage string, err error) {
	// Generate request ID for tracking
	requestID := r.generateRequestID(c)

	// Log the actual error for debugging with context
	r.logError(c, statusCode, errorCode, userMessage, err, requestID)

	// For HTML requests, redirect with flash message
	if r.IsHTMLRequest(c) {
		r.handleHTMLError(c, statusCode, userMessage, errorCode)
		return
	}

	// For API requests, return JSON error response
	response := StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:      errorCode,
			Message:   userMessage,
			Details:   r.getErrorDetails(err),
			RequestID: requestID,
			Timestamp: r.getCurrentTimestamp(),
		},
	}
	c.AbortWithStatusJSON(statusCode, response)
}

// ValidationError sends a validation error response (HTTP 400)
func (r *ResponseHelper) ValidationError(c *gin.Context, userMessage string, validationErrors []ValidationErrorResponse) {
	// Generate request ID for tracking
	requestID := r.generateRequestID(c)

	// Log validation errors for monitoring
	r.logValidationError(c, userMessage, validationErrors, requestID)

	// For HTML requests, handle validation errors with flash messages
	if r.IsHTMLRequest(c) {
		var messages []string
		for _, verr := range validationErrors {
			messages = append(messages, verr.Message)
		}
		r.handleHTMLValidationError(c, messages)
		return
	}

	// For API requests, return detailed validation errors
	response := StandardResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:      "VALIDATION_ERROR",
			Message:   userMessage,
			RequestID: requestID,
			Timestamp: r.getCurrentTimestamp(),
		},
		Data: validationErrors,
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, response)
}

// NotFound sends a not found error response (HTTP 404)
func (r *ResponseHelper) NotFound(c *gin.Context, resource string) {
	message := "Record not found"
	if resource != "" {
		message = resource + " not found"
	}

	r.Error(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

// Unauthorized sends an unauthorized error response (HTTP 401)
func (r *ResponseHelper) Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Authentication required"
	}
	r.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// Forbidden sends a forbidden error response (HTTP 403)
func (r *ResponseHelper) Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Access denied"
	}
	r.Error(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

// InternalError sends an internal server error response (HTTP 500)
func (r *ResponseHelper) InternalError(c *gin.Context, message string, err error) {
	if message == "" {
		message = "Internal server error"
	}
	r.Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", message, err)
}

// BadRequest sends a bad request error response (HTTP 400)
func (r *ResponseHelper) BadRequest(c *gin.Context, message string, err error) {
	if message == "" {
		message = "Bad request"
	}
	r.Error(c, http.StatusBadRequest, "BAD_REQUEST", message, err)
}

// Conflict sends a conflict error response (HTTP 409)
func (r *ResponseHelper) Conflict(c *gin.Context, message string) {
	if message == "" {
		message = "Resource conflict"
	}
	r.Error(c, http.StatusConflict, "CONFLICT", message, nil)
}

// RedirectWithMessage redirects with a flash message (for HTML requests)
func (r *ResponseHelper) RedirectWithMessage(c *gin.Context, path, message string) {
	flashMessage(c, message)
	c.Redirect(http.StatusFound, path)
}

// Helper methods

// IsHTMLRequest determines if the request expects an HTML response
func (r *ResponseHelper) IsHTMLRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	contentType := c.GetHeader("Content-Type")

	// Check for explicit JSON request
	if accept == "application/json" || c.Query("format") == "json" {
		return false
	}

	// Check for form submission (typical HTML request)
	if contentType == "application/x-www-form-urlencoded" || contentType == "multipart/form-data" {
		return true
	}

	// Default to HTML for browser requests (no explicit Accept header or general browser Accept)
	return accept == "" || accept == "*/*" ||
		(accept != "application/json" && c.Query("format") != "json")
}

// handleHTMLSuccess handles successful responses for HTML requests
func (r *ResponseHelper) handleHTMLSuccess(c *gin.Context, data interface{}, message string) {
	if message != "" {
		flashMessage(c, message)
	}

	// For HTML requests, we typically redirect or render a template
	// The specific behavior depends on the controller context
	// This is a fallback that should be overridden by specific controllers
	if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" {
		// For modification operations, redirect to a sensible default
		c.Redirect(http.StatusFound, "/")
	}
	// For GET requests, the controller should handle template rendering directly
}

// handleHTMLError handles error responses for HTML requests
func (r *ResponseHelper) handleHTMLError(c *gin.Context, statusCode int, message, errorCode string) {
	flashMessage(c, message)

	// Determine appropriate redirect based on error type and current path
	redirectPath := r.getErrorRedirectPath(c, statusCode, errorCode)
	c.Redirect(http.StatusFound, redirectPath)
}

// handleHTMLValidationError handles validation errors for HTML requests
func (r *ResponseHelper) handleHTMLValidationError(c *gin.Context, messages []string) {
	for _, msg := range messages {
		flashMessage(c, msg)
	}

	// For validation errors, typically redirect back to the form
	referer := c.GetHeader("Referer")
	if referer != "" {
		c.Redirect(http.StatusFound, referer)
	} else {
		c.Redirect(http.StatusFound, "/")
	}
}

// getErrorRedirectPath determines the appropriate redirect path for errors
func (r *ResponseHelper) getErrorRedirectPath(c *gin.Context, statusCode int, errorCode string) string {
	switch statusCode {
	case http.StatusUnauthorized:
		return "/auth/login"
	case http.StatusForbidden:
		return "/"
	case http.StatusNotFound:
		return "/"
	default:
		// For other errors, try to redirect to a sensible location based on current path
		referer := c.GetHeader("Referer")
		if referer != "" {
			return referer
		}
		return "/"
	}
}

// getErrorDetails extracts error details for logging/debugging
func (r *ResponseHelper) getErrorDetails(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// generateRequestID creates a unique request ID for error tracking
func (r *ResponseHelper) generateRequestID(c *gin.Context) string {
	// Try to get existing request ID from context or headers
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}
	if requestID := c.GetString("request_id"); requestID != "" {
		return requestID
	}

	// Generate a simple request ID based on remote address and timestamp
	return fmt.Sprintf("%s-%d", strings.ReplaceAll(c.ClientIP(), ".", ""),
		c.GetInt64("start_time"))
}

// getCurrentTimestamp returns current timestamp in ISO format
func (r *ResponseHelper) getCurrentTimestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

// logError logs errors with comprehensive context information
func (r *ResponseHelper) logError(c *gin.Context, statusCode int, errorCode, userMessage string, err error, requestID string) {
	// Get caller information for better debugging
	_, file, line, ok := runtime.Caller(3) // Go up 3 levels to get the actual controller method
	caller := "unknown"
	if ok {
		// Extract just the filename and line number
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			caller = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}

	// Create structured log entry
	logEntry := log.WithFields(log.Fields{
		"request_id":   requestID,
		"status_code":  statusCode,
		"error_code":   errorCode,
		"user_message": userMessage,
		"method":       c.Request.Method,
		"path":         c.Request.URL.Path,
		"remote_addr":  c.ClientIP(),
		"user_agent":   c.GetHeader("User-Agent"),
		"caller":       caller,
	})

	// Add authenticated user info if available
	if userID, exists := c.Get("user_id"); exists {
		logEntry = logEntry.WithField("user_id", userID)
	}
	if username, exists := c.Get("username"); exists {
		logEntry = logEntry.WithField("username", username)
	}

	// Log at appropriate level based on status code
	switch {
	case statusCode >= 500:
		if err != nil {
			logEntry.WithError(err).Error("Internal server error")
		} else {
			logEntry.Error("Internal server error")
		}
	case statusCode >= 400:
		if err != nil {
			logEntry.WithError(err).Warn("Client error")
		} else {
			logEntry.Warn("Client error")
		}
	default:
		if err != nil {
			logEntry.WithError(err).Info("Request error")
		} else {
			logEntry.Info("Request error")
		}
	}
}

// logValidationError logs validation errors for monitoring
func (r *ResponseHelper) logValidationError(c *gin.Context, userMessage string, validationErrors []ValidationErrorResponse, requestID string) {
	// Extract field names and messages for logging
	var fields []string
	var messages []string
	for _, verr := range validationErrors {
		fields = append(fields, verr.Field)
		messages = append(messages, verr.Message)
	}

	log.WithFields(log.Fields{
		"request_id":          requestID,
		"user_message":        userMessage,
		"validation_fields":   fields,
		"validation_messages": messages,
		"method":              c.Request.Method,
		"path":                c.Request.URL.Path,
		"remote_addr":         c.ClientIP(),
		"user_agent":          c.GetHeader("User-Agent"),
	}).Warn("Validation error")
}

// DatabaseError handles database-specific errors with appropriate logging and user messages
func (r *ResponseHelper) DatabaseError(c *gin.Context, operation string, err error) {
	requestID := r.generateRequestID(c)

	// Log database error with operation context
	log.WithFields(log.Fields{
		"request_id":  requestID,
		"operation":   operation,
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"remote_addr": c.ClientIP(),
	}).WithError(err).Error("Database operation failed")

	// Provide user-friendly message based on operation
	var userMessage string
	switch operation {
	case "create":
		userMessage = "Failed to save data. Please try again."
	case "update":
		userMessage = "Failed to update data. Please try again."
	case "delete":
		userMessage = "Failed to delete data. Please try again."
	case "find", "query":
		userMessage = "Failed to retrieve data. Please try again."
	default:
		userMessage = "Database operation failed. Please try again."
	}

	r.Error(c, http.StatusInternalServerError, "DATABASE_ERROR", userMessage, err)
}

// AuthenticationError handles authentication-specific errors
func (r *ResponseHelper) AuthenticationError(c *gin.Context, reason string, err error) {
	requestID := r.generateRequestID(c)

	// Log authentication error
	log.WithFields(log.Fields{
		"request_id":  requestID,
		"reason":      reason,
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"remote_addr": c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
	}).WithError(err).Warn("Authentication failed")

	// Provide generic message for security
	r.Error(c, http.StatusUnauthorized, "AUTHENTICATION_FAILED", "Authentication failed", nil)
}

// Global response helper instance
var Response = NewResponseHelper()
