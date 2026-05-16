// Package services provides unit tests for profile display functionality
// testing specific scenarios and edge cases for user profile management.
package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProfileDisplayUnitTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.Entry{}, &models.UserFollow{}, &models.BackupCode{},
		&models.ProfileChange{}, &models.SocialNotification{}, &models.PasswordReset{})
	require.NoError(t, err)

	return db
}

func TestProfileDisplayService_GetEnhancedUserProfile(t *testing.T) {
	db := setupProfileDisplayUnitTestDB(t)
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	service := NewProfileDisplayService(db, socialService, avatarService)

	// Create test user
	user := models.User{
		Username:  "testuser",
		Password:  "testpassword123",
		Email:     "test@example.com",
		Firstname: "Test",
		Lastname:  "User",
		Bio:       "Test bio",
		IsFounder: true,
	}
	db.Create(&user)

	// Create some test data
	entry := models.Entry{
		Name:      "Test Restaurant",
		Submitter: user.Username,
	}
	db.Create(&entry)

	// Create a referral
	referral := models.User{
		Username:   "referral",
		Password:   "testpassword123",
		Email:      "referral@example.com",
		ReferredBy: &user.ID,
	}
	db.Create(&referral)

	t.Run("should return enhanced profile with correct data", func(t *testing.T) {
		profile, err := service.GetEnhancedUserProfile(user.ID, user.ID)
		require.NoError(t, err)

		assert.Equal(t, user.Username, profile.User.Username)
		assert.Equal(t, user.Firstname, profile.User.Firstname)
		assert.Equal(t, user.Lastname, profile.User.Lastname)
		assert.Equal(t, user.Bio, profile.User.Bio)
		assert.True(t, profile.User.IsFounder)
		assert.Equal(t, 1, profile.EntryCount)
		assert.Equal(t, 1, profile.ReferralCount)
		assert.True(t, profile.CanEdit)
		assert.True(t, profile.IsVisible)
		assert.False(t, profile.ShowFollowButton)
	})

	t.Run("should handle nonexistent user", func(t *testing.T) {
		_, err := service.GetEnhancedUserProfile(999, user.ID)
		assert.Error(t, err)
	})
}

func TestProfileDisplayService_GetUserProfileByUsername(t *testing.T) {
	db := setupProfileDisplayUnitTestDB(t)
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	service := NewProfileDisplayService(db, socialService, avatarService)

	// Create test user
	user := models.User{
		Username: "testuser",
		Password: "testpassword123",
		Email:    "test@example.com",
	}
	db.Create(&user)

	t.Run("should return profile by username", func(t *testing.T) {
		profile, err := service.GetUserProfileByUsername("testuser", user.ID)
		require.NoError(t, err)
		assert.Equal(t, "testuser", profile.User.Username)
	})

	t.Run("should handle nonexistent username", func(t *testing.T) {
		_, err := service.GetUserProfileByUsername("nonexistent", user.ID)
		assert.Error(t, err)
	})
}

func TestProfileDisplayService_GetAllUserEntries(t *testing.T) {
	db := setupProfileDisplayUnitTestDB(t)
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	service := NewProfileDisplayService(db, socialService, avatarService)

	// Create test user
	user := models.User{
		Username: "testuser",
		Password: "testpassword123",
		Email:    "test@example.com",
	}
	db.Create(&user)

	// Create test entries
	for i := 0; i < 15; i++ {
		entry := models.Entry{
			Name:      "Restaurant " + string(rune(i)),
			Submitter: user.Username,
		}
		db.Create(&entry)
	}

	t.Run("should return paginated entries", func(t *testing.T) {
		entries, total, err := service.GetAllUserEntries(user.ID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(15), total)
		assert.Len(t, entries, 10)
	})

	t.Run("should handle pagination offset", func(t *testing.T) {
		entries, total, err := service.GetAllUserEntries(user.ID, 10, 10)
		require.NoError(t, err)
		assert.Equal(t, int64(15), total)
		assert.Len(t, entries, 5)
	})

	t.Run("should handle nonexistent user", func(t *testing.T) {
		_, _, err := service.GetAllUserEntries(999, 10, 0)
		assert.Error(t, err)
	})
}

