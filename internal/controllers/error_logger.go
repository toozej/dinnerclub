package controllers

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ErrorLogger provides structured error logging capabilities
type ErrorLogger struct {
	config *ErrorConfig
}

// NewErrorLogger creates a new error logger with the given configuration
// for structured logging of different types of application errors.
func NewErrorLogger(config *ErrorConfig) *ErrorLogger {
	return &ErrorLogger{
		config: config,
	}
}

// LogDatabaseError logs database-related errors with appropriate context
// including operation details, caller information, and additional metadata.
func (el *ErrorLogger) LogDatabaseError(c *gin.Context, operation string, err error, additionalFields map[string]interface{}) {
	fields := el.getBaseFields(c)
	fields["error_type"] = "database"
	fields["operation"] = operation

	// Add additional fields if provided
	for k, v := range additionalFields {
		fields[k] = v
	}

	// Add caller information
	if caller := el.getCaller(3); caller != "" {
		fields["caller"] = caller
	}

	logEntry := log.WithFields(fields)
	if err != nil {
		logEntry = logEntry.WithError(err)
	}

	logEntry.Error("Database operation failed")
}

// LogValidationError logs validation errors with field-specific information
// including failed field names, error messages, and validation context.
func (el *ErrorLogger) LogValidationError(c *gin.Context, validationErrors []ValidationErrorResponse, additionalFields map[string]interface{}) {
	fields := el.getBaseFields(c)
	fields["error_type"] = "validation"

	// Extract validation details
	var fieldNames []string
	var errorMessages []string
	for _, verr := range validationErrors {
		fieldNames = append(fieldNames, verr.Field)
		errorMessages = append(errorMessages, verr.Message)
	}

	fields["validation_fields"] = fieldNames
	fields["validation_messages"] = errorMessages
	fields["validation_count"] = len(validationErrors)

	// Add additional fields if provided
	for k, v := range additionalFields {
		fields[k] = v
	}

	// Add caller information
	if caller := el.getCaller(3); caller != "" {
		fields["caller"] = caller
	}

	log.WithFields(fields).Warn("Validation failed")
}

// LogAuthenticationError logs authentication-related errors
// with failure reasons and security-conscious error details.
func (el *ErrorLogger) LogAuthenticationError(c *gin.Context, reason string, err error, additionalFields map[string]interface{}) {
	fields := el.getBaseFields(c)
	fields["error_type"] = "authentication"
	fields["auth_failure_reason"] = reason

	// Add additional fields if provided
	for k, v := range additionalFields {
		fields[k] = v
	}

	// Add caller information
	if caller := el.getCaller(3); caller != "" {
		fields["caller"] = caller
	}

	logEntry := log.WithFields(fields)
	if err != nil && el.config.LogSensitiveData {
		// Only log error details in development
		logEntry = logEntry.WithError(err)
	}

	logEntry.Warn("Authentication failed")
}

// LogBusinessLogicError logs application-specific business logic errors
// with component and operation context for debugging application flow issues.
func (el *ErrorLogger) LogBusinessLogicError(c *gin.Context, component string, operation string, err error, additionalFields map[string]interface{}) {
	fields := el.getBaseFields(c)
	fields["error_type"] = "business_logic"
	fields["component"] = component
	fields["operation"] = operation

	// Add additional fields if provided
	for k, v := range additionalFields {
		fields[k] = v
	}

	// Add caller information
	if caller := el.getCaller(3); caller != "" {
		fields["caller"] = caller
	}

	logEntry := log.WithFields(fields)
	if err != nil {
		logEntry = logEntry.WithError(err)
	}

	logEntry.Error("Business logic error")
}

// LogExternalServiceError logs errors related to external service calls
// including service name, operation details, and integration failure context.
func (el *ErrorLogger) LogExternalServiceError(c *gin.Context, service string, operation string, err error, additionalFields map[string]interface{}) {
	fields := el.getBaseFields(c)
	fields["error_type"] = "external_service"
	fields["service"] = service
	fields["operation"] = operation

	// Add additional fields if provided
	for k, v := range additionalFields {
		fields[k] = v
	}

	// Add caller information
	if caller := el.getCaller(3); caller != "" {
		fields["caller"] = caller
	}

	logEntry := log.WithFields(fields)
	if err != nil {
		logEntry = logEntry.WithError(err)
	}

	logEntry.Error("External service error")
}

// LogSecurityEvent logs security-related events and potential threats
// with appropriate severity levels and detailed security context information.
func (el *ErrorLogger) LogSecurityEvent(c *gin.Context, eventType string, severity string, description string, additionalFields map[string]interface{}) {
	fields := el.getBaseFields(c)
	fields["event_type"] = "security"
	fields["security_event"] = eventType
	fields["severity"] = severity
	fields["description"] = description

	// Add additional fields if provided
	for k, v := range additionalFields {
		fields[k] = v
	}

	// Add caller information
	if caller := el.getCaller(3); caller != "" {
		fields["caller"] = caller
	}

	logEntry := log.WithFields(fields)

	// Log at appropriate level based on severity
	switch strings.ToLower(severity) {
	case "critical", "high":
		logEntry.Error("Security event")
	case "medium":
		logEntry.Warn("Security event")
	default:
		logEntry.Info("Security event")
	}
}

// getBaseFields returns common fields for all log entries
func (el *ErrorLogger) getBaseFields(c *gin.Context) log.Fields {
	fields := log.Fields{
		"method":      c.Request.Method,
		"path":        c.Request.URL.Path,
		"remote_addr": c.ClientIP(),
		"user_agent":  c.GetHeader("User-Agent"),
	}

	// Add request ID if available
	if requestID := c.GetString("request_id"); requestID != "" {
		fields["request_id"] = requestID
	}

	// Add user information if available (but be careful about sensitive data)
	if userID, exists := c.Get("user_id"); exists && userID != nil {
		fields["user_id"] = userID
	}

	if username, exists := c.Get("username"); exists && username != "" {
		fields["username"] = username
	}

	// Add query parameters (but filter sensitive ones)
	if len(c.Request.URL.RawQuery) > 0 && el.config.LogSensitiveData {
		fields["query_params"] = c.Request.URL.RawQuery
	}

	return fields
}

// getCaller returns caller information for debugging
func (el *ErrorLogger) getCaller(skip int) string {
	if !el.config.IncludeStackTrace {
		return ""
	}

	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	// Extract just the filename
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		return fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// Global error logger instance
var GlobalErrorLogger = NewErrorLogger(GetErrorConfig())

// UpdateErrorLogger updates the global error logger with new configuration
// to change logging behavior at runtime based on environment or settings.
func UpdateErrorLogger(config *ErrorConfig) {
	GlobalErrorLogger = NewErrorLogger(config)
}
