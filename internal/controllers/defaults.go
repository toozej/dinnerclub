package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// sets default gin context variables used across the entire site
func SetDefaults(cityCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("citycode", strings.ToUpper(cityCode))
	}
}