func TestProfileDisplayService_UpdatePrivacySettings(t *testing.T) {
	db := setupProfileDisplayUnitTestDB(t)
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	service := NewProfileDisplayService(db, socialService, avatarService)

	// Create test user
	user := models.User{
		Username:     "testuser",
		Password:     "testpassword123",
		Email:        "test@example.com",
		PrivacyLevel: 0,
	}
	db.Create(&user)

	t.Run("should update privacy level successfully", func(t *testing.T) {
		err := service.UpdatePrivacySettings(user.ID, 2)
		assert.NoError(t, err)

		// Verify update
		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.Equal(t, 2, updatedUser.PrivacyLevel)
	})

	t.Run("should reject invalid privacy levels", func(t *testing.T) {
		err := service.UpdatePrivacySettings(user.ID, -1)
		assert.Error(t, err)

		err = service.UpdatePrivacySettings(user.ID, 3)
		assert.Error(t, err)
	})

	t.Run("should handle nonexistent user", func(t *testing.T) {
		err := service.UpdatePrivacySettings(999, 1)
		assert.Error(t, err)
	})
}

func TestProfileDisplayService_PrivacyVisibility(t *testing.T) {
	db := setupProfileDisplayUnitTestDB(t)
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	service := NewProfileDisplayService(db, socialService, avatarService)

	// Create users
	privateUser := models.User{
		Username:     "privateuser",
		Password:     "testpassword123",
		Email:        "private@example.com",
		PrivacyLevel: 2, // Private
	}
	db.Create(&privateUser)

	followersOnlyUser := models.User{
		Username:     "followersonly",
		Password:     "testpassword123",
		Email:        "followers@example.com",
		PrivacyLevel: 1, // Followers only
	}
	db.Create(&followersOnlyUser)

	publicUser := models.User{
		Username:     "publicuser",
		Password:     "testpassword123",
		Email:        "public@example.com",
		PrivacyLevel: 0, // Public
	}
	db.Create(&publicUser)

	viewer := models.User{
		Username: "viewer",
		Password: "testpassword123",
		Email:    "viewer@example.com",
	}
	db.Create(&viewer)

	t.Run("private user should not be visible to others", func(t *testing.T) {
		profile, err := service.GetEnhancedUserProfile(privateUser.ID, viewer.ID)
		require.NoError(t, err)
		assert.False(t, profile.IsVisible)
		assert.False(t, profile.ShowFollowButton)
		assert.False(t, profile.ShowContactInfo)
	})

	t.Run("public user should be visible to everyone", func(t *testing.T) {
		profile, err := service.GetEnhancedUserProfile(publicUser.ID, viewer.ID)
		require.NoError(t, err)
		assert.True(t, profile.IsVisible)
		assert.True(t, profile.ShowFollowButton)
		assert.False(t, profile.ShowContactInfo)
	})

	t.Run("followers only user should not be visible to non-followers", func(t *testing.T) {
		profile, err := service.GetEnhancedUserProfile(followersOnlyUser.ID, viewer.ID)
		require.NoError(t, err)
		assert.False(t, profile.IsVisible)
		assert.True(t, profile.ShowFollowButton)
		assert.False(t, profile.ShowContactInfo)
	})

	t.Run("followers only user should be visible to followers", func(t *testing.T) {
		// Create follow relationship
		follow := models.UserFollow{
			FollowerID:  viewer.ID,
			FollowingID: followersOnlyUser.ID,
		}
		db.Create(&follow)

		profile, err := service.GetEnhancedUserProfile(followersOnlyUser.ID, viewer.ID)
		require.NoError(t, err)
		assert.True(t, profile.IsVisible)
		assert.True(t, profile.ShowFollowButton)
		assert.True(t, profile.ShowContactInfo)
	})

	t.Run("user should always see their own profile", func(t *testing.T) {
		profile, err := service.GetEnhancedUserProfile(privateUser.ID, privateUser.ID)
		require.NoError(t, err)
		assert.True(t, profile.IsVisible)
		assert.False(t, profile.ShowFollowButton)
		assert.True(t, profile.ShowContactInfo)
		assert.True(t, profile.CanEdit)
	})
}

