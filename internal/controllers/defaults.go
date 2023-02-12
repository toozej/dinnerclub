package controllers

import (
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// sets default gin context variables used across the entire site
func SetDefaults(cityCode string, referralCode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("citycode", strings.ToUpper(cityCode))
		c.Set("referralcode", referralCode)
	}
}

func flashMessage(c *gin.Context, message string) {
	session := sessions.Default(c)
	session.AddFlash(message)
	if err := session.Save(); err != nil {
		log.Printf("error in flashMessage saving session: %s", err)
	}
}

func flashes(c *gin.Context) []interface{} {
	session := sessions.Default(c)
	flashes := session.Flashes()
	if len(flashes) != 0 {
		if err := session.Save(); err != nil {
			log.Printf("error in flashes saving session: %s", err)
		}
	}
	return flashes
}
