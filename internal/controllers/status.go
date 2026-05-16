package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetHealth handles GET /health requests and returns service health status
// for monitoring systems with support for both JSON and plain text responses.
func GetHealth(c *gin.Context) {
	// Health endpoint typically returns plain text for monitoring systems
	if c.GetHeader("Accept") == "application/json" {
		Response.Success(c, gin.H{"status": "ok"}, "Service is healthy")
	} else {
		c.String(http.StatusOK, "ok")
	}
}

// GetStatus handles GET /status requests and returns detailed service operational status
// for authenticated users with comprehensive system health information.
// TODO include database connectivity health
func GetStatus(c *gin.Context) {
	Response.Success(c, gin.H{"status": "ok"}, "Service is operational")
}