func TestProfileDisplayService_GetProfileStatistics(t *testing.T) {
	db := setupProfileDisplayUnitTestDB(t)
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	service := NewProfileDisplayService(db, socialService, avatarService)

	// Create test user
	user := models.User{
		Username:  "testuser",
		Password:  "testpassword123",
		Email:     "test@example.com",
		IsFounder: true,
	}
	db.Create(&user)

	// Create test data
	entry := models.Entry{
		Name:      "Test Restaurant",
		Submitter: user.Username,
	}
	db.Create(&entry)

	referral := models.User{
		Username:   "referral",
		Password:   "testpassword123",
		Email:      "referral@example.com",
		ReferredBy: &user.ID,
	}
	db.Create(&referral)

	t.Run("should return comprehensive statistics", func(t *testing.T) {
		stats, err := service.GetProfileStatistics(user.ID)
		require.NoError(t, err)

		assert.Equal(t, 1, stats["entries"])
		assert.Equal(t, 0, stats["followers"])
		assert.Equal(t, 0, stats["following"])
		assert.Equal(t, 1, stats["referrals"])
		assert.Equal(t, 0, stats["total_likes"])
		assert.True(t, stats["is_founder"].(bool))
		assert.NotNil(t, stats["join_date"])
	})

	t.Run("should handle nonexistent user", func(t *testing.T) {
		_, err := service.GetProfileStatistics(999)
		assert.Error(t, err)
	})
}

func TestProfileDisplayService_GetUserEntriesByCategory(t *testing.T) {
	db := setupProfileDisplayUnitTestDB(t)
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	service := NewProfileDisplayService(db, socialService, avatarService)

	// Create test user
	user := models.User{
		Username: "testuser",
		Password: "testpassword123",
		Email:    "test@example.com",
	}
	db.Create(&user)

	// Create test entries with different categories
	entries := []models.Entry{
		{Name: "Italian Place", Submitter: user.Username, Cuisine: "Italian", Location: "Downtown", FoodRating: 5},
		{Name: "Pizza Joint", Submitter: user.Username, Cuisine: "Italian", Location: "Uptown", FoodRating: 4},
		{Name: "Sushi Bar", Submitter: user.Username, Cuisine: "Japanese", Location: "Downtown", FoodRating: 5},
		{Name: "Taco Shop", Submitter: user.Username, Cuisine: "Mexican", Location: "Downtown", FoodRating: 3},
	}

	for _, entry := range entries {
		db.Create(&entry)
	}

	t.Run("should filter by cuisine", func(t *testing.T) {
		entries, total, err := service.GetUserEntriesByCategory(user.ID, "cuisine", "Italian", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, entries, 2)
		for _, entry := range entries {
			assert.Equal(t, "Italian", entry.Cuisine)
		}
	})

	t.Run("should filter by location", func(t *testing.T) {
		entries, total, err := service.GetUserEntriesByCategory(user.ID, "location", "Downtown", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total)
		assert.Len(t, entries, 3)
		for _, entry := range entries {
			assert.Equal(t, "Downtown", entry.Location)
		}
	})

	t.Run("should filter by rating", func(t *testing.T) {
		entries, total, err := service.GetUserEntriesByCategory(user.ID, "rating", "5", 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, entries, 2)
		for _, entry := range entries {
			assert.Equal(t, 5, entry.FoodRating)
		}
	})

	t.Run("should handle invalid category", func(t *testing.T) {
		_, _, err := service.GetUserEntriesByCategory(user.ID, "invalid", "value", 10, 0)
		assert.Error(t, err)
	})

	t.Run("should handle nonexistent user", func(t *testing.T) {
		_, _, err := service.GetUserEntriesByCategory(999, "cuisine", "Italian", 10, 0)
		assert.Error(t, err)
	})
}

