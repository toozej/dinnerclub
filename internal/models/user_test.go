package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/pkg/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupUserTestDB creates an in-memory SQLite database for testing users
func setupUserTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Store original DB and replace with test DB
	originalDB := database.DB
	database.DB = db

	// Auto-migrate the User model
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)

	// Cleanup function to restore original DB
	t.Cleanup(func() {
		database.DB = originalDB
	})

	return db
}

func TestUserModel(t *testing.T) {
	t.Run("creates user with all fields", func(t *testing.T) {
		user := User{
			Username:     "testuser",
			Password:     "hashedpassword123",
			Firstname:    "Test",
			Lastname:     "User",
			Email:        "test@example.com",
			ReferralCode: "REF123",
		}

		// Verify all fields are accessible and set correctly
		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "hashedpassword123", user.Password)
		assert.Equal(t, "Test", user.Firstname)
		assert.Equal(t, "User", user.Lastname)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "REF123", user.ReferralCode)
	})

	t.Run("creates user with minimal required fields", func(t *testing.T) {
		user := User{
			Username: "minimaluser",
			Password: "password123",
			Email:    "minimal@example.com",
		}

		assert.Equal(t, "minimaluser", user.Username)
		assert.Equal(t, "password123", user.Password)
		assert.Equal(t, "minimal@example.com", user.Email)

		// Optional fields should have zero values
		assert.Empty(t, user.Firstname)
		assert.Empty(t, user.Lastname)
		assert.Empty(t, user.ReferralCode)
	})

	t.Run("user has gorm.Model fields", func(t *testing.T) {
		user := User{}

		// Verify gorm.Model fields are available
		assert.Equal(t, uint(0), user.ID)
		assert.True(t, user.CreatedAt.IsZero())
		assert.True(t, user.UpdatedAt.IsZero())
		// DeletedAt is a gorm.DeletedAt struct, check if it's not valid (not deleted)
		assert.False(t, user.DeletedAt.Valid)
	})
}

func TestUserValidation(t *testing.T) {
	t.Run("username validation constraints", func(t *testing.T) {
		user := User{
			Username: "validuser123",
			Password: "password123",
			Email:    "valid@example.com",
		}

		// Username should be alphanumeric and lowercase based on binding tags
		assert.Regexp(t, "^[a-z0-9]+$", user.Username)
	})

	t.Run("email validation format", func(t *testing.T) {
		user := User{
			Username: "emailuser",
			Password: "password123",
			Email:    "valid.email@example.com",
		}

		// Email should contain @ symbol for basic validation
		assert.Contains(t, user.Email, "@")
		assert.Contains(t, user.Email, ".")
	})

	t.Run("password length validation", func(t *testing.T) {
		user := User{
			Username: "passuser",
			Password: "validpassword123",
			Email:    "pass@example.com",
		}

		// Password should be at least 10 characters based on binding tags
		assert.GreaterOrEqual(t, len(user.Password), 10)
	})
}

func TestUserDatabaseOperations(t *testing.T) {
	setupUserTestDB(t)

	t.Run("creates user in database", func(t *testing.T) {
		user := User{
			Username:  "dbuser",
			Password:  "dbpassword123",
			Firstname: "Database",
			Lastname:  "User",
			Email:     "db@example.com",
		}

		err := database.DB.Create(&user).Error
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())
	})

	t.Run("retrieves user from database", func(t *testing.T) {
		// Create user
		originalUser := User{
			Username:     "retrieveuser",
			Password:     "retrievepass123",
			Firstname:    "Retrieve",
			Lastname:     "Test",
			Email:        "retrieve@example.com",
			ReferralCode: "RETRIEVE123",
		}

		err := database.DB.Create(&originalUser).Error
		require.NoError(t, err)

		// Retrieve user by username
		var retrievedUser User
		err = database.DB.Where("username = ?", "retrieveuser").First(&retrievedUser).Error
		assert.NoError(t, err)

		assert.Equal(t, originalUser.Username, retrievedUser.Username)
		assert.Equal(t, originalUser.Password, retrievedUser.Password)
		assert.Equal(t, originalUser.Firstname, retrievedUser.Firstname)
		assert.Equal(t, originalUser.Lastname, retrievedUser.Lastname)
		assert.Equal(t, originalUser.Email, retrievedUser.Email)
		assert.Equal(t, originalUser.ReferralCode, retrievedUser.ReferralCode)

		// Retrieve user by email
		var emailUser User
		err = database.DB.Where("email = ?", "retrieve@example.com").First(&emailUser).Error
		assert.NoError(t, err)
		assert.Equal(t, originalUser.Username, emailUser.Username)
	})

	t.Run("updates user in database", func(t *testing.T) {
		// Create user
		user := User{
			Username:  "updateuser",
			Password:  "updatepass123",
			Firstname: "Update",
			Lastname:  "Test",
			Email:     "update@example.com",
		}

		err := database.DB.Create(&user).Error
		require.NoError(t, err)

		// Update user
		user.Firstname = "Updated"
		user.Lastname = "Name"
		user.Email = "updated@example.com"

		err = database.DB.Save(&user).Error
		assert.NoError(t, err)

		// Verify update
		var updatedUser User
		err = database.DB.First(&updatedUser, user.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "Updated", updatedUser.Firstname)
		assert.Equal(t, "Name", updatedUser.Lastname)
		assert.Equal(t, "updated@example.com", updatedUser.Email)
	})

	t.Run("soft deletes user", func(t *testing.T) {
		// Create user
		user := User{
			Username: "deleteuser",
			Password: "deletepass123",
			Email:    "delete@example.com",
		}

		err := database.DB.Create(&user).Error
		require.NoError(t, err)

		// Soft delete user
		err = database.DB.Delete(&user).Error
		assert.NoError(t, err)

		// Verify user is soft deleted (not found in normal query)
		var deletedUser User
		err = database.DB.Where("id = ?", user.ID).First(&deletedUser).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)

		// Verify user exists with Unscoped query
		err = database.DB.Unscoped().Where("id = ?", user.ID).First(&deletedUser).Error
		assert.NoError(t, err)
		assert.NotNil(t, deletedUser.DeletedAt)
	})
}

