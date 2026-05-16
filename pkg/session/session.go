// Package session provides session management functionality for the dinnerclub web application
// using gin-contrib/sessions with cookie-based storage.
package session

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// InitSession creates and returns a session middleware using gin-contrib/sessions.
// This replaces the old gocondor session initialization.
func InitSession(sessionSecret string) gin.HandlerFunc {
	store := cookie.NewStore([]byte(sessionSecret))
	return sessions.Sessions("dinnerclub-session", store)
}

// NewStore creates and returns a session store using cookie-based storage.
func NewStore(sessionSecret string) sessions.Store {
	return cookie.NewStore([]byte(sessionSecret))
}

// FlashMessage adds a flash message to the session.
// This function is safe to call even when session middleware is not configured (e.g., in tests).
func FlashMessage(c *gin.Context, message string) {
	// Try to get the session - this will panic if sessions aren't set up
	// We'll handle this with a recover to make it safe for tests
	defer func() {
		if r := recover(); r != nil {
			log.Debug("Session middleware not available, skipping flash message")
		}
	}()

	session := sessions.Default(c)
	session.AddFlash(message)
	if err := session.Save(); err != nil {
		log.Printf("error in FlashMessage saving session: %s", err)
	}
}

// GetFlashes retrieves flash messages from the session.
// This function is safe to call even when session middleware is not configured (e.g., in tests).
func GetFlashes(c *gin.Context) []any {
	// Try to get the session - this will panic if sessions aren't set up
	// We'll handle this with a recover to make it safe for tests
	defer func() {
		if r := recover(); r != nil {
			log.Debug("Session middleware not available, returning empty flashes")
		}
	}()

	session := sessions.Default(c)
	flashes := session.Flashes()
	if len(flashes) != 0 {
		if err := session.Save(); err != nil {
			log.Printf("error in GetFlashes saving session: %s", err)
		}
	}

	// Return empty slice instead of nil to ensure JSON serializes as [] not null
	if flashes == nil {
		return []any{}
	}
	return flashes
}