func TestProfileDisplayService_GetUserEntryCategories(t *testing.T) {
	db := setupProfileDisplayUnitTestDB(t)
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	service := NewProfileDisplayService(db, socialService, avatarService)

	// Create test user
	user := models.User{
		Username: "testuser",
		Password: "testpassword123",
		Email:    "test@example.com",
	}
	db.Create(&user)

	// Create test entries with different categories
	entries := []models.Entry{
		{Name: "Italian Place 1", Submitter: user.Username, Cuisine: "Italian", Location: "Downtown", MealService: "Dinner", FoodRating: 5},
		{Name: "Italian Place 2", Submitter: user.Username, Cuisine: "Italian", Location: "Uptown", MealService: "Lunch", FoodRating: 4},
		{Name: "Sushi Bar", Submitter: user.Username, Cuisine: "Japanese", Location: "Downtown", MealService: "Dinner", FoodRating: 5},
		{Name: "Taco Shop", Submitter: user.Username, Cuisine: "Mexican", Location: "Downtown", MealService: "Lunch", FoodRating: 3},
	}

	for _, entry := range entries {
		db.Create(&entry)
	}

	t.Run("should return all categories with counts", func(t *testing.T) {
		categories, err := service.GetUserEntryCategories(user.ID)
		require.NoError(t, err)

		// Check cuisine categories
		assert.Equal(t, 2, categories["cuisine"]["Italian"])
		assert.Equal(t, 1, categories["cuisine"]["Japanese"])
		assert.Equal(t, 1, categories["cuisine"]["Mexican"])

		// Check location categories
		assert.Equal(t, 3, categories["location"]["Downtown"])
		assert.Equal(t, 1, categories["location"]["Uptown"])

		// Check meal service categories
		assert.Equal(t, 2, categories["meal_service"]["Dinner"])
		assert.Equal(t, 2, categories["meal_service"]["Lunch"])

		// Check rating categories
		assert.Equal(t, 2, categories["rating"]["5"])
		assert.Equal(t, 1, categories["rating"]["4"])
		assert.Equal(t, 1, categories["rating"]["3"])
	})

	t.Run("should handle user with no entries", func(t *testing.T) {
		emptyUser := models.User{
			Username: "emptyuser",
			Password: "testpassword123",
			Email:    "empty@example.com",
		}
		db.Create(&emptyUser)

		categories, err := service.GetUserEntryCategories(emptyUser.ID)
		require.NoError(t, err)

		assert.Empty(t, categories["cuisine"])
		assert.Empty(t, categories["location"])
		assert.Empty(t, categories["meal_service"])
		assert.Empty(t, categories["rating"])
	})

	t.Run("should handle nonexistent user", func(t *testing.T) {
		_, err := service.GetUserEntryCategories(999)
		assert.Error(t, err)
	})
}

