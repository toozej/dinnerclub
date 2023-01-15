package session

import (
	"github.com/gin-gonic/gin"
	"github.com/gocondor/core/sessions"
)

// New initiates new session
func InitSession(sessionSecret string) gin.HandlerFunc {
	ses := sessions.New(true)
	return ses.InitiateCookieStore(sessionSecret, "mysession")
}
