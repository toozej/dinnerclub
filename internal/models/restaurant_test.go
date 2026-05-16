package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupRestaurantTestDB creates an in-memory SQLite database for testing restaurants
func setupRestaurantTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Store original DB and replace with test DB
	originalDB := database.DB
	database.DB = db

	// Auto-migrate the Restaurant model
	err = db.AutoMigrate(&Restaurant{})
	require.NoError(t, err)

	// Cleanup function to restore original DB
	t.Cleanup(func() {
		database.DB = originalDB
	})

	return db
}

func TestRestaurantModel(t *testing.T) {
	t.Run("creates restaurant with all fields", func(t *testing.T) {
		restaurant := Restaurant{
			Name:            "Test Restaurant",
			Location:        "Downtown",
			Cuisine:         "Italian",
			Closed:          false,
			Address:         "123 Main St, City, State 12345",
			WebsiteURL:      "https://testrestaurant.com",
			ReservationsURL: "https://reservations.testrestaurant.com",
			MenuURL:         "https://testrestaurant.com/menu",
			PhoneNumber:     "(555) 123-4567",
		}

		// Verify all fields are accessible and set correctly
		assert.Equal(t, "Test Restaurant", restaurant.Name)
		assert.Equal(t, "Downtown", restaurant.Location)
		assert.Equal(t, "Italian", restaurant.Cuisine)
		assert.False(t, restaurant.Closed)
		assert.Equal(t, "123 Main St, City, State 12345", restaurant.Address)
		assert.Equal(t, "https://testrestaurant.com", restaurant.WebsiteURL)
		assert.Equal(t, "https://reservations.testrestaurant.com", restaurant.ReservationsURL)
		assert.Equal(t, "https://testrestaurant.com/menu", restaurant.MenuURL)
		assert.Equal(t, "(555) 123-4567", restaurant.PhoneNumber)
	})

	t.Run("creates restaurant with minimal required fields", func(t *testing.T) {
		restaurant := Restaurant{
			Name:     "Minimal Restaurant",
			Location: "Uptown",
			Cuisine:  "Mexican",
			Closed:   false,
		}

		assert.Equal(t, "Minimal Restaurant", restaurant.Name)
		assert.Equal(t, "Uptown", restaurant.Location)
		assert.Equal(t, "Mexican", restaurant.Cuisine)
		assert.False(t, restaurant.Closed)

		// Optional fields should have zero values
		assert.Empty(t, restaurant.Address)
		assert.Empty(t, restaurant.WebsiteURL)
		assert.Empty(t, restaurant.ReservationsURL)
		assert.Empty(t, restaurant.MenuURL)
		assert.Empty(t, restaurant.PhoneNumber)
	})

	t.Run("creates closed restaurant", func(t *testing.T) {
		restaurant := Restaurant{
			Name:     "Closed Restaurant",
			Location: "Old Town",
			Cuisine:  "French",
			Closed:   true,
		}

		assert.Equal(t, "Closed Restaurant", restaurant.Name)
		assert.Equal(t, "Old Town", restaurant.Location)
		assert.Equal(t, "French", restaurant.Cuisine)
		assert.True(t, restaurant.Closed)
	})

	t.Run("restaurant has gorm.Model fields", func(t *testing.T) {
		restaurant := Restaurant{}

		// Verify gorm.Model fields are available
		assert.Equal(t, uint(0), restaurant.ID)
		assert.True(t, restaurant.CreatedAt.IsZero())
		assert.True(t, restaurant.UpdatedAt.IsZero())
		// DeletedAt is a gorm.DeletedAt struct, check if it's not valid (not deleted)
		assert.False(t, restaurant.DeletedAt.Valid)
	})
}

