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

// setupControllerTestDB creates an in-memory SQLite database for testing
func setupControllerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Store original DB and replace with test DB
	originalDB := database.DB
	database.DB = db

	// Auto-migrate the models
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

func setupControllerTestRouter() *gin.Engine {
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

	// Initialize validation system
	InitializeValidation()

	// Add default middleware
	router.Use(SetDefaults("TEST", "test123"))

	return router
}

func createControllerTestUser(t *testing.T, username, password, email string) *models.User {
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

// Test core controller functionality without template rendering
func TestControllerCore(t *testing.T) {
	setupControllerTestDB(t)
	router := setupControllerTestRouter()

	// Test health endpoint
	router.GET("/health", GetHealth)

	t.Run("health endpoint works", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "ok", w.Body.String())
	})

	// Test status endpoint
	router.GET("/status", GetStatus)

	t.Run("status endpoint works", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/status", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"success":true`)
	})
}

func TestAuthenticationCore(t *testing.T) {
	setupControllerTestDB(t)
	router := setupControllerTestRouter()

	router.POST("/auth/login", PostLogin)
	router.POST("/auth/logout", PostLogout)
	router.POST("/auth/register", PostRegister)

	t.Run("login with valid credentials", func(t *testing.T) {
		createControllerTestUser(t, "testuser", "validpassword123", "test@example.com")

		loginData := LoginCreds{
			Username: "testuser",
			Password: "validpassword123",
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

	t.Run("login fails with invalid credentials", func(t *testing.T) {
		loginData := LoginCreds{
			Username: "nonexistent",
			Password: "wrongpassword123",
		}
		jsonData, _ := json.Marshal(loginData)

		req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("register new user", func(t *testing.T) {
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
	})
}

func TestUserControllerCore(t *testing.T) {
	setupControllerTestDB(t)
	router := setupControllerTestRouter()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			c.Set("user_id", uint(1))
			c.Set("username", "testuser")
		}
		c.Next()
	})

	router.GET("/users/:id", GetUserByID)
	router.GET("/users/username/:username", GetUserByUsername)

	t.Run("get user by ID", func(t *testing.T) {
		createControllerTestUser(t, "testuser1", "password123", "test1@example.com")

		req, _ := http.NewRequest("GET", "/users/1", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"username":"testuser1"`)
	})

	t.Run("get user by username", func(t *testing.T) {
		createControllerTestUser(t, "findme", "password123", "findme@example.com")

		req, _ := http.NewRequest("GET", "/users/username/findme", nil)
		req.Header.Set("Accept", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"username":"findme"`)
	})
}

func TestEntryControllerCore(t *testing.T) {
	setupControllerTestDB(t)
	router := setupControllerTestRouter()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})

	router.POST("/entries/new", PostEntry)

	t.Run("create entry", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("name", "Test Restaurant")
		formData.Set("location", "N") // Use valid location code
		formData.Set("cuisine", "Test Cuisine")
		formData.Set("visited", "true")

		req, _ := http.NewRequest("POST", "/entries/new", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should redirect on success
		assert.Equal(t, http.StatusFound, w.Code)

		// Verify entry was created
		var entry models.Entry
		err := database.DB.Where("name = ?", "Test Restaurant").First(&entry).Error
		assert.NoError(t, err)
		assert.Equal(t, "testuser", entry.Submitter)
	})
}

func TestValidationInit(t *testing.T) {
	t.Run("initializes validation system", func(t *testing.T) {
		assert.NotPanics(t, func() {
			InitializeValidation()
		})
	})
}

func TestDefaults(t *testing.T) {
	router := gin.New()
	store := cookie.NewStore([]byte("test-secret"))
	router.Use(sessions.Sessions("test-session", store))
	router.Use(SetDefaults("pdx", "test123"))

	router.GET("/test", func(c *gin.Context) {
		cityCode := c.MustGet("citycode").(string)
		referralCode := c.MustGet("referralcode").(string)
		c.JSON(http.StatusOK, gin.H{
			"citycode":     cityCode,
			"referralcode": referralCode,
		})
	})

	t.Run("sets defaults correctly", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"citycode":"PDX"`)
		assert.Contains(t, w.Body.String(), `"referralcode":"test123"`)
	})
}

func TestUtilityFunctions(t *testing.T) {
	t.Run("hash password works", func(t *testing.T) {
		password := "testpassword123"
		hashed, err := hashPassword(password)

		assert.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed)

		// Verify the hash can be compared
		err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
		assert.NoError(t, err)
	})

	t.Run("GetCurrentUsername works", func(t *testing.T) {
		setupControllerTestDB(t)
		router := setupControllerTestRouter()

		router.GET("/test-username", func(c *gin.Context) {
			c.Set("username", "jwtuser")
			username := GetCurrentUsername(c)
			c.JSON(http.StatusOK, gin.H{"username": username})
		})

		req, _ := http.NewRequest("GET", "/test-username", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"username":"jwtuser"`)
	})
}

func TestContextTimeout(t *testing.T) {
	setupControllerTestDB(t)

	t.Run("database operations use timeout context", func(t *testing.T) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Test that WithContext works
		db := database.WithContext(ctx)
		assert.NotNil(t, db)

		// Test a simple query with timeout
		var count int64
		err := db.Model(&models.User{}).Count(&count).Error
		assert.NoError(t, err)
	})
}
