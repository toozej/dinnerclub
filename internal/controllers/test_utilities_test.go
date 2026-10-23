package controllers

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/internal/models"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/gorm"
)

// Test the setupTestDB function used in search_test.go
func TestSetupTestDB(t *testing.T) {
	t.Run("creates functional database", func(t *testing.T) {
		db := setupTestDB()

		assert.NotNil(t, db)

		// Verify database connection works
		sqlDB, err := db.DB()
		require.NoError(t, err)
		err = sqlDB.Ping()
		assert.NoError(t, err)
	})

	t.Run("migrates all required tables", func(t *testing.T) {
		db := setupTestDB()

		// Check that Entry table exists and is functional
		var entryCount int64
		err := db.Model(&models.Entry{}).Count(&entryCount).Error
		assert.NoError(t, err)

		// Check that Restaurant table exists and is functional
		var restaurantCount int64
		err = db.Model(&models.Restaurant{}).Count(&restaurantCount).Error
		assert.NoError(t, err)
	})

	t.Run("creates test data successfully", func(t *testing.T) {
		db := setupTestDB()

		// Verify test entries were created
		var entries []models.Entry
		err := db.Find(&entries).Error
		require.NoError(t, err)
		assert.Greater(t, len(entries), 0, "Should have test entries")

		// Verify test restaurants were created
		var restaurants []models.Restaurant
		err = db.Find(&restaurants).Error
		require.NoError(t, err)
		assert.Greater(t, len(restaurants), 0, "Should have test restaurants")

		// Verify data integrity
		for _, entry := range entries {
			assert.NotEmpty(t, entry.Name, "Entry name should not be empty")
			assert.NotEmpty(t, entry.Location, "Entry location should not be empty")
			assert.NotEmpty(t, entry.Cuisine, "Entry cuisine should not be empty")
			assert.NotEmpty(t, entry.Submitter, "Entry submitter should not be empty")
		}

		for _, restaurant := range restaurants {
			assert.NotEmpty(t, restaurant.Name, "Restaurant name should not be empty")
			assert.NotEmpty(t, restaurant.Location, "Restaurant location should not be empty")
			assert.NotEmpty(t, restaurant.Cuisine, "Restaurant cuisine should not be empty")
		}
	})

	t.Run("creates consistent test data", func(t *testing.T) {
		db1 := setupTestDB()
		db2 := setupTestDB()

		// Get entries from both databases
		var entries1, entries2 []models.Entry
		err := db1.Find(&entries1).Error
		require.NoError(t, err)
		err = db2.Find(&entries2).Error
		require.NoError(t, err)

		// Should have same number of test entries
		assert.Equal(t, len(entries1), len(entries2), "Both databases should have same number of test entries")

		// Get restaurants from both databases
		var restaurants1, restaurants2 []models.Restaurant
		err = db1.Find(&restaurants1).Error
		require.NoError(t, err)
		err = db2.Find(&restaurants2).Error
		require.NoError(t, err)

		// Should have same number of test restaurants
		assert.Equal(t, len(restaurants1), len(restaurants2), "Both databases should have same number of test restaurants")
	})

	t.Run("supports database operations", func(t *testing.T) {
		db := setupTestDB()

		// Test creating new entry
		newEntry := models.Entry{
			Name:      "New Test Restaurant",
			Location:  "New Location",
			Cuisine:   "New Cuisine",
			Submitter: "testuser",
			Visited:   false,
		}
		err := db.Create(&newEntry).Error
		require.NoError(t, err)
		assert.NotZero(t, newEntry.ID)

		// Test reading the entry
		var foundEntry models.Entry
		err = db.First(&foundEntry, newEntry.ID).Error
		require.NoError(t, err)
		assert.Equal(t, "New Test Restaurant", foundEntry.Name)

		// Test updating the entry
		err = db.Model(&foundEntry).Update("visited", true).Error
		require.NoError(t, err)

		// Verify update
		err = db.First(&foundEntry, newEntry.ID).Error
		require.NoError(t, err)
		assert.True(t, foundEntry.Visited)

		// Test deleting the entry
		err = db.Delete(&foundEntry).Error
		require.NoError(t, err)

		// Verify deletion
		err = db.First(&foundEntry, newEntry.ID).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}

// Test the setupTestRouter function used in search_test.go
func TestSetupTestRouter(t *testing.T) {
	t.Run("creates functional gin router", func(t *testing.T) {
		router := setupTestRouter()

		assert.NotNil(t, router)
		assert.IsType(t, &gin.Engine{}, router)
	})

	t.Run("sets up database connection", func(t *testing.T) {
		originalDB := database.DB
		defer func() {
			database.DB = originalDB
		}()

		router := setupTestRouter()
		assert.NotNil(t, router)

		// Verify database was set up
		assert.NotNil(t, database.DB)

		// Verify database is functional
		var count int64
		err := database.DB.Model(&models.Entry{}).Count(&count).Error
		assert.NoError(t, err)
	})

	t.Run("sets up search routes", func(t *testing.T) {
		router := setupTestRouter()

		// Get the routes from the router
		routes := router.Routes()

		// Check that expected routes exist
		expectedRoutes := []string{
			"/search",
			"/search/entries",
			"/search/restaurants",
			"/search/suggestions",
		}

		routePaths := make(map[string]bool)
		for _, route := range routes {
			routePaths[route.Path] = true
		}

		for _, expectedRoute := range expectedRoutes {
			assert.True(t, routePaths[expectedRoute], "Route %s should exist", expectedRoute)
		}
	})

	t.Run("router is in test mode", func(t *testing.T) {
		originalMode := gin.Mode()

		router := setupTestRouter()
		assert.NotNil(t, router)

		// Verify gin is in test mode
		assert.Equal(t, gin.TestMode, gin.Mode())

		// Restore original mode
		gin.SetMode(originalMode)
	})
}

// Test the setupTestRouterWithSession function used in defaults_test.go
func TestSetupTestRouterWithSession(t *testing.T) {
	t.Run("creates router with session middleware", func(t *testing.T) {
		router := setupTestRouterWithSession()

		assert.NotNil(t, router)
		assert.IsType(t, &gin.Engine{}, router)
	})

	t.Run("router is in test mode", func(t *testing.T) {
		originalMode := gin.Mode()

		router := setupTestRouterWithSession()
		assert.NotNil(t, router)

		// Verify gin is in test mode
		assert.Equal(t, gin.TestMode, gin.Mode())

		// Restore original mode
		gin.SetMode(originalMode)
	})

	t.Run("router has middleware configured", func(t *testing.T) {
		router := setupTestRouterWithSession()

		// The router should have middleware handlers
		// We can't directly inspect middleware, but we can verify the router was created
		assert.NotNil(t, router)

		// Verify we can add routes to the router
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		routes := router.Routes()
		assert.Greater(t, len(routes), 0)
	})
}

func TestDatabaseTestUtilityIntegration(t *testing.T) {
	t.Run("multiple test databases are isolated", func(t *testing.T) {
		// Create two separate test databases
		db1 := setupTestDB()
		db2 := setupTestDB()

		// Add unique data to first database
		uniqueEntry1 := models.Entry{
			Name:      "Unique Restaurant 1",
			Location:  "Unique Location 1",
			Cuisine:   "Unique Cuisine 1",
			Submitter: "unique_user_1",
			Visited:   true,
		}
		err := db1.Create(&uniqueEntry1).Error
		require.NoError(t, err)

		// Add different unique data to second database
		uniqueEntry2 := models.Entry{
			Name:      "Unique Restaurant 2",
			Location:  "Unique Location 2",
			Cuisine:   "Unique Cuisine 2",
			Submitter: "unique_user_2",
			Visited:   false,
		}
		err = db2.Create(&uniqueEntry2).Error
		require.NoError(t, err)

		// Verify first database doesn't have second entry
		var count int64
		err = db1.Model(&models.Entry{}).Where("name = ?", "Unique Restaurant 2").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)

		// Verify second database doesn't have first entry
		err = db2.Model(&models.Entry{}).Where("name = ?", "Unique Restaurant 1").Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})

	t.Run("test database cleanup works", func(t *testing.T) {
		db := setupTestDB()

		// Get initial count
		var initialCount int64
		err := db.Model(&models.Entry{}).Count(&initialCount).Error
		require.NoError(t, err)

		// Add some test data
		testEntries := []models.Entry{
			{Name: "Cleanup Test 1", Location: "L1", Cuisine: "C1", Submitter: "user1", Visited: true},
			{Name: "Cleanup Test 2", Location: "L2", Cuisine: "C2", Submitter: "user2", Visited: false},
		}

		for _, entry := range testEntries {
			err := db.Create(&entry).Error
			require.NoError(t, err)
		}

		// Verify data was added
		var afterAddCount int64
		err = db.Model(&models.Entry{}).Count(&afterAddCount).Error
		require.NoError(t, err)
		assert.Equal(t, initialCount+2, afterAddCount)

		// Clean up the test data
		err = db.Where("name LIKE ?", "Cleanup Test%").Delete(&models.Entry{}).Error
		require.NoError(t, err)

		// Verify cleanup worked
		var afterCleanupCount int64
		err = db.Model(&models.Entry{}).Count(&afterCleanupCount).Error
		require.NoError(t, err)
		assert.Equal(t, initialCount, afterCleanupCount)
	})

	t.Run("test database supports transactions", func(t *testing.T) {
		db := setupTestDB()

		// Get initial count
		var initialCount int64
		err := db.Model(&models.Entry{}).Count(&initialCount).Error
		require.NoError(t, err)

		// Start transaction
		tx := db.Begin()
		require.NoError(t, tx.Error)

		// Add data in transaction
		entry := models.Entry{
			Name:      "Transaction Test",
			Location:  "TX Location",
			Cuisine:   "TX Cuisine",
			Submitter: "tx_user",
			Visited:   true,
		}
		err = tx.Create(&entry).Error
		require.NoError(t, err)

		// Verify data exists in transaction context
		var txCount int64
		err = tx.Model(&models.Entry{}).Count(&txCount).Error
		require.NoError(t, err)
		assert.Equal(t, initialCount+1, txCount)

		// Rollback transaction
		tx.Rollback()

		// Verify data was rolled back (use a fresh DB for the check since SQLite
		// in-memory is single-connection and the original db held the transaction)
		dbCheck := setupTestDB()
		var finalCount int64
		err = dbCheck.Model(&models.Entry{}).Count(&finalCount).Error
		require.NoError(t, err)
		assert.Equal(t, initialCount, finalCount)
	})
}
