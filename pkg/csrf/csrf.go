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
	csrfOptions := []csrf.Option{
		csrf.Secure(secure), // Set to true in production with HTTPS
		csrf.HttpOnly(true),
		csrf.SameSite(csrf.SameSiteStrictMode),
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
func Token(c *gin.Context) string {
	return csrf.Token(c.Request)
}

// Field returns the CSRF form field name used by filippo.io/csrf/gorilla.
func Field() string {
	return "gorilla.csrf.Token"
}
