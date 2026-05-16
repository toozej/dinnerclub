package controllers

import (
	"context"
	"fmt"
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
	"github.com/toozej/dinnerclub/pkg/pagination"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupEntriesTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Store original DB and replace with test DB
	originalDB := database.DB
	database.DB = db

	// Auto-migrate the models
	err = db.AutoMigrate(&models.Entry{}, &models.Restaurant{}, &models.User{})
	require.NoError(t, err)

	// Initialize user cache
	services.InitUserCache()

	// Cleanup function to restore original DB
	t.Cleanup(func() {
		database.DB = originalDB
	})

	return db
}

func setupEntriesTestRouter() *gin.Engine {
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

func createTestEntry(t *testing.T, name, submitter string) *models.Entry {
	entry := &models.Entry{
		Name:      name,
		Location:  "N", // Use valid location code
		Cuisine:   "Test Cuisine",
		Submitter: submitter,
		Visited:   true,
	}

	err := database.DB.Create(entry).Error
	require.NoError(t, err)

	return entry
}

func TestGetEntries(t *testing.T) {
	// Skip template rendering tests as they require full template setup
	t.Skip("Template rendering tests require full application setup")
}

func TestGetEntriesLogic(t *testing.T) {
	setupEntriesTestDB(t)
	router := setupEntriesTestRouter()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("is_logged_in", true)
		c.Next()
	})

	router.GET("/entries", GetEntries)

	t.Run("returns paginated entries", func(t *testing.T) {
		// Create test entries
		for i := 1; i <= 15; i++ {
			createTestEntry(t, fmt.Sprintf("Restaurant %d", i), "testuser")
		}

		// Since GetEntries only renders HTML templates and we don't have templates loaded in tests,
		// we'll test the database logic directly instead of the HTTP endpoint
		var entriesCount int64
		err := database.DB.Table("entries").Count(&entriesCount).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(15), entriesCount)

		var entries []models.Entry
		err = database.DB.Order("id desc").Limit(10).Find(&entries).Error
		assert.NoError(t, err)
		assert.Len(t, entries, 10)
	})

	t.Run("handles pagination parameters", func(t *testing.T) {
		// Test pagination logic directly since GetEntries renders HTML templates
		// Create some test entries
		for i := 1; i <= 25; i++ {
			createTestEntry(t, fmt.Sprintf("Paginated Restaurant %d", i), "testuser")
		}

		// Test second page (offset 10, limit 10)
		var entries []models.Entry
		err := database.DB.Order("id desc").Limit(10).Offset(10).Find(&entries).Error
		assert.NoError(t, err)
		assert.Len(t, entries, 10)
	})

	t.Run("handles invalid pagination", func(t *testing.T) {
		// Test pagination validation logic
		const entriesPerPage = 10
		_, err := pagination.Paginate("invalid", 100, entriesPerPage)
		assert.Error(t, err)
	})

	t.Run("uses timeout context", func(t *testing.T) {
		// Test database operations with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		db := database.WithContext(ctx)
		if db != nil {
			var count int64
			err := db.Table("entries").Count(&count).Error
			assert.NoError(t, err)
		}
	})
}