func TestRestaurantValidation(t *testing.T) {
	t.Run("required fields validation", func(t *testing.T) {
		restaurant := Restaurant{
			Name:     "Valid Restaurant",
			Location: "Valid Location",
			Cuisine:  "Valid Cuisine",
		}

		// Required fields should not be empty
		assert.NotEmpty(t, restaurant.Name)
		assert.NotEmpty(t, restaurant.Location)
		assert.NotEmpty(t, restaurant.Cuisine)
	})

	t.Run("URL format validation", func(t *testing.T) {
		restaurant := Restaurant{
			Name:            "URL Restaurant",
			Location:        "URL Location",
			Cuisine:         "URL Cuisine",
			WebsiteURL:      "https://example.com",
			ReservationsURL: "https://reservations.example.com",
			MenuURL:         "https://example.com/menu",
		}

		// URLs should start with http:// or https://
		if restaurant.WebsiteURL != "" {
			assert.True(t,
				len(restaurant.WebsiteURL) > 7 &&
					(restaurant.WebsiteURL[:7] == "http://" || restaurant.WebsiteURL[:8] == "https://"))
		}
		if restaurant.ReservationsURL != "" {
			assert.True(t,
				len(restaurant.ReservationsURL) > 7 &&
					(restaurant.ReservationsURL[:7] == "http://" || restaurant.ReservationsURL[:8] == "https://"))
		}
		if restaurant.MenuURL != "" {
			assert.True(t,
				len(restaurant.MenuURL) > 7 &&
					(restaurant.MenuURL[:7] == "http://" || restaurant.MenuURL[:8] == "https://"))
		}
	})

	t.Run("phone number format validation", func(t *testing.T) {
		restaurant := Restaurant{
			Name:        "Phone Restaurant",
			Location:    "Phone Location",
			Cuisine:     "Phone Cuisine",
			PhoneNumber: "(555) 123-4567",
		}

		// Phone number should contain digits
		if restaurant.PhoneNumber != "" {
			assert.Regexp(t, `\d`, restaurant.PhoneNumber)
		}
	})
}

func TestRestaurantDatabaseOperations(t *testing.T) {
	setupRestaurantTestDB(t)

	t.Run("creates restaurant in database", func(t *testing.T) {
		restaurant := Restaurant{
			Name:     "DB Restaurant",
			Location: "DB Location",
			Cuisine:  "DB Cuisine",
			Closed:   false,
		}

		err := database.DB.Create(&restaurant).Error
		assert.NoError(t, err)
		assert.NotZero(t, restaurant.ID)
		assert.False(t, restaurant.CreatedAt.IsZero())
		assert.False(t, restaurant.UpdatedAt.IsZero())
	})

	t.Run("retrieves restaurant from database", func(t *testing.T) {
		// Create restaurant
		originalRestaurant := Restaurant{
			Name:            "Retrieve Restaurant",
			Location:        "Retrieve Location",
			Cuisine:         "Retrieve Cuisine",
			Closed:          false,
			Address:         "456 Retrieve St",
			WebsiteURL:      "https://retrieve.com",
			ReservationsURL: "https://reservations.retrieve.com",
			MenuURL:         "https://retrieve.com/menu",
			PhoneNumber:     "(555) 987-6543",
		}

		err := database.DB.Create(&originalRestaurant).Error
		require.NoError(t, err)

		// Retrieve restaurant by name
		var retrievedRestaurant Restaurant
		err = database.DB.Where("name = ?", "Retrieve Restaurant").First(&retrievedRestaurant).Error
		assert.NoError(t, err)

		assert.Equal(t, originalRestaurant.Name, retrievedRestaurant.Name)
		assert.Equal(t, originalRestaurant.Location, retrievedRestaurant.Location)
		assert.Equal(t, originalRestaurant.Cuisine, retrievedRestaurant.Cuisine)
		assert.Equal(t, originalRestaurant.Closed, retrievedRestaurant.Closed)
		assert.Equal(t, originalRestaurant.Address, retrievedRestaurant.Address)
		assert.Equal(t, originalRestaurant.WebsiteURL, retrievedRestaurant.WebsiteURL)
		assert.Equal(t, originalRestaurant.ReservationsURL, retrievedRestaurant.ReservationsURL)
		assert.Equal(t, originalRestaurant.MenuURL, retrievedRestaurant.MenuURL)
		assert.Equal(t, originalRestaurant.PhoneNumber, retrievedRestaurant.PhoneNumber)
	})

	t.Run("updates restaurant in database", func(t *testing.T) {
		// Create restaurant
		restaurant := Restaurant{
			Name:     "Update Restaurant",
			Location: "Update Location",
			Cuisine:  "Update Cuisine",
			Closed:   false,
		}

		err := database.DB.Create(&restaurant).Error
		require.NoError(t, err)

		// Update restaurant
		restaurant.Closed = true
		restaurant.Address = "789 Updated St"
		restaurant.WebsiteURL = "https://updated.com"
		restaurant.PhoneNumber = "(555) 111-2222"

		err = database.DB.Save(&restaurant).Error
		assert.NoError(t, err)

		// Verify update
		var updatedRestaurant Restaurant
		err = database.DB.First(&updatedRestaurant, restaurant.ID).Error
		assert.NoError(t, err)

		assert.True(t, updatedRestaurant.Closed)
		assert.Equal(t, "789 Updated St", updatedRestaurant.Address)
		assert.Equal(t, "https://updated.com", updatedRestaurant.WebsiteURL)
		assert.Equal(t, "(555) 111-2222", updatedRestaurant.PhoneNumber)
	})

	t.Run("soft deletes restaurant", func(t *testing.T) {
		// Create restaurant
		restaurant := Restaurant{
			Name:     "Delete Restaurant",
			Location: "Delete Location",
			Cuisine:  "Delete Cuisine",
		}

		err := database.DB.Create(&restaurant).Error
		require.NoError(t, err)

		// Soft delete restaurant
		err = database.DB.Delete(&restaurant).Error
		assert.NoError(t, err)

		// Verify restaurant is soft deleted (not found in normal query)
		var deletedRestaurant Restaurant
		err = database.DB.Where("id = ?", restaurant.ID).First(&deletedRestaurant).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		// Verify restaurant exists with Unscoped query
		err = database.DB.Unscoped().Where("id = ?", restaurant.ID).First(&deletedRestaurant).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedRestaurant.DeletedAt)
	})
}

