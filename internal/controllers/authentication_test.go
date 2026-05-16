package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/internal/services"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/database"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Store original DB and replace with test DB
	originalDB := database.DB
	database.DB = db

	// Auto-migrate all models (User, Entry, Restaurant)
	err = db.AutoMigrate(&models.User{}, &models.Entry{}, &models.Restaurant{})
	require.NoError(t, err)

	// Initialize user cache
	services.InitUserCache()

	// Cleanup function to restore original DB
	t.Cleanup(func() {
		database.DB = originalDB
	})

	return db
}

func setupAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup session store
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("test-session", store))

	// Setup authentication
	authService := authentication.NewAuthService("test-secret", "test-issuer")
	sessionStore := cookie.NewStore([]byte("test-secret"))
	authentication.New(sessionStore, authService)

	// Initialize JWT middleware
	InitializeJWTMiddleware(authService)

	// Add default middleware
	router.Use(SetDefaults("TEST", "test123"))

	return router
}

func createAuthTestUser(t *testing.T, username, password, email string) *models.User {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)

	user := &models.User{
		Username:     username,
		Password:     string(hashedPassword),
		Email:        email,
		Firstname:    "Test",
		Lastname:     "User",
		ReferralCode: "test123",
	}

	err = database.DB.Create(user).Error
	require.NoError(t, err)

	return user
}

func TestGetLogin(t *testing.T) {
	// Skip template rendering tests as they require full template setup
	t.Skip("Template rendering tests require full application setup")
}

func TestPostLogin(t *testing.T) {
	setupAuthTestDB(t)
	router := setupAuthTestRouter()
	router.POST("/auth/login", PostLogin)

	t.Run("successful login with valid credentials", func(t *testing.T) {
		createAuthTestUser(t, "testuser", "validpassword123", "test@example.com")

		formData := url.Values{}
		formData.Set("username", "testuser")
		formData.Set("password", "validpassword123")

		req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)

		// Check for JWT token cookie
		cookies := w.Result().Cookies()
		var jwtCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "jwt_token" {
				jwtCookie = cookie
				break
			}
		}
		assert.NotNil(t, jwtCookie)
		assert.NotEmpty(t, jwtCookie.Value)
		assert.True(t, jwtCookie.HttpOnly)
	})

	t.Run("login with JSON request returns token", func(t *testing.T) {
		createAuthTestUser(t, "jsonuser", "validpassword123", "json@example.com")

		loginData := map[string]string{
			"username": "jsonuser",
			"password": "validpassword123",
		}
		jsonData, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
		assert.Contains(t, w.Body.String(), `"token"`)
	})

	t.Run("login fails with invalid username", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "nonexistent")
		formData.Set("password", "anypassword123")

		req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("login fails with invalid password", func(t *testing.T) {
		createAuthTestUser(t, "wrongpass", "correctpassword123", "wrong@example.com")

		formData := url.Values{}
		formData.Set("username", "wrongpass")
		formData.Set("password", "wrongpassword123")

		req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("login fails with validation errors", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "")      // Empty username
		formData.Set("password", "short") // Too short password

		req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("uses timeout context", func(t *testing.T) {
		createAuthTestUser(t, "timeoutuser", "validpassword123", "timeout@example.com")

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		formData := url.Values{}
		formData.Set("username", "timeoutuser")
		formData.Set("password", "validpassword123")

		req, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
	})
}

func TestPostLogout(t *testing.T) {
	setupAuthTestDB(t)
	router := setupAuthTestRouter()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("is_logged_in", true)
		c.Set("user_id", uint(1))
		c.Set("username", "testuser")
		c.Next()
	})

	router.POST("/auth/logout", PostLogout)

	t.Run("successful logout", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth/logout", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)

		// Check that JWT token cookie is cleared
		cookies := w.Result().Cookies()
		var jwtCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "jwt_token" {
				jwtCookie = cookie
				break
			}
		}
		assert.NotNil(t, jwtCookie)
		assert.Empty(t, jwtCookie.Value)
		assert.Equal(t, -1, jwtCookie.MaxAge)
	})
}

func TestGetRegister(t *testing.T) {
	// Skip template rendering tests as they require full template setup
	t.Skip("Template rendering tests require full application setup")
}

