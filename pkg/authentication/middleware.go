// Package authentication provides authentication and session management functionality
// for the dinnerclub web application. It handles user login, logout, session management,
// and JWT token validation.
package authentication

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// JWTMiddleware provides JWT-based authentication middleware for Gin routes.
type JWTMiddleware struct {
	authService *AuthService
}

// NewJWTMiddleware creates a new JWT middleware instance with the provided auth service.
func NewJWTMiddleware(authService *AuthService) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
	}
}

// AuthRequired is middleware that requires valid JWT authentication
func (m *JWTMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromRequest(c)
		if token == "" {
			log.Debug("No JWT token provided in request")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "MISSING_TOKEN",
			})
			return
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			log.Debugf("Invalid JWT token: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			return
		}

		// Set user information in context for use by handlers
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_logged_in", true)

		c.Next()
	}
}

// OptionalAuth is middleware that extracts JWT information if present but doesn't require it
func (m *JWTMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromRequest(c)
		if token == "" {
			// No token provided, set default values
			c.Set("is_logged_in", false)
			c.Set("user_id", uint(0))
			c.Set("username", "")
			c.Next()
			return
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			log.Debugf("Invalid JWT token in optional auth: %s", err.Error())
			// Invalid token, set default values
			c.Set("is_logged_in", false)
			c.Set("user_id", uint(0))
			c.Set("username", "")
			c.Next()
			return
		}

		// Valid token, set user information
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_logged_in", true)

		c.Next()
	}
}

// RefreshToken is middleware that handles token refresh requests
func (m *JWTMiddleware) RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := m.extractTokenFromRequest(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "No token provided for refresh",
				"code":  "MISSING_TOKEN",
			})
			return
		}

		newToken, err := m.authService.RefreshToken(token)
		if err != nil {
			log.Debugf("Failed to refresh token: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Cannot refresh invalid or expired token",
				"code":  "REFRESH_FAILED",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":   newToken,
			"message": "Token refreshed successfully",
		})
	}
}

// extractTokenFromRequest extracts JWT token from various sources in the request
func (m *JWTMiddleware) extractTokenFromRequest(c *gin.Context) string {
	// First, try to get token from Authorization header
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token != "" {
			return token
		}
	}

	// Fallback: try to get token from query parameter (for WebSocket or special cases)
	token := c.Query("token")
	if token != "" {
		return token
	}

	// Fallback: try to get token from cookie (for browser-based requests)
	cookie, err := c.Cookie("jwt_token")
	if err == nil && cookie != "" {
		return cookie
	}

	return ""
}

// GetUserID is a helper function to extract user ID from the Gin context.
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	id, ok := userID.(uint)
	return id, ok
}

// GetUsername is a helper function to extract username from the Gin context.
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}

	name, ok := username.(string)
	return name, ok
}

// IsLoggedIn is a helper function to check if user is authenticated in the current context.
func IsLoggedIn(c *gin.Context) bool {
	loggedIn, exists := c.Get("is_logged_in")
	if !exists {
		return false
	}

	isAuth, ok := loggedIn.(bool)
	return ok && isAuth
}