func TestProfileDisplayService_PrivacyFunctionality(t *testing.T) {
	db := setupProfileDisplayUnitTestDB(t)
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	service := NewProfileDisplayService(db, socialService, avatarService)

	// Create test user
	user := models.User{
		Username:     "testuser",
		Password:     "testpassword123",
		Email:        "test@example.com",
		PrivacyLevel: 0,
	}
	db.Create(&user)

	t.Run("should get privacy settings", func(t *testing.T) {
		settings, err := service.GetPrivacySettings(user.ID)
		require.NoError(t, err)

		assert.Equal(t, 0, settings["privacy_level"])
		assert.Equal(t, "Public - Anyone can see your profile and entries", settings["privacy_description"])
		assert.True(t, settings["can_be_found"].(bool))
		assert.True(t, settings["show_in_directory"].(bool))
	})

	t.Run("should check profile access correctly", func(t *testing.T) {
		// Create viewer
		viewer := models.User{
			Username: "viewer",
			Password: "testpassword123",
			Email:    "viewer@example.com",
		}
		db.Create(&viewer)

		// Test public profile access
		canAccess, reason, err := service.CheckProfileAccess(user.ID, viewer.ID)
		require.NoError(t, err)
		assert.True(t, canAccess)
		assert.Equal(t, "public", reason)

		// Test own profile access
		canAccess, reason, err = service.CheckProfileAccess(user.ID, user.ID)
		require.NoError(t, err)
		assert.True(t, canAccess)
		assert.Equal(t, "own_profile", reason)

		// Update to followers only
		err = service.UpdatePrivacySettings(user.ID, 1)
		require.NoError(t, err)

		// Test followers only without follow relationship
		canAccess, reason, err = service.CheckProfileAccess(user.ID, viewer.ID)
		require.NoError(t, err)
		assert.False(t, canAccess)
		assert.Equal(t, "followers_only", reason)

		// Create follow relationship
		follow := models.UserFollow{
			FollowerID:  viewer.ID,
			FollowingID: user.ID,
		}
		db.Create(&follow)

		// Test followers only with follow relationship
		canAccess, reason, err = service.CheckProfileAccess(user.ID, viewer.ID)
		require.NoError(t, err)
		assert.True(t, canAccess)
		assert.Equal(t, "follower", reason)

		// Update to private
		err = service.UpdatePrivacySettings(user.ID, 2)
		require.NoError(t, err)

		// Test private profile
		canAccess, reason, err = service.CheckProfileAccess(user.ID, viewer.ID)
		require.NoError(t, err)
		assert.False(t, canAccess)
		assert.Equal(t, "private", reason)
	})

	t.Run("should filter visible users correctly", func(t *testing.T) {
		// Create users with different privacy levels
		publicUser := models.User{
			Username:     "publicuser",
			Password:     "testpassword123",
			Email:        "public@example.com",
			PrivacyLevel: 0,
		}
		db.Create(&publicUser)

		followersOnlyUser := models.User{
			Username:     "followersonly",
			Password:     "testpassword123",
			Email:        "followers@example.com",
			PrivacyLevel: 1,
		}
		db.Create(&followersOnlyUser)

		privateUser := models.User{
			Username:     "privateuser",
			Password:     "testpassword123",
			Email:        "private@example.com",
			PrivacyLevel: 2,
		}
		db.Create(&privateUser)

		viewer := models.User{
			Username: "viewer",
			Password: "testpassword123",
			Email:    "viewer@example.com",
		}
		db.Create(&viewer)

		// Test without login (should only see public users)
		users, total, err := service.GetVisibleUsers(0, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(2), total) // publicUser + original user (now private)

		// Verify only public users are returned
		for _, u := range users {
			assert.Equal(t, 0, u.PrivacyLevel)
		}

		// Test with login but no follow relationships
		users, total, err = service.GetVisibleUsers(viewer.ID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(3), total) // publicUser + original user + viewer themselves

		// Create follow relationship
		follow := models.UserFollow{
			FollowerID:  viewer.ID,
			FollowingID: followersOnlyUser.ID,
		}
		db.Create(&follow)

		// Test with follow relationship
		users, total, err = service.GetVisibleUsers(viewer.ID, 10, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(4), total) // publicUser + original user + viewer + followersOnlyUser
	})

	t.Run("should handle privacy settings errors", func(t *testing.T) {
		// Test invalid privacy level
		err := service.UpdatePrivacySettings(user.ID, 5)
		assert.Error(t, err)

		// Test nonexistent user
		err = service.UpdatePrivacySettings(999, 1)
		assert.Error(t, err)

		// Test get settings for nonexistent user
		_, err = service.GetPrivacySettings(999)
		assert.Error(t, err)

		// Test check access for nonexistent user
		_, _, err = service.CheckProfileAccess(999, user.ID)
		assert.Error(t, err)
	})
}