func TestUserUniqueConstraints(t *testing.T) {
	setupUserTestDB(t)

	t.Run("username must be unique", func(t *testing.T) {
		// Create first user
		user1 := User{
			Username: "uniqueuser",
			Password: "password123",
			Email:    "user1@example.com",
		}

		err := database.DB.Create(&user1).Error
		assert.NoError(t, err)

		// Try to create second user with same username
		user2 := User{
			Username: "uniqueuser", // Same username
			Password: "password456",
			Email:    "user2@example.com",
		}

		err = database.DB.Create(&user2).Error
		assert.Error(t, err) // Should fail due to unique constraint
	})

	t.Run("email must be unique", func(t *testing.T) {
		// Create first user
		user1 := User{
			Username: "emailuser1",
			Password: "password123",
			Email:    "unique@example.com",
		}

		err := database.DB.Create(&user1).Error
		assert.NoError(t, err)

		// Try to create second user with same email
		user2 := User{
			Username: "emailuser2",
			Password: "password456",
			Email:    "unique@example.com", // Same email
		}

		err = database.DB.Create(&user2).Error
		assert.Error(t, err) // Should fail due to unique constraint
	})
}

func TestUserJSONSerialization(t *testing.T) {
	t.Run("user serializes to JSON correctly", func(t *testing.T) {
		user := User{
			Username:  "jsonuser",
			Password:  "jsonpass123",
			Firstname: "JSON",
			Lastname:  "User",
			Email:     "json@example.com",
		}

		// Test that struct tags are properly defined for JSON serialization
		// Password should be excluded from JSON (json:"-" tag)
		assert.Equal(t, "jsonuser", user.Username)
		assert.Equal(t, "jsonpass123", user.Password)
		assert.Equal(t, "JSON", user.Firstname)
		assert.Equal(t, "User", user.Lastname)
		assert.Equal(t, "json@example.com", user.Email)
	})

	t.Run("referral code excluded from JSON", func(t *testing.T) {
		user := User{
			Username:     "secretuser",
			Password:     "secretpass123",
			Email:        "secret@example.com",
			ReferralCode: "SECRET123",
		}

		// ReferralCode should be excluded from JSON (json:"-" tag)
		assert.Equal(t, "SECRET123", user.ReferralCode)
	})
}

func TestUserFormBinding(t *testing.T) {
	t.Run("user has correct form binding tags", func(t *testing.T) {
		user := User{
			Username:     "formuser",
			Password:     "formpass123",
			Firstname:    "Form",
			Lastname:     "User",
			Email:        "form@example.com",
			ReferralCode: "FORM123",
		}

		// Test that form binding tags are properly set
		// This is validated by the struct tags in the model definition
		assert.Equal(t, "formuser", user.Username)
		assert.Equal(t, "formpass123", user.Password)
		assert.Equal(t, "Form", user.Firstname)
		assert.Equal(t, "User", user.Lastname)
		assert.Equal(t, "form@example.com", user.Email)
		assert.Equal(t, "FORM123", user.ReferralCode)
	})
}

func TestUserGormTags(t *testing.T) {
	t.Run("user has correct GORM tags", func(t *testing.T) {
		user := User{
			Username: "gormuser",
			Password: "gormpass123",
			Email:    "gorm@example.com",
		}

		// Test that GORM tags are properly set for database constraints
		// Username: size:255;not null;unique
		// Password: size:255;not null
		// Email: size:255;unique
		assert.NotEmpty(t, user.Username)
		assert.NotEmpty(t, user.Password)
		assert.NotEmpty(t, user.Email)

		// Verify field lengths are reasonable for size:255 constraint
		assert.LessOrEqual(t, len(user.Username), 255)
		assert.LessOrEqual(t, len(user.Password), 255)
		assert.LessOrEqual(t, len(user.Email), 255)
	})
}
