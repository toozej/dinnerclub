package controllers

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ErrorHandlingMiddleware adds error handling context and request tracking
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add start time for request tracking
		c.Set("start_time", time.Now().Unix())

		// Add request ID if not already present
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateSimpleRequestID(c)
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Log incoming request
		log.WithFields(log.Fields{
			"request_id":  requestID,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"remote_addr": c.ClientIP(),
			"user_agent":  c.GetHeader("User-Agent"),
		}).Info("Incoming request")

		// Process request
		c.Next()

		// Log request completion
		statusCode := c.Writer.Status()
		logLevel := log.InfoLevel
		if statusCode >= 400 {
			logLevel = log.WarnLevel
		}
		if statusCode >= 500 {
			logLevel = log.ErrorLevel
		}

		log.WithFields(log.Fields{
			"request_id":  requestID,
			"status_code": statusCode,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"remote_addr": c.ClientIP(),
		}).Log(logLevel, "Request completed")
	}
}

// generateSimpleRequestID creates a basic request ID
func generateSimpleRequestID(c *gin.Context) string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), c.ClientIP())
}

// RecoveryMiddleware handles panics and converts them to proper error responses
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := c.GetString("request_id")

		// Log the panic with full context
		log.WithFields(log.Fields{
			"request_id":  requestID,
			"panic":       recovered,
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"remote_addr": c.ClientIP(),
			"user_agent":  c.GetHeader("User-Agent"),
		}).Error("Panic recovered")

		// Return appropriate error response
		Response.InternalError(c, "An unexpected error occurred",
			fmt.Errorf("panic recovered: %v", recovered))
	})
}
