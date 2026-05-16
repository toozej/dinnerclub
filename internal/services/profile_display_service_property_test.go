// Package services provides property-based tests for profile display functionality
// validating correctness properties for user profile information display,
// statistics accuracy, and user name navigation.
package services

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupProfileDisplayTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto migrate the schema
	db.AutoMigrate(&models.User{}, &models.Entry{}, &models.UserFollow{}, &models.BackupCode{},
		&models.ProfileChange{}, &models.SocialNotification{}, &models.PasswordReset{})

	return db
}

// Property 50: Profile Information Display
// For any user profile, the displayed information should accurately reflect the user's stored data
func TestProperty50_ProfileInformationDisplay(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("profile information should accurately reflect stored user data", prop.ForAll(
		func(username, firstname, lastname, bio string, isFounder bool, privacyLevel int) bool {
			db := setupProfileDisplayTestDB()
			socialService := NewSocialService(db)
			avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
			profileService := NewProfileDisplayService(db, socialService, avatarService)

			// Ensure privacy level is valid
			if privacyLevel < 0 || privacyLevel > 2 {
				privacyLevel = 0
			}

			// Create test user
			user := models.User{
				Username:     username,
				Password:     "testpassword123",
				Firstname:    firstname,
				Lastname:     lastname,
				Email:        username + "@test.com",
				Bio:          bio,
				IsFounder:    isFounder,
				PrivacyLevel: privacyLevel,
			}
			db.Create(&user)

			// Get enhanced profile
			profile, err := profileService.GetEnhancedUserProfile(user.ID, user.ID)
			if err != nil {
				return false
			}

			// Verify all profile information matches stored data
			return profile.User.Username == username &&
				profile.User.Firstname == firstname &&
				profile.User.Lastname == lastname &&
				profile.User.Bio == bio &&
				profile.User.IsFounder == isFounder &&
				profile.User.PrivacyLevel == privacyLevel &&
				profile.CanEdit == true && // User viewing their own profile
				profile.IsVisible == true // User can always see their own profile
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 50 }),
		gen.AlphaString(),
		gen.AlphaString(),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) <= 500 }),
		gen.Bool(),
		gen.IntRange(0, 2),
	))

	properties.TestingRun(t)
}

// Property 51: Profile Statistics Accuracy
// For any user, their profile statistics should accurately reflect their actual data counts
func TestProperty51_ProfileStatisticsAccuracy(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("profile statistics should accurately reflect actual data counts", prop.ForAll(
		func(numEntries, numFollowers, numFollowing, numReferrals int) bool {
			db := setupProfileDisplayTestDB()
			socialService := NewSocialService(db)
			avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
			profileService := NewProfileDisplayService(db, socialService, avatarService)

			// Limit numbers to reasonable ranges for testing
			numEntries = numEntries % 20
			numFollowers = numFollowers % 10
			numFollowing = numFollowing % 10
			numReferrals = numReferrals % 5

			// Create test user
			user := models.User{
				Username: "testuser",
				Password: "testpassword123",
				Email:    "test@example.com",
			}
			db.Create(&user)

			// Create entries
			for i := 0; i < numEntries; i++ {
				entry := models.Entry{
					Name:      "Restaurant " + string(rune(i)),
					Submitter: user.Username,
					Location:  "Location " + string(rune(i)),
					Cuisine:   "Cuisine " + string(rune(i)),
				}
				db.Create(&entry)
			}

			// Create followers
			for i := 0; i < numFollowers; i++ {
				follower := models.User{
					Username: "follower" + string(rune(i)),
					Password: "testpassword123",
					Email:    "follower" + string(rune(i)) + "@example.com",
				}
				db.Create(&follower)

				follow := models.UserFollow{
					FollowerID:  follower.ID,
					FollowingID: user.ID,
				}
				db.Create(&follow)
			}

			// Create following relationships
			for i := 0; i < numFollowing; i++ {
				following := models.User{
					Username: "following" + string(rune(i)),
					Password: "testpassword123",
					Email:    "following" + string(rune(i)) + "@example.com",
				}
				db.Create(&following)

				follow := models.UserFollow{
					FollowerID:  user.ID,
					FollowingID: following.ID,
				}
				db.Create(&follow)
			}

			// Create referrals
			for i := 0; i < numReferrals; i++ {
				referral := models.User{
					Username:   "referral" + string(rune(i)),
					Password:   "testpassword123",
					Email:      "referral" + string(rune(i)) + "@example.com",
					ReferredBy: &user.ID,
				}
				db.Create(&referral)
			}

			// Get profile statistics
			profile, err := profileService.GetEnhancedUserProfile(user.ID, user.ID)
			if err != nil {
				return false
			}

			// Verify statistics accuracy
			return profile.EntryCount == numEntries &&
				profile.FollowerCount == numFollowers &&
				profile.FollowingCount == numFollowing &&
				profile.ReferralCount == numReferrals
		},
		gen.IntRange(0, 20),
		gen.IntRange(0, 10),
		gen.IntRange(0, 10),
		gen.IntRange(0, 5),
	))

	properties.TestingRun(t)
}