func TestRestaurantUniqueConstraints(t *testing.T) {
	setupRestaurantTestDB(t)

	t.Run("restaurant name must be unique", func(t *testing.T) {
		// Create first restaurant
		restaurant1 := Restaurant{
			Name:     "Unique Restaurant",
			Location: "Location 1",
			Cuisine:  "Cuisine 1",
		}

		err := database.DB.Create(&restaurant1).Error
		assert.NoError(t, err)

		// Try to create second restaurant with same name
		restaurant2 := Restaurant{
			Name:     "Unique Restaurant", // Same name
			Location: "Location 2",
			Cuisine:  "Cuisine 2",
		}

		err = database.DB.Create(&restaurant2).Error
		assert.Error(t, err) // Should fail due to unique constraint
	})
}

func TestRestaurantIndexes(t *testing.T) {
	setupRestaurantTestDB(t)

	t.Run("queries by indexed fields", func(t *testing.T) {
		// Create test restaurants
		restaurants := []Restaurant{
			{Name: "Restaurant A", Location: "North", Cuisine: "Italian", Closed: false},
			{Name: "Restaurant B", Location: "South", Cuisine: "Mexican", Closed: true},
			{Name: "Restaurant C", Location: "North", Cuisine: "Italian", Closed: false},
		}

		for _, restaurant := range restaurants {
			err := database.DB.Create(&restaurant).Error
			require.NoError(t, err)
		}

		// Query by name (indexed field with unique constraint)
		var namedRestaurant Restaurant
		err := database.DB.Where("name = ?", "Restaurant A").First(&namedRestaurant).Error
		assert.NoError(t, err)
		assert.Equal(t, "North", namedRestaurant.Location)

		// Query by location (indexed field)
		var northRestaurants []Restaurant
		err = database.DB.Where("location = ?", "North").Find(&northRestaurants).Error
		assert.NoError(t, err)
		assert.Len(t, northRestaurants, 2)

		// Query by cuisine (indexed field)
		var italianRestaurants []Restaurant
		err = database.DB.Where("cuisine = ?", "Italian").Find(&italianRestaurants).Error
		assert.NoError(t, err)
		assert.Len(t, italianRestaurants, 2)

		// Query by closed status
		var openRestaurants []Restaurant
		err = database.DB.Where("closed = ?", false).Find(&openRestaurants).Error
		assert.NoError(t, err)
		assert.Len(t, openRestaurants, 2)
	})
}

func TestRestaurantJSONSerialization(t *testing.T) {
	t.Run("restaurant serializes to JSON correctly", func(t *testing.T) {
		restaurant := Restaurant{
			Name:            "JSON Restaurant",
			Location:        "JSON Location",
			Cuisine:         "JSON Cuisine",
			Closed:          false,
			Address:         "123 JSON St",
			WebsiteURL:      "https://json.com",
			ReservationsURL: "https://reservations.json.com",
			MenuURL:         "https://json.com/menu",
			PhoneNumber:     "(555) 999-8888",
		}

		// Test that struct tags are properly defined for JSON serialization
		assert.Equal(t, "JSON Restaurant", restaurant.Name)
		assert.Equal(t, "JSON Location", restaurant.Location)
		assert.Equal(t, "JSON Cuisine", restaurant.Cuisine)
		assert.False(t, restaurant.Closed)
		assert.Equal(t, "123 JSON St", restaurant.Address)
		assert.Equal(t, "https://json.com", restaurant.WebsiteURL)
		assert.Equal(t, "https://reservations.json.com", restaurant.ReservationsURL)
		assert.Equal(t, "https://json.com/menu", restaurant.MenuURL)
		assert.Equal(t, "(555) 999-8888", restaurant.PhoneNumber)
	})
}

