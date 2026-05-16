// Package authentication provides authentication and session management functionality
// for the dinnerclub web application. It handles user login, logout, session management,
// and JWT token validation.
package authentication

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Auth is authentication management struct that handles user sessions and JWT tokens.
type Auth struct {
	Ses         sessions.Store
	AuthService *AuthService
}

var auth *Auth

// user id key
const USER_ID = "userId"

// New creates and initializes a new Auth instance with the provided session store and auth service.
func New(sessionStore sessions.Store, authService *AuthService) *Auth {
	auth = &Auth{
		Ses:         sessionStore,
		AuthService: authService,
	}

	return auth
}

// Resolve returns the global Auth instance that was previously initialized.
func Resolve() *Auth {
	return auth
}

// Login logs the user in by id
func (a *Auth) Login(userId uint, c *gin.Context) error {
	// store the user in the session
	session := sessions.Default(c)
	session.Set(USER_ID, userId)
	return session.Save()
}

// UserID returns authenticated user id
func (a *Auth) UserID(c *gin.Context) (uint, error) {
	session := sessions.Default(c)
	userIDInterface := session.Get(USER_ID)
	if userIDInterface == nil {
		return 0, fmt.Errorf("user not logged in")
	}

	// Convert to uint
	id, err := strconv.ParseInt(fmt.Sprintf("%v", userIDInterface), 10, 64)
	if err != nil {
		return 0, err
	}
	if id < 0 {
		return 0, fmt.Errorf("invalid user ID: negative value")
	}
	userID := uint(id)

	return userID, nil
}

// Logout logs the user out by id
func (a *Auth) Logout(c *gin.Context) error {
	// delete the user id from the session
	session := sessions.Default(c)
	session.Delete(USER_ID)
	return session.Save()
}

// Check checks if a user is logged in
func (a *Auth) Check(c *gin.Context) (bool, error) {
	session := sessions.Default(c)
	userIDInterface := session.Get(USER_ID)

	// if session doesn't have user id return false
	if userIDInterface == nil {
		return false, nil
	}

	sessionUserId, err := strconv.ParseInt(fmt.Sprintf("%v", userIDInterface), 10, 64)
	if err != nil {
		return false, err
	}
	if sessionUserId < 0 {
		return false, fmt.Errorf("invalid session user ID: negative value")
	}

	// extract the token from Authorization header
	token := a.extractTokenFromHeader(c)
	if token == "" {
		// No token provided, rely on session only
		return sessionUserId > 0, nil
	}

	// validate the token and extract user id
	claims, err := a.AuthService.ValidateToken(token)
	if err != nil {
		// Token is invalid, rely on session only
		return sessionUserId > 0, nil
	}

	// if the session's user id matches the token user id he is authenticated
	if uint(sessionUserId) == claims.UserID {
		return true, nil
	}

	return false, nil
}

// extractTokenFromHeader extracts JWT token from Authorization header
func (a *Auth) extractTokenFromHeader(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check if it starts with "Bearer "
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