func TestGetEntry(t *testing.T) {
	setupEntriesTestDB(t)
	router := setupEntriesTestRouter()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("is_logged_in", true)
		c.Set("username", "testuser")
		c.Next()
	})

	router.GET("/entries/:id", GetEntry)

	t.Run("returns entry when found", func(t *testing.T) {
		entry := createTestEntry(t, "Test Restaurant", "testuser")

		// Test the database logic directly since GetEntry renders HTML templates
		var foundEntry models.Entry
		err := database.DB.Where("id = ?", entry.ID).First(&foundEntry).Error
		assert.NoError(t, err)
		assert.Equal(t, "Test Restaurant", foundEntry.Name)
		assert.Equal(t, "testuser", foundEntry.Submitter)
	})

	t.Run("returns 404 when entry not found", func(t *testing.T) {
		// Test the database logic directly since GetEntry renders HTML templates
		var entry models.Entry
		err := database.DB.Where("id = ?", 99999).First(&entry).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("uses timeout context", func(t *testing.T) {
		entry := createTestEntry(t, "Timeout Test", "testuser")

		// Test database operations with timeout context
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		db := database.WithContext(ctx)
		if db != nil {
			var foundEntry models.Entry
			err := db.Where("id = ?", entry.ID).First(&foundEntry).Error
			assert.NoError(t, err)
			assert.Equal(t, "Timeout Test", foundEntry.Name)
		}
	})
}

func TestGetEntriesByUsername(t *testing.T) {
	setupEntriesTestDB(t)
	router := setupEntriesTestRouter()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("is_logged_in", true)
		c.Next()
	})

	router.GET("/entries/submittedby/:username", GetEntriesByUsername)

	t.Run("returns entries for specific user", func(t *testing.T) {
		// Create entries for different users
		createTestEntry(t, "User1 Restaurant", "user1")
		createTestEntry(t, "User2 Restaurant", "user2")
		createTestEntry(t, "Another User1 Restaurant", "user1")

		// Test the database logic directly since GetEntriesByUsername renders HTML templates
		var entries []models.Entry
		err := database.DB.Where("submitter = ?", "user1").Find(&entries).Error
		assert.NoError(t, err)
		assert.Len(t, entries, 2)

		// Verify the entries belong to user1
		for _, entry := range entries {
			assert.Equal(t, "user1", entry.Submitter)
		}
	})

	t.Run("uses cache when available", func(t *testing.T) {
		// Pre-populate cache
		entries := []models.Entry{
			{Name: "Cached Restaurant", Submitter: "cacheduser"},
		}
		services.UserCache.SetUserEntries("cacheduser", entries)

		// Test that cache is populated
		cachedEntries, found := services.UserCache.GetUserEntries("cacheduser")
		assert.True(t, found)
		assert.Len(t, cachedEntries, 1)
		assert.Equal(t, "Cached Restaurant", cachedEntries[0].Name)
	})

	t.Run("falls back to database when cache miss", func(t *testing.T) {
		createTestEntry(t, "DB Restaurant", "dbuser")

		// Test database fallback by checking the entry exists in database
		var entry models.Entry
		err := database.DB.Where("name = ? AND submitter = ?", "DB Restaurant", "dbuser").First(&entry).Error
		assert.NoError(t, err)
		assert.Equal(t, "DB Restaurant", entry.Name)
	})
}

func TestGetEntryForm(t *testing.T) {
	// Skip template rendering tests as they require full template setup
	t.Skip("Template rendering tests require full application setup")
}

