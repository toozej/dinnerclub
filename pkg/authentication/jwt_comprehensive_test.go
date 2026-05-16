package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_ComprehensiveTokenGeneration(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	tests := []struct {
		name     string
		userID   uint
		username string
		wantErr  bool
	}{
		{
			name:     "Valid user data",
			userID:   123,
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "Zero user ID",
			userID:   0,
			username: "testuser",
			wantErr:  false, // Should still work
		},
		{
			name:     "Empty username",
			userID:   123,
			username: "",
			wantErr:  false, // Should still work
		},
		{
			name:     "Long username",
			userID:   123,
			username: "very_long_username_that_exceeds_normal_length_limits_but_should_still_work",
			wantErr:  false,
		},
		{
			name:     "Special characters in username",
			userID:   123,
			username: "user@domain.com",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := authService.GenerateToken(tt.userID, tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token can be parsed
				claims, err := authService.ValidateToken(token)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, claims.UserID)
				assert.Equal(t, tt.username, claims.Username)
			}
		})
	}
}

func TestAuthService_TokenValidation(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	// Generate a valid token
	validToken, err := authService.GenerateToken(123, "testuser")
	assert.NoError(t, err)

	tests := []struct {
		name       string
		token      string
		wantErr    bool
		wantClaims *Claims
	}{
		{
			name:    "Valid token",
			token:   validToken,
			wantErr: false,
			wantClaims: &Claims{
				UserID:   123,
				Username: "testuser",
			},
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "Invalid token format",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "Malformed JWT",
			token:   "not.a.jwt",
			wantErr: true,
		},
		{
			name:    "Token with wrong signature",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := authService.ValidateToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				if tt.wantClaims != nil {
					assert.Equal(t, tt.wantClaims.UserID, claims.UserID)
					assert.Equal(t, tt.wantClaims.Username, claims.Username)
				}
			}
		})
	}
}

func TestAuthService_TokenExpiration(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	// Create a token that expires immediately
	claims := &Claims{
		UserID:   123,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    authService.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(authService.secretKey)
	assert.NoError(t, err)

	// Try to validate expired token
	_, err = authService.ValidateToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestAuthService_TokenRefresh(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	// Generate original token
	originalToken, err := authService.GenerateToken(123, "testuser")
	assert.NoError(t, err)

	// Wait a moment to ensure different timestamps
	time.Sleep(1 * time.Second)

	// Refresh token
	newToken, err := authService.RefreshToken(originalToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newToken)
	assert.NotEqual(t, originalToken, newToken)

	// Validate new token
	claims, err := authService.ValidateToken(newToken)
	assert.NoError(t, err)
	assert.Equal(t, uint(123), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)

	// Test refresh with invalid token
	_, err = authService.RefreshToken("invalid.token")
	assert.Error(t, err)
}

func TestAuthService_ExtractUserID(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	tests := []struct {
		name     string
		userID   uint
		username string
	}{
		{"Normal user", 123, "testuser"},
		{"Zero ID", 0, "testuser"},
		{"Large ID", 999999, "testuser"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := authService.GenerateToken(tt.userID, tt.username)
			assert.NoError(t, err)

			extractedID, err := authService.ExtractUserIDFromToken(token)
			assert.NoError(t, err)
			assert.Equal(t, tt.userID, extractedID)
		})
	}

	// Test with invalid token
	_, err := authService.ExtractUserIDFromToken("invalid.token")
	assert.Error(t, err)
}

func TestJWTMiddleware_ComprehensiveAuth(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")
	middleware := NewJWTMiddleware(authService)

	// Generate test token
	token, err := authService.GenerateToken(123, "testuser")
	assert.NoError(t, err)

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
		expectUserID   bool
	}{
		{
			name: "Valid Authorization header",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+token)
			},
			expectedStatus: http.StatusOK,
			expectUserID:   true,
		},
		{
			name: "Valid token in query parameter",
			setupRequest: func(req *http.Request) {
				q := req.URL.Query()
				q.Add("token", token)
				req.URL.RawQuery = q.Encode()
			},
			expectedStatus: http.StatusOK,
			expectUserID:   true,
		},
		{
			name: "Valid token in cookie",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "jwt_token",
					Value: token,
				})
			},
			expectedStatus: http.StatusOK,
			expectUserID:   true,
		},
		{
			name: "No token provided",
			setupRequest: func(req *http.Request) {
				// No token
			},
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name: "Invalid Authorization header format",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "InvalidFormat "+token)
			},
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
		{
			name: "Malformed token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid.token.here")
			},
			expectedStatus: http.StatusUnauthorized,
			expectUserID:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.Use(middleware.AuthRequired())
			router.GET("/protected", func(c *gin.Context) {
				userID, _ := GetUserID(c)
				username, _ := GetUsername(c)
				isLoggedIn := IsLoggedIn(c)

				c.JSON(http.StatusOK, gin.H{
					"user_id":      userID,
					"username":     username,
					"is_logged_in": isLoggedIn,
				})
			})

			req := httptest.NewRequest("GET", "/protected", nil)
			tt.setupRequest(req)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectUserID {
				// Should have successful response with user data
				assert.Contains(t, w.Body.String(), "testuser")
			}
		})
	}
}