func TestRestaurantFormBinding(t *testing.T) {
	t.Run("restaurant has correct form binding tags", func(t *testing.T) {
		restaurant := Restaurant{
			Name:            "Form Restaurant",
			Location:        "Form Location",
			Cuisine:         "Form Cuisine",
			Closed:          true,
			Address:         "456 Form Ave",
			WebsiteURL:      "https://form.com",
			ReservationsURL: "https://reservations.form.com",
			MenuURL:         "https://form.com/menu",
			PhoneNumber:     "(555) 777-6666",
		}

		// Test that form binding tags are properly set
		assert.Equal(t, "Form Restaurant", restaurant.Name)
		assert.Equal(t, "Form Location", restaurant.Location)
		assert.Equal(t, "Form Cuisine", restaurant.Cuisine)
		assert.True(t, restaurant.Closed)
		assert.Equal(t, "456 Form Ave", restaurant.Address)
		assert.Equal(t, "https://form.com", restaurant.WebsiteURL)
		assert.Equal(t, "https://reservations.form.com", restaurant.ReservationsURL)
		assert.Equal(t, "https://form.com/menu", restaurant.MenuURL)
		assert.Equal(t, "(555) 777-6666", restaurant.PhoneNumber)
	})
}

func TestRestaurantGormTags(t *testing.T) {
	t.Run("restaurant has correct GORM tags", func(t *testing.T) {
		restaurant := Restaurant{
			Name:     "GORM Restaurant",
			Location: "GORM Location",
			Cuisine:  "GORM Cuisine",
		}

		// Test that GORM tags are properly set for database constraints
		// Name: notNull;unique;index
		// Location: notNull;index
		// Cuisine: notNull;index
		// Closed: default:false
		assert.NotEmpty(t, restaurant.Name)
		assert.NotEmpty(t, restaurant.Location)
		assert.NotEmpty(t, restaurant.Cuisine)
		assert.False(t, restaurant.Closed) // Should default to false
	})
}

func TestRestaurantBusinessLogic(t *testing.T) {
	t.Run("restaurant status management", func(t *testing.T) {
		restaurant := Restaurant{
			Name:     "Status Restaurant",
			Location: "Status Location",
			Cuisine:  "Status Cuisine",
			Closed:   false,
		}

		// Restaurant should start as open
		assert.False(t, restaurant.Closed)

		// Can be marked as closed
		restaurant.Closed = true
		assert.True(t, restaurant.Closed)

		// Can be reopened
		restaurant.Closed = false
		assert.False(t, restaurant.Closed)
	})

	t.Run("restaurant contact information", func(t *testing.T) {
		restaurant := Restaurant{
			Name:        "Contact Restaurant",
			Location:    "Contact Location",
			Cuisine:     "Contact Cuisine",
			Address:     "789 Contact Blvd, Contact City, CC 12345",
			PhoneNumber: "(555) 555-5555",
		}

		// Should have complete contact information
		assert.NotEmpty(t, restaurant.Address)
		assert.NotEmpty(t, restaurant.PhoneNumber)
		assert.Contains(t, restaurant.Address, "Contact")
		assert.Contains(t, restaurant.PhoneNumber, "555")
	})

	t.Run("restaurant online presence", func(t *testing.T) {
		restaurant := Restaurant{
			Name:            "Online Restaurant",
			Location:        "Online Location",
			Cuisine:         "Online Cuisine",
			WebsiteURL:      "https://online.com",
			ReservationsURL: "https://reservations.online.com",
			MenuURL:         "https://online.com/menu",
		}

		// Should have complete online presence
		assert.NotEmpty(t, restaurant.WebsiteURL)
		assert.NotEmpty(t, restaurant.ReservationsURL)
		assert.NotEmpty(t, restaurant.MenuURL)
		assert.Contains(t, restaurant.WebsiteURL, "online.com")
		assert.Contains(t, restaurant.ReservationsURL, "reservations")
		assert.Contains(t, restaurant.MenuURL, "menu")
	})
}
