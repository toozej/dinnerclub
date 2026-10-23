// Package csrf provides CSRF (Cross-Site Request Forgery) protection middleware
// for the dinnerclub web application using the filippo.io/csrf/gorilla package.
package csrf

import (
	"net/http"

	csrf "filippo.io/csrf/gorilla"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Middleware creates a CSRF protection middleware for Gin using filippo.io/csrf/gorilla.
func Middleware(csrfSecret string, secure bool) gin.HandlerFunc {
	// Create CSRF protection with options
	// Note: Secure, HttpOnly, and SameSite options are deprecated in filippo.io/csrf/gorilla
	// as the library no longer relies on cookies for its CSRF protection mechanism.
	csrfOptions := []csrf.Option{
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Warnf("CSRF token validation failed for %s %s", r.Method, r.URL.Path)
			// Don't write response here - let the application handle it
		})),
	}

	csrfProtection := csrf.Protect([]byte(csrfSecret), csrfOptions...)

	return func(c *gin.Context) {
		csrfProtection(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Request = r

			// Check if CSRF validation failed
			if reason := csrf.FailureReason(r); reason != nil {
				log.Warnf("CSRF token validation failed for %s %s: %s", r.Method, r.URL.Path, reason.Error())
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CSRF token validation failed"})
				return
			}

			c.Next()
		})).ServeHTTP(c.Writer, c.Request)
	}
}

// Token returns the CSRF token for the current request context.
// Note: The filippo.io/csrf/gorilla library no longer relies on tokens,
// so this returns an empty string. CSRF protection is handled by the middleware.
func Token(c *gin.Context) string {
	return ""
}

// Field returns the CSRF form field name used by filippo.io/csrf/gorilla.
func Field() string {
	return "gorilla.csrf.Token"
}