// Property 52: User Name Navigation
// For any valid username, navigation to their profile should work correctly and display appropriate information
func TestProperty52_UserNameNavigation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("user name navigation should work correctly for valid usernames", prop.ForAll(
		func(username string, isPublic bool) bool {
			db := setupProfileDisplayTestDB()
			socialService := NewSocialService(db)
			avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
			profileService := NewProfileDisplayService(db, socialService, avatarService)

			// Create test user with appropriate privacy level
			privacyLevel := 0
			if !isPublic {
				privacyLevel = 2 // Private
			}

			user := models.User{
				Username:     username,
				Password:     "testpassword123",
				Email:        username + "@test.com",
				PrivacyLevel: privacyLevel,
			}
			db.Create(&user)

			// Create a viewer user
			viewer := models.User{
				Username: "viewer",
				Password: "testpassword123",
				Email:    "viewer@test.com",
			}
			db.Create(&viewer)

			// Test navigation by username
			profile, err := profileService.GetUserProfileByUsername(username, viewer.ID)
			if err != nil {
				return false
			}

			// Verify profile belongs to correct user
			if profile.User.Username != username {
				return false
			}

			// Verify visibility is correctly determined
			expectedVisibility := isPublic
			if profile.IsVisible != expectedVisibility {
				return false
			}

			// Verify follow button is shown for public profiles when not viewing own profile
			expectedShowFollowButton := isPublic && (viewer.ID != user.ID)
			return profile.ShowFollowButton == expectedShowFollowButton
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 50 }),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Additional property test for privacy enforcement
func TestProperty_PrivacyEnforcement(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("privacy settings should be correctly enforced", prop.ForAll(
		func(privacyLevel int, isFollowing bool) bool {
			db := setupProfileDisplayTestDB()
			socialService := NewSocialService(db)
			avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
			profileService := NewProfileDisplayService(db, socialService, avatarService)

			// Ensure privacy level is valid
			if privacyLevel < 0 || privacyLevel > 2 {
				privacyLevel = privacyLevel % 3
				if privacyLevel < 0 {
					privacyLevel = 0
				}
			}

			// Create test user
			user := models.User{
				Username:     "testuser",
				Password:     "testpassword123",
				Email:        "test@example.com",
				PrivacyLevel: privacyLevel,
			}
			db.Create(&user)

			// Create viewer
			viewer := models.User{
				Username: "viewer",
				Password: "testpassword123",
				Email:    "viewer@example.com",
			}
			db.Create(&viewer)

			// Create following relationship if needed
			if isFollowing {
				follow := models.UserFollow{
					FollowerID:  viewer.ID,
					FollowingID: user.ID,
				}
				db.Create(&follow)
			}

			// Get profile
			profile, err := profileService.GetEnhancedUserProfile(user.ID, viewer.ID)
			if err != nil {
				return false
			}

			// Determine expected visibility based on privacy level and relationship
			var expectedVisibility bool
			var expectedShowContact bool

			switch privacyLevel {
			case 0: // Public
				expectedVisibility = true
				expectedShowContact = false
			case 1: // Followers only
				expectedVisibility = isFollowing
				expectedShowContact = isFollowing
			case 2: // Private
				expectedVisibility = false
				expectedShowContact = false
			}

			return profile.IsVisible == expectedVisibility &&
				profile.ShowContactInfo == expectedShowContact
		},
		gen.IntRange(0, 2),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Property 53: Profile Entry Display
// For any user with entries, their profile should correctly display recent entries with proper formatting
func TestProperty53_ProfileEntryDisplay(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("profile should correctly display recent entries", prop.ForAll(
		func(numEntries int) bool {
			db := setupProfileDisplayTestDB()
			socialService := NewSocialService(db)
			avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
			profileService := NewProfileDisplayService(db, socialService, avatarService)

			// Limit entries to reasonable range
			numEntries = (numEntries % 10) + 1 // 1-10 entries

			// Create test user
			user := models.User{
				Username: "testuser",
				Password: "testpassword123",
				Email:    "test@example.com",
			}
			db.Create(&user)

			// Create entries
			for i := 0; i < numEntries; i++ {
				entry := models.Entry{
					Name:       "Restaurant " + string(rune(i)),
					Submitter:  user.Username,
					Location:   "Location " + string(rune(i)),
					Comments:   "Comments " + string(rune(i)),
					Cuisine:    "Cuisine " + string(rune(i%3)), // Create some variety
					FoodRating: (i % 5) + 1,                    // Ratings 1-5
				}
				db.Create(&entry)
			}

			// Get profile
			profile, err := profileService.GetEnhancedUserProfile(user.ID, user.ID)
			if err != nil {
				return false
			}

			// Verify entry display
			expectedRecentCount := numEntries
			if numEntries > 5 {
				expectedRecentCount = 5 // Recent entries are limited to 5
			}

			if len(profile.RecentEntries) != expectedRecentCount {
				return false
			}

			// Verify entries are ordered by creation date (most recent first)
			for i := 1; i < len(profile.RecentEntries); i++ {
				if profile.RecentEntries[i-1].CreatedAt.Before(profile.RecentEntries[i].CreatedAt) {
					return false
				}
			}

			// Verify entry count is accurate
			return profile.EntryCount == numEntries
		},
		gen.IntRange(1, 15),
	))

	properties.TestingRun(t)
}

// Property 54: Profile Navigation Options
// For any user profile, navigation options should be appropriate based on viewer permissions and relationship
func TestProperty54_ProfileNavigationOptions(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("profile navigation options should be appropriate for viewer", prop.ForAll(
		func(isOwnProfile, hasEntries bool) bool {
			db := setupProfileDisplayTestDB()
			socialService := NewSocialService(db)
			avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
			profileService := NewProfileDisplayService(db, socialService, avatarService)

			// Create test user
			user := models.User{
				Username: "testuser",
				Password: "testpassword123",
				Email:    "test@example.com",
			}
			db.Create(&user)

			// Create viewer (same user if own profile)
			var viewerID uint
			if isOwnProfile {
				viewerID = user.ID
			} else {
				viewer := models.User{
					Username: "viewer",
					Password: "testpassword123",
					Email:    "viewer@example.com",
				}
				db.Create(&viewer)
				viewerID = viewer.ID
			}

			// Create entries if needed
			if hasEntries {
				for i := 0; i < 3; i++ {
					entry := models.Entry{
						Name:      "Restaurant " + string(rune(i)),
						Submitter: user.Username,
						Location:  "Location " + string(rune(i)),
						Cuisine:   "Cuisine " + string(rune(i)),
					}
					db.Create(&entry)
				}
			}

			// Get profile
			profile, err := profileService.GetEnhancedUserProfile(user.ID, viewerID)
			if err != nil {
				return false
			}

			// Verify navigation options
			expectedCanEdit := isOwnProfile
			expectedShowFollowButton := !isOwnProfile && profile.IsVisible

			if profile.CanEdit != expectedCanEdit {
				return false
			}

			if profile.ShowFollowButton != expectedShowFollowButton {
				return false
			}

			// Verify entry count matches actual entries
			expectedEntryCount := 0
			if hasEntries {
				expectedEntryCount = 3
			}

			return profile.EntryCount == expectedEntryCount
		},
		gen.Bool(),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Property test for entry categories and filtering
func TestProperty_EntryCategories(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("entry categories should accurately reflect user's entries", prop.ForAll(
		func(cuisines []string, locations []string) bool {
			db := setupProfileDisplayTestDB()
			socialService := NewSocialService(db)
			avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
			profileService := NewProfileDisplayService(db, socialService, avatarService)

			// Limit array sizes
			if len(cuisines) > 5 {
				cuisines = cuisines[:5]
			}
			if len(locations) > 5 {
				locations = locations[:5]
			}

			// Filter out empty strings
			var validCuisines, validLocations []string
			for _, c := range cuisines {
				if c != "" {
					validCuisines = append(validCuisines, c)
				}
			}
			for _, l := range locations {
				if l != "" {
					validLocations = append(validLocations, l)
				}
			}

			if len(validCuisines) == 0 || len(validLocations) == 0 {
				return true // Skip if no valid data
			}

			// Create test user
			user := models.User{
				Username: "testuser",
				Password: "testpassword123",
				Email:    "test@example.com",
			}
			db.Create(&user)

			// Create entries with different categories
			entryCount := 0
			for _, cuisine := range validCuisines {
				for _, location := range validLocations {
					entry := models.Entry{
						Name:       "Restaurant " + string(rune(entryCount)),
						Submitter:  user.Username,
						Cuisine:    cuisine,
						Location:   location,
						FoodRating: (entryCount % 5) + 1,
					}
					db.Create(&entry)
					entryCount++
				}
			}

			// Get entry categories
			categories, err := profileService.GetUserEntryCategories(user.ID)
			if err != nil {
				return false
			}

			// Verify cuisine categories
			if len(categories["cuisine"]) != len(validCuisines) {
				return false
			}

			for _, cuisine := range validCuisines {
				if count, exists := categories["cuisine"][cuisine]; !exists || count != len(validLocations) {
					return false
				}
			}

			// Verify location categories
			if len(categories["location"]) != len(validLocations) {
				return false
			}

			for _, location := range validLocations {
				if count, exists := categories["location"][location]; !exists || count != len(validCuisines) {
					return false
				}
			}

			return true
		},
		gen.SliceOf(gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 20 })),
		gen.SliceOf(gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 20 })),
	))

	properties.TestingRun(t)
}

// Property 55: Privacy Settings Enforcement
// For any privacy level and viewer relationship, access controls should be correctly enforced
func TestProperty55_PrivacySettingsEnforcement(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("privacy settings should be correctly enforced across all scenarios", prop.ForAll(
		func(privacyLevel int, hasFollowRelationship, isLoggedIn bool) bool {
			db := setupProfileDisplayTestDB()
			socialService := NewSocialService(db)
			avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
			profileService := NewProfileDisplayService(db, socialService, avatarService)

			// Ensure privacy level is valid
			if privacyLevel < 0 || privacyLevel > 2 {
				privacyLevel = privacyLevel % 3
				if privacyLevel < 0 {
					privacyLevel = 0
				}
			}

			// Create target user with specific privacy level
			targetUser := models.User{
				Username:     "targetuser",
				Password:     "testpassword123",
				Email:        "target@example.com",
				PrivacyLevel: privacyLevel,
			}
			db.Create(&targetUser)

			// Create viewer (if logged in)
			var viewerID uint = 0
			if isLoggedIn {
				viewer := models.User{
					Username: "viewer",
					Password: "testpassword123",
					Email:    "viewer@example.com",
				}
				db.Create(&viewer)
				viewerID = viewer.ID

				// Create follow relationship if specified
				if hasFollowRelationship {
					follow := models.UserFollow{
						FollowerID:  viewer.ID,
						FollowingID: targetUser.ID,
					}
					db.Create(&follow)
				}
			}

			// Test profile access
			canAccess, reason, err := profileService.CheckProfileAccess(targetUser.ID, viewerID)
			if err != nil {
				return false
			}

			// Determine expected access based on privacy level and relationship
			var expectedAccess bool
			var expectedReason string

			switch privacyLevel {
			case 0: // Public
				expectedAccess = true
				expectedReason = "public"
			case 1: // Followers only
				if !isLoggedIn {
					expectedAccess = false
					expectedReason = "login_required"
				} else if hasFollowRelationship {
					expectedAccess = true
					expectedReason = "follower"
				} else {
					expectedAccess = false
					expectedReason = "followers_only"
				}
			case 2: // Private
				expectedAccess = false
				expectedReason = "private"
			}

			// Verify access control
			if canAccess != expectedAccess || reason != expectedReason {
				return false
			}

			// Test profile visibility through GetEnhancedUserProfile
			profile, err := profileService.GetEnhancedUserProfile(targetUser.ID, viewerID)
			if err != nil {
				return false
			}

			// Verify visibility matches access control
			expectedVisibility := expectedAccess
			if profile.IsVisible != expectedVisibility {
				return false
			}

			// Verify follow button display
			expectedShowFollowButton := expectedAccess && isLoggedIn && (viewerID != targetUser.ID)
			if profile.ShowFollowButton != expectedShowFollowButton {
				return false
			}

			// Verify contact info display
			expectedShowContactInfo := false
			if privacyLevel == 1 && hasFollowRelationship && isLoggedIn {
				expectedShowContactInfo = true
			}
			if viewerID == targetUser.ID {
				expectedShowContactInfo = true // User can always see their own contact info
			}

			return profile.ShowContactInfo == expectedShowContactInfo
		},
		gen.IntRange(0, 2),
		gen.Bool(),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Additional property test for visible users functionality
func TestProperty_VisibleUsersFiltering(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("visible users should be correctly filtered based on privacy settings", prop.ForAll(
		func(numPublic, numFollowersOnly, numPrivate int, isLoggedIn bool) bool {
			db := setupProfileDisplayTestDB()
			socialService := NewSocialService(db)
			avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
			profileService := NewProfileDisplayService(db, socialService, avatarService)

			// Limit numbers to reasonable ranges
			numPublic = (numPublic % 5) + 1
			numFollowersOnly = (numFollowersOnly % 3) + 1
			numPrivate = (numPrivate % 3) + 1

			var viewerID uint = 0
			var followingIDs []uint

			// Create viewer if logged in
			if isLoggedIn {
				viewer := models.User{
					Username: "viewer",
					Password: "testpassword123",
					Email:    "viewer@example.com",
				}
				db.Create(&viewer)
				viewerID = viewer.ID
			}

			// Create public users
			for i := 0; i < numPublic; i++ {
				user := models.User{
					Username:     "public" + string(rune(i)),
					Password:     "testpassword123",
					Email:        "public" + string(rune(i)) + "@example.com",
					PrivacyLevel: 0,
				}
				db.Create(&user)
			}

			// Create followers-only users
			for i := 0; i < numFollowersOnly; i++ {
				user := models.User{
					Username:     "followers" + string(rune(i)),
					Password:     "testpassword123",
					Email:        "followers" + string(rune(i)) + "@example.com",
					PrivacyLevel: 1,
				}
				db.Create(&user)

				// Make viewer follow half of them (if logged in)
				if isLoggedIn && i%2 == 0 {
					follow := models.UserFollow{
						FollowerID:  viewerID,
						FollowingID: user.ID,
					}
					db.Create(&follow)
					followingIDs = append(followingIDs, user.ID)
				}
			}

			// Create private users
			for i := 0; i < numPrivate; i++ {
				user := models.User{
					Username:     "private" + string(rune(i)),
					Password:     "testpassword123",
					Email:        "private" + string(rune(i)) + "@example.com",
					PrivacyLevel: 2,
				}
				db.Create(&user)
			}

			// Get visible users
			users, total, err := profileService.GetVisibleUsers(viewerID, 100, 0)
			if err != nil {
				return false
			}

			// Calculate expected count
			expectedCount := numPublic // All public users are visible
			if isLoggedIn {
				expectedCount += 1                 // Viewer can see themselves
				expectedCount += len(followingIDs) // Followed users with followers-only privacy
			}
			// Private users are never visible to others

			// Verify total count
			if int(total) != expectedCount {
				return false
			}

			// Verify returned users count
			if len(users) != expectedCount {
				return false
			}

			// Verify no private users are returned (unless it's the viewer themselves)
			for _, user := range users {
				if user.PrivacyLevel == 2 && user.ID != viewerID {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 5),
		gen.IntRange(1, 3),
		gen.IntRange(1, 3),
		gen.Bool(),
	))

	properties.TestingRun(t)
}

// Unit tests for specific edge cases
func TestProfileDisplayService_EdgeCases(t *testing.T) {
	db := setupProfileDisplayTestDB()
	socialService := NewSocialService(db)
	avatarService := NewAvatarService(db, "test_uploads", 1024*1024, []string{"jpg", "png"})
	profileService := NewProfileDisplayService(db, socialService, avatarService)

	t.Run("nonexistent user should return error", func(t *testing.T) {
		_, err := profileService.GetUserProfileByUsername("nonexistent", 0)
		assert.Error(t, err)
	})

	t.Run("user viewing own profile should always be visible", func(t *testing.T) {
		user := models.User{
			Username:     "privateuser",
			Password:     "testpassword123",
			Email:        "private@example.com",
			PrivacyLevel: 2, // Private
		}
		db.Create(&user)

		profile, err := profileService.GetEnhancedUserProfile(user.ID, user.ID)
		require.NoError(t, err)
		assert.True(t, profile.IsVisible)
		assert.True(t, profile.CanEdit)
		assert.False(t, profile.ShowFollowButton)
	})

	t.Run("privacy level update should work correctly", func(t *testing.T) {
		user := models.User{
			Username:     "testuser",
			Password:     "testpassword123",
			Email:        "test@example.com",
			PrivacyLevel: 0,
		}
		db.Create(&user)

		// Update privacy level
		err := profileService.UpdatePrivacySettings(user.ID, 2)
		assert.NoError(t, err)

		// Verify update
		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.Equal(t, 2, updatedUser.PrivacyLevel)
	})

	t.Run("invalid privacy level should return error", func(t *testing.T) {
		user := models.User{
			Username: "testuser",
			Password: "testpassword123",
			Email:    "test@example.com",
		}
		db.Create(&user)

		err := profileService.UpdatePrivacySettings(user.ID, 5)
		assert.Error(t, err)
	})
}