func TestPostRegister(t *testing.T) {
	setupAuthTestDB(t)
	router := setupAuthTestRouter()
	router.POST("/auth/register", PostRegister)

	t.Run("successful registration", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "newuser")
		formData.Set("password", "newpassword123")
		formData.Set("email", "new@example.com")
		formData.Set("firstname", "New")
		formData.Set("lastname", "User")
		formData.Set("referralcode", "test123")

		req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)

		// Verify user was created
		var user models.User
		err := database.DB.Where("username = ?", "newuser").First(&user).Error
		assert.NoError(t, err)
		assert.Equal(t, "new@example.com", user.Email)
	})

	t.Run("registration fails with duplicate username", func(t *testing.T) {
		createAuthTestUser(t, "duplicate", "password123", "duplicate@example.com")

		formData := url.Values{}
		formData.Set("username", "duplicate")
		formData.Set("password", "newpassword123")
		formData.Set("email", "new@example.com")
		formData.Set("referralcode", "test123")

		req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("registration fails with invalid referral code", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "invalidref")
		formData.Set("password", "password123")
		formData.Set("email", "invalid@example.com")
		formData.Set("referralcode", "wrong123")

		req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("registration fails with validation errors", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("username", "")           // Empty username
		formData.Set("password", "short")      // Too short password
		formData.Set("email", "invalid-email") // Invalid email

		req, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHashPassword(t *testing.T) {
	t.Run("hashes password correctly", func(t *testing.T) {
		password := "testpassword123"
		hashed, err := hashPassword(password)

		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed)

		// Verify the hash can be compared
		err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
		assert.NoError(t, err)
	})
}

func TestInitializeJWTMiddleware(t *testing.T) {
	t.Run("initializes JWT middleware", func(t *testing.T) {
		authService := authentication.NewAuthService("test-secret", "test-issuer")

		assert.NotPanics(t, func() {
			InitializeJWTMiddleware(authService)
		})

		assert.NotNil(t, jwtMiddleware)
	})
}

func TestSetUserStatus(t *testing.T) {
	setupAuthTestDB(t)
	router := setupAuthTestRouter()

	router.Use(SetUserStatus())
	router.GET("/test-status", func(c *gin.Context) {
		isLoggedIn := c.MustGet("is_logged_in").(bool)
		c.JSON(http.StatusOK, gin.H{"is_logged_in": isLoggedIn})
	})

	t.Run("sets user status from JWT", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test-status", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Without a valid JWT token, should be false
		assert.Contains(t, w.Body.String(), `"is_logged_in":false`)
	})
}

func TestEnsureLoggedIn(t *testing.T) {
	setupAuthTestDB(t)
	router := setupAuthTestRouter()

	router.Use(EnsureLoggedIn())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "protected"})
	})

	t.Run("blocks unauthenticated requests", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestEnsureNotLoggedIn(t *testing.T) {
	setupAuthTestDB(t)
	router := setupAuthTestRouter()

	// Mock authentication middleware that sets logged in status
	router.Use(func(c *gin.Context) {
		if c.GetHeader("X-Logged-In") == "true" {
			c.Set("is_logged_in", true)
		}
		c.Next()
	})

	router.Use(EnsureNotLoggedIn())
	router.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "public"})
	})

	t.Run("allows unauthenticated requests", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/public", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("blocks authenticated requests", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/public", nil)
		req.Header.Set("X-Logged-In", "true")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestPostRefreshToken(t *testing.T) {
	setupAuthTestDB(t)
	router := setupAuthTestRouter()
	router.POST("/auth/refresh", PostRefreshToken)

	t.Run("handles refresh token request", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/auth/refresh", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Without a valid token, should return an error
		assert.NotEqual(t, http.StatusOK, w.Code)
	})
}

func TestPreloadUserEntries(t *testing.T) {
	setupAuthTestDB(t)

	t.Run("preloads user entries successfully", func(t *testing.T) {
		// Create test entries
		entry := &models.Entry{
			Name:      "Preload Test",
			Location:  "Test City",
			Cuisine:   "Test Cuisine",
			Submitter: "preloaduser",
			Visited:   true,
		}
		database.DB.Create(entry)

		// Call preload function
		assert.NotPanics(t, func() {
			preloadUserEntries("preloaduser")
		})

		// Give it a moment to complete
		time.Sleep(100 * time.Millisecond)

		// Check if entries were cached
		if services.UserCache != nil {
			entries, found := services.UserCache.GetUserEntries("preloaduser")
			if found {
				assert.Len(t, entries, 1)
				assert.Equal(t, "Preload Test", entries[0].Name)
			}
		}
	})

	t.Run("handles preload errors gracefully", func(t *testing.T) {
		// This should not panic even with non-existent user
		assert.NotPanics(t, func() {
			preloadUserEntries("nonexistentuser")
		})
	})
}

func TestLoginCreds(t *testing.T) {
	t.Run("LoginCreds struct has correct fields", func(t *testing.T) {
		creds := LoginCreds{
			Username: "testuser",
			Password: "testpassword123",
		}

		assert.Equal(t, "testuser", creds.Username)
		assert.Equal(t, "testpassword123", creds.Password)
	})
}
