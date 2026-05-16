package authentication

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAuthService_GenerateToken(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	userID := uint(123)
	username := "testuser"

	token, err := authService.GenerateToken(userID, username)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestAuthService_ValidateToken(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	userID := uint(123)
	username := "testuser"

	// Generate a token
	token, err := authService.GenerateToken(userID, username)
	assert.NoError(t, err)

	// Validate the token
	claims, err := authService.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "test-issuer", claims.Issuer)
}

func TestAuthService_ValidateInvalidToken(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	// Try to validate an invalid token
	_, err := authService.ValidateToken("invalid-token")
	assert.Error(t, err)
}

func TestAuthService_RefreshToken(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	userID := uint(123)
	username := "testuser"

	// Generate original token
	originalToken, err := authService.GenerateToken(userID, username)
	assert.NoError(t, err)

	// Wait a moment to ensure different timestamps
	time.Sleep(time.Second * 1)

	// Refresh the token
	newToken, err := authService.RefreshToken(originalToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newToken)

	// Validate the new token
	claims, err := authService.ValidateToken(newToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
}

func TestAuthService_ExtractUserIDFromToken(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	userID := uint(456)
	username := "anotheruser"

	// Generate a token
	token, err := authService.GenerateToken(userID, username)
	assert.NoError(t, err)

	// Extract user ID
	extractedUserID, err := authService.ExtractUserIDFromToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, extractedUserID)
}
