package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestJWTMiddleware_AuthRequired(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test auth service
	authService := NewAuthService("test-secret-key", "test-issuer")
	middleware := NewJWTMiddleware(authService)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "No token provided",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Authentication required",
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "Invalid or expired token",
		},
		{
			name:           "Valid token",
			token:          generateValidToken(t, authService),
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up request with token if provided
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			c.Request = req

			// Create a test handler that sets a flag if reached
			handlerCalled := false
			testHandler := func(c *gin.Context) {
				handlerCalled = true
				c.Status(http.StatusOK)
			}

			// Apply middleware and handler
			middleware.AuthRequired()(c)
			if !c.IsAborted() {
				testHandler(c)
			}

			// Check results
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				assert.True(t, handlerCalled, "Handler should be called for valid token")

				// Check that user info is set in context
				userID, exists := c.Get("user_id")
				assert.True(t, exists, "user_id should be set in context")
				assert.Equal(t, uint(1), userID)

				username, exists := c.Get("username")
				assert.True(t, exists, "username should be set in context")
				assert.Equal(t, "testuser", username)

				isLoggedIn, exists := c.Get("is_logged_in")
				assert.True(t, exists, "is_logged_in should be set in context")
				assert.True(t, isLoggedIn.(bool))
			} else {
				assert.False(t, handlerCalled, "Handler should not be called for invalid/missing token")
			}
		})
	}
}

func TestJWTMiddleware_OptionalAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService := NewAuthService("test-secret-key", "test-issuer")
	middleware := NewJWTMiddleware(authService)

	tests := []struct {
		name             string
		token            string
		expectedLoggedIn bool
		expectedUserID   uint
		expectedUsername string
	}{
		{
			name:             "No token provided",
			token:            "",
			expectedLoggedIn: false,
			expectedUserID:   0,
			expectedUsername: "",
		},
		{
			name:             "Invalid token",
			token:            "invalid-token",
			expectedLoggedIn: false,
			expectedUserID:   0,
			expectedUsername: "",
		},
		{
			name:             "Valid token",
			token:            generateValidToken(t, authService),
			expectedLoggedIn: true,
			expectedUserID:   1,
			expectedUsername: "testuser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up request with token if provided
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			c.Request = req

			// Apply middleware
			middleware.OptionalAuth()(c)

			// Handler should always be called with optional auth
			c.Status(http.StatusOK)

			// Check context values
			isLoggedIn, exists := c.Get("is_logged_in")
			assert.True(t, exists, "is_logged_in should always be set")
			assert.Equal(t, tt.expectedLoggedIn, isLoggedIn.(bool))

			userID, exists := c.Get("user_id")
			assert.True(t, exists, "user_id should always be set")
			assert.Equal(t, tt.expectedUserID, userID.(uint))

			username, exists := c.Get("username")
			assert.True(t, exists, "username should always be set")
			assert.Equal(t, tt.expectedUsername, username.(string))

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestJWTMiddleware_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService := NewAuthService("test-secret-key", "test-issuer")
	middleware := NewJWTMiddleware(authService)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectNewToken bool
	}{
		{
			name:           "No token provided",
			token:          "",
			expectedStatus: http.StatusBadRequest,
			expectNewToken: false,
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectNewToken: false,
		},
		{
			name:           "Valid token",
			token:          generateValidToken(t, authService),
			expectedStatus: http.StatusOK,
			expectNewToken: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			// Set up request with token if provided
			req := httptest.NewRequest("POST", "/auth/refresh", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", "Bearer "+tt.token)
			}
			c.Request = req

			// Apply middleware
			middleware.RefreshToken()(c)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectNewToken {
				// Should return JSON with new token
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
			}
		})
	}
}

func TestExtractTokenFromRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	authService := NewAuthService("test-secret-key", "test-issuer")
	middleware := NewJWTMiddleware(authService)

	tests := []struct {
		name          string
		setupRequest  func(*http.Request, *gin.Context)
		expectedToken string
	}{
		{
			name: "Token from Authorization header",
			setupRequest: func(req *http.Request, c *gin.Context) {
				req.Header.Set("Authorization", "Bearer test-token-123")
			},
			expectedToken: "test-token-123",
		},
		{
			name: "Token from query parameter",
			setupRequest: func(req *http.Request, c *gin.Context) {
				c.Request.URL.RawQuery = "token=query-token-456"
			},
			expectedToken: "query-token-456",
		},
		{
			name: "Token from cookie",
			setupRequest: func(req *http.Request, c *gin.Context) {
				req.AddCookie(&http.Cookie{
					Name:  "jwt_token",
					Value: "cookie-token-789",
				})
			},
			expectedToken: "cookie-token-789",
		},
		{
			name: "No token",
			setupRequest: func(req *http.Request, c *gin.Context) {
				// No token setup
			},
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/test", nil)
			c.Request = req

			tt.setupRequest(req, c)

			token := middleware.extractTokenFromRequest(c)
			assert.Equal(t, tt.expectedToken, token)
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetUserID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Test when user_id is not set
		userID, exists := GetUserID(c)
		assert.False(t, exists)
		assert.Equal(t, uint(0), userID)

		// Test when user_id is set
		c.Set("user_id", uint(123))
		userID, exists = GetUserID(c)
		assert.True(t, exists)
		assert.Equal(t, uint(123), userID)
	})

	t.Run("GetUsername", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Test when username is not set
		username, exists := GetUsername(c)
		assert.False(t, exists)
		assert.Equal(t, "", username)

		// Test when username is set
		c.Set("username", "testuser")
		username, exists = GetUsername(c)
		assert.True(t, exists)
		assert.Equal(t, "testuser", username)
	})

	t.Run("IsLoggedIn", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Test when is_logged_in is not set
		isLoggedIn := IsLoggedIn(c)
		assert.False(t, isLoggedIn)

		// Test when is_logged_in is false
		c.Set("is_logged_in", false)
		isLoggedIn = IsLoggedIn(c)
		assert.False(t, isLoggedIn)

		// Test when is_logged_in is true
		c.Set("is_logged_in", true)
		isLoggedIn = IsLoggedIn(c)
		assert.True(t, isLoggedIn)
	})
}

// Helper function to generate a valid token for testing
func generateValidToken(t *testing.T, authService *AuthService) string {
	token, err := authService.GenerateToken(1, "testuser")
	assert.NoError(t, err)
	return token
}
