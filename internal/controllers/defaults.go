package controllers

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/pkg/session"
)

// SetDefaults creates middleware that sets default Gin context variables
// used across the entire site including city code and referral code.
func SetDefaults(cityCode string, referralCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("citycode", strings.ToUpper(cityCode))
		c.Set("referralcode", referralCode)
	}
}

// flashMessage adds a flash message to the session for display on the next request.
func flashMessage(c *gin.Context, message string) {
	session.FlashMessage(c, message)
}

// Flashes retrieves flash messages from the session for display in templates
// and clears them after retrieval to prevent duplicate display.
func Flashes(c *gin.Context) []any {
	return session.GetFlashes(c)
}
