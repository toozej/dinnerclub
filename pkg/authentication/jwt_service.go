// Package authentication provides authentication and session management functionality
// for the dinnerclub web application. It handles user login, logout, session management,
// and JWT token validation.
package authentication

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService handles JWT token operations including generation, validation, and refresh.
type AuthService struct {
	secretKey []byte
	issuer    string
}

// Claims represents the JWT claims structure containing user information and standard JWT claims.
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// NewAuthService creates a new AuthService instance with the provided secret key and issuer.
func NewAuthService(secretKey string, issuer string) *AuthService {
	return &AuthService{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

// GenerateToken creates a new JWT token for the given user with a 24-hour expiration.
func (a *AuthService) GenerateToken(userID uint, username string) (string, error) {
	// Set token expiration to 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)

	// Create the claims
	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    a.issuer,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with secret key
	tokenString, err := token.SignedString(a.secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims if the token is valid.
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Check if token is valid and extract claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken creates a new token from an existing valid token, extending the expiration time.
func (a *AuthService) RefreshToken(tokenString string) (string, error) {
	// Validate the existing token
	claims, err := a.ValidateToken(tokenString)
	if err != nil {
		return "", fmt.Errorf("cannot refresh invalid token: %w", err)
	}

	// Generate a new token with the same user information
	return a.GenerateToken(claims.UserID, claims.Username)
}

// ExtractUserIDFromToken extracts the user ID from a token string without full validation.
func (a *AuthService) ExtractUserIDFromToken(tokenString string) (uint, error) {
	claims, err := a.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
