package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /health
func HealthGet(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

// GET /status
// TODO include database connectivity health
func StatusGet(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