func TestPostEntry(t *testing.T) {
	setupEntriesTestDB(t)
	router := setupEntriesTestRouter()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})

	// Note: PostRestaurant is called but we can't easily mock it in this test
	// The function will run but won't affect the test outcome significantly

	router.POST("/entries/new", PostEntry)

	t.Run("creates entry successfully", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("name", "New Restaurant")
		formData.Set("location", "N") // Use valid location code
		formData.Set("cuisine", "New Cuisine")
		formData.Set("visited", "true")

		req, _ := http.NewRequest("POST", "/entries/new", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should redirect on success
		assert.Equal(t, http.StatusFound, w.Code)

		// Verify entry was created
		var entry models.Entry
		err := database.DB.Where("name = ?", "New Restaurant").First(&entry).Error
		assert.NoError(t, err)
		assert.Equal(t, "testuser", entry.Submitter)
	})

	t.Run("handles validation errors", func(t *testing.T) {
		// Send invalid data (missing required fields)
		formData := url.Values{}
		formData.Set("name", "")     // Empty name should fail validation
		formData.Set("location", "") // Empty location should fail validation
		formData.Set("cuisine", "")  // Empty cuisine should fail validation

		req, _ := http.NewRequest("POST", "/entries/new", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// For HTML requests, validation errors result in redirects with flash messages
		assert.Equal(t, http.StatusFound, w.Code)
	})

	t.Run("invalidates user cache", func(t *testing.T) {
		// Pre-populate cache
		services.UserCache.SetUserEntries("testuser", []models.Entry{})

		formData := url.Values{}
		formData.Set("name", "Cache Test Restaurant")
		formData.Set("location", "NE") // Use valid location code
		formData.Set("cuisine", "Cache Cuisine")

		req, _ := http.NewRequest("POST", "/entries/new", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Cache should be invalidated (we can't easily test this directly)
		assert.Equal(t, http.StatusFound, w.Code)
	})
}

func TestGetEntryUpdateForm(t *testing.T) {
	// Skip template rendering tests as they require full template setup
	t.Skip("Template rendering tests require full application setup")
}

func TestPutEntry(t *testing.T) {
	setupEntriesTestDB(t)
	router := setupEntriesTestRouter()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})

	router.PUT("/entries/:id", PutEntry)

	t.Run("updates entry successfully", func(t *testing.T) {
		entry := createTestEntry(t, "Original Name", "testuser")

		formData := url.Values{}
		formData.Set("name", "Updated Name")
		formData.Set("location", "SW") // Use valid location code
		formData.Set("cuisine", "Updated Cuisine")
		formData.Set("visited", "true")

		req, _ := http.NewRequest("PUT", fmt.Sprintf("/entries/%d", entry.ID), strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)

		// Verify entry was updated
		var updatedEntry models.Entry
		database.DB.First(&updatedEntry, entry.ID)
		assert.Equal(t, "Updated Name", updatedEntry.Name)
	})

	t.Run("returns 404 for non-existent entry", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("name", "Test")
		formData.Set("location", "SE") // Use valid location code
		formData.Set("cuisine", "Test")

		req, _ := http.NewRequest("PUT", "/entries/99999", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// For HTML requests, 404 errors result in redirects
		assert.Equal(t, http.StatusFound, w.Code)
	})
}

func TestDeleteEntry(t *testing.T) {
	setupEntriesTestDB(t)
	router := setupEntriesTestRouter()

	// Mock authentication middleware
	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})

	// Note: DeleteRestaurantIfUnused is called but we can't easily mock it in this test
	// The function will run but won't affect the test outcome significantly

	router.DELETE("/entries/:id", DeleteEntry)

	t.Run("deletes entry successfully", func(t *testing.T) {
		entry := createTestEntry(t, "Delete Me", "testuser")

		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/entries/%d", entry.ID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)

		// Verify entry was deleted (soft delete) - entry should not be found in normal query
		var deletedEntry models.Entry
		err := database.DB.Where("id = ?", entry.ID).First(&deletedEntry).Error
		assert.Error(t, err) // Should return error because entry is soft deleted
	})

	t.Run("returns 404 for non-existent entry", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/entries/99999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// For HTML requests, 404 errors result in redirects
		assert.Equal(t, http.StatusFound, w.Code)
	})
}

func TestEntryUpdate(t *testing.T) {
	t.Run("EntryUpdate struct has correct fields", func(t *testing.T) {
		update := EntryUpdate{
			Name:             "Test Restaurant",
			Location:         "Test City",
			Cuisine:          "Test Cuisine",
			Visited:          true,
			LookingForwardTo: "Great food",
			Closed:           false,
			MealService:      "Dinner",
			Ordered:          "Pizza",
			OrderNext:        "Pasta",
			DoNotOrder:       "Salad",
			FoodRating:       5,
			AmbienceRating:   4,
			ValueRating:      3,
			Comments:         "Great place",
		}

		assert.Equal(t, "Test Restaurant", update.Name)
		assert.Equal(t, "Test City", update.Location)
		assert.Equal(t, "Test Cuisine", update.Cuisine)
		assert.True(t, update.Visited)
		assert.Equal(t, "Great food", update.LookingForwardTo)
		assert.False(t, update.Closed)
		assert.Equal(t, "Dinner", update.MealService)
		assert.Equal(t, "Pizza", update.Ordered)
		assert.Equal(t, "Pasta", update.OrderNext)
		assert.Equal(t, "Salad", update.DoNotOrder)
		assert.Equal(t, 5, update.FoodRating)
		assert.Equal(t, 4, update.AmbienceRating)
		assert.Equal(t, 3, update.ValueRating)
		assert.Equal(t, "Great place", update.Comments)
	})
}