func TestJWTMiddleware_OptionalAuthComprehensive(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")
	middleware := NewJWTMiddleware(authService)

	token, err := authService.GenerateToken(123, "testuser")
	assert.NoError(t, err)

	tests := []struct {
		name         string
		setupRequest func(*http.Request)
		expectAuth   bool
	}{
		{
			name: "With valid token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer "+token)
			},
			expectAuth: true,
		},
		{
			name: "Without token",
			setupRequest: func(req *http.Request) {
				// No token
			},
			expectAuth: false,
		},
		{
			name: "With invalid token",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid.token")
			},
			expectAuth: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()

			router.Use(middleware.OptionalAuth())
			router.GET("/optional", func(c *gin.Context) {
				isLoggedIn := IsLoggedIn(c)
				userID, _ := GetUserID(c)
				username, _ := GetUsername(c)

				c.JSON(http.StatusOK, gin.H{
					"is_logged_in": isLoggedIn,
					"user_id":      userID,
					"username":     username,
				})
			})

			req := httptest.NewRequest("GET", "/optional", nil)
			tt.setupRequest(req)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			if tt.expectAuth {
				assert.Contains(t, w.Body.String(), `"is_logged_in":true`)
				assert.Contains(t, w.Body.String(), "testuser")
			} else {
				assert.Contains(t, w.Body.String(), `"is_logged_in":false`)
			}
		})
	}
}

func TestJWTMiddleware_ConcurrentAccess(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")

	// Test concurrent token generation and validation
	const numGoroutines = 10
	const tokensPerGoroutine = 5

	results := make(chan error, numGoroutines*tokensPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < tokensPerGoroutine; j++ {
				userID := uint(goroutineID*100 + j)
				username := "user" + string(rune('A'+goroutineID))

				// Generate token
				token, err := authService.GenerateToken(userID, username)
				if err != nil {
					results <- err
					continue
				}

				// Validate token
				claims, err := authService.ValidateToken(token)
				if err != nil {
					results <- err
					continue
				}

				if claims.UserID != userID || claims.Username != username {
					results <- assert.AnError
					continue
				}

				results <- nil
			}
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines*tokensPerGoroutine; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

func TestAuthService_DifferentSecretKeys(t *testing.T) {
	authService1 := NewAuthService("secret-key-1", "issuer-1")
	authService2 := NewAuthService("secret-key-2", "issuer-2")

	// Generate token with first service
	token, err := authService1.GenerateToken(123, "testuser")
	assert.NoError(t, err)

	// Try to validate with second service (different secret)
	_, err = authService2.ValidateToken(token)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "signature is invalid")

	// Validate with correct service
	claims, err := authService1.ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, uint(123), claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
}

func TestJWTMiddleware_ExtractTokenFromRequestEdgeCases(t *testing.T) {
	authService := NewAuthService("test-secret-key", "test-issuer")
	middleware := NewJWTMiddleware(authService)
	tests := []struct {
		name         string
		setupRequest func(*http.Request)
		expectToken  bool
	}{
		{
			name: "Multiple Authorization headers",
			setupRequest: func(req *http.Request) {
				req.Header.Add("Authorization", "Bearer token1")
				req.Header.Add("Authorization", "Bearer token2")
			},
			expectToken: true, // Should use first one
		},
		{
			name: "Authorization header with extra spaces",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "  Bearer   token123   ")
			},
			expectToken: true,
		},
		{
			name: "Token in both header and query",
			setupRequest: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer header-token")
				q := req.URL.Query()
				q.Add("token", "query-token")
				req.URL.RawQuery = q.Encode()
			},
			expectToken: true, // Should prefer header
		},
		{
			name: "Empty cookie value",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{
					Name:  "jwt_token",
					Value: "",
				})
			},
			expectToken: false,
		},
		{
			name: "Multiple cookies with same name",
			setupRequest: func(req *http.Request) {
				req.AddCookie(&http.Cookie{Name: "jwt_token", Value: "token1"})
				req.AddCookie(&http.Cookie{Name: "jwt_token", Value: "token2"})
			},
			expectToken: true, // Should use first one
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			tt.setupRequest(req)

			// Create a gin context to test token extraction
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			token := middleware.extractTokenFromRequest(c)

			if tt.expectToken {
				assert.NotEmpty(t, token)
			} else {
				assert.Empty(t, token)
			}
		})
	}
}
