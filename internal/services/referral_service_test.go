package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathrand "math/rand/v2"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// setupReferralTestDB creates an in-memory SQLite database for referral testing
func setupReferralTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.BackupCode{},
		&models.UserFollow{},
		&models.ProfileChange{},
		&models.EmailForwardingRule{},
		&models.SocialNotification{},
		&models.PasswordReset{},
	)
	require.NoError(t, err)

	return db
}

// generateRandomUser creates a random user for testing
func generateRandomUser() models.User {
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	suffix := hex.EncodeToString(randomBytes)

	return models.User{
		Username:  fmt.Sprintf("user_%s", suffix),
		Password:  "password123456",
		Firstname: fmt.Sprintf("First_%s", suffix),
		Lastname:  fmt.Sprintf("Last_%s", suffix),
		Email:     fmt.Sprintf("user_%s@example.com", suffix),
	}
}

// Property 1: Referral Message Generation
// **Feature: advanced-user-system, Property 1: Referral Message Generation**
// **Validates: Requirements 1.2, 1.3**
func TestProperty_ReferralMessageGeneration(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupReferralTestDB(t)
			service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

			// Generate random user
			user := generateRandomUser()
			result := db.Create(&user)
			require.NoError(t, result.Error)

			// Generate random custom message (sometimes empty)
			var customMessage string
			if mathrand.IntN(2) == 0 {
				customMessage = fmt.Sprintf("Custom message %d", mathrand.IntN(1000))
			}

			// Test the property: For any user, generating a referral message should always include their referral code and proper invitation formatting
			message, err := service.GenerateReferralMessage(user.ID, customMessage)

			// Assertions
			require.NoError(t, err)
			assert.NotEmpty(t, message)

			// Verify the message contains required components
			assert.Contains(t, message, user.Firstname+" "+user.Lastname, "Message should contain referrer name")
			assert.Contains(t, message, "dinner.club", "Message should contain dinner.club")
			assert.Contains(t, message, "https://nycdinnerclub.com/register", "Message should contain join URL")

			// Verify referral code is present and valid format (16 hex characters)
			referralCodeRegex := regexp.MustCompile(`ref=([A-F0-9]{16})`)
			matches := referralCodeRegex.FindStringSubmatch(message)
			assert.Len(t, matches, 2, "Message should contain a valid referral code")

			if len(matches) == 2 {
				referralCode := matches[1]
				assert.Len(t, referralCode, 16, "Referral code should be 16 characters")
				assert.Regexp(t, "^[A-F0-9]+$", referralCode, "Referral code should be uppercase hex")
			}

			// Verify custom message is included when provided
			if customMessage != "" {
				assert.Contains(t, message, customMessage, "Message should contain custom text when provided")
			} else {
				// Should contain default message
				assert.Contains(t, message, "I've been using this great platform", "Message should contain default text when no custom message")
			}

			// Verify message structure
			assert.Contains(t, message, "Hi!", "Message should start with greeting")
			assert.Contains(t, message, "Happy dining!", "Message should end with sign-off")
		})
	}
}

// Property 2: Referral Tracking Consistency
// **Feature: advanced-user-system, Property 2: Referral Tracking Consistency**
// **Validates: Requirements 1.4**
func TestProperty_ReferralTrackingConsistency(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupReferralTestDB(t)
			service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

			// Generate random referrer
			referrer := generateRandomUser()
			result := db.Create(&referrer)
			require.NoError(t, result.Error)

			// Generate referral code for referrer
			referralCode, err := service.GenerateReferralCode(referrer.ID)
			require.NoError(t, err)

			// Generate random new user (referee)
			referee := generateRandomUser()
			result = db.Create(&referee)
			require.NoError(t, result.Error)

			// Test the property: For any valid referral code used during signup,
			// the system should correctly establish and track the referral relationship
			err = service.ProcessReferral(referralCode, referee.ID)
			require.NoError(t, err)

			// Verify referral relationship is correctly established
			var updatedReferee models.User
			err = db.First(&updatedReferee, referee.ID).Error
			require.NoError(t, err)

			// Assertions for referral tracking consistency
			assert.NotNil(t, updatedReferee.ReferredBy, "Referee should have a referrer")
			assert.Equal(t, referrer.ID, *updatedReferee.ReferredBy, "Referral relationship should be correctly established")

			// Verify referral appears in referrer's statistics
			stats, err := service.GetReferralStats(referrer.ID)
			require.NoError(t, err)
			assert.Equal(t, 1, stats.TotalReferrals, "Referrer should have 1 total referral")
			assert.Equal(t, 1, stats.ActiveReferrals, "Referrer should have 1 active referral")

			// Verify referral appears in referrer's history
			history, err := service.GetReferralHistory(referrer.ID)
			require.NoError(t, err)
			assert.Len(t, history, 1, "Referrer should have 1 referral in history")
			assert.Equal(t, referee.ID, history[0].ID, "Referral history should contain the referee")

			// Test with multiple referrals from same referrer
			if mathrand.IntN(3) == 0 { // Randomly test multiple referrals
				secondReferee := generateRandomUser()
				db.Create(&secondReferee)

				err = service.ProcessReferral(referralCode, secondReferee.ID)
				require.NoError(t, err)

				// Verify updated statistics
				stats, err = service.GetReferralStats(referrer.ID)
				require.NoError(t, err)
				assert.Equal(t, 2, stats.TotalReferrals, "Referrer should have 2 total referrals")

				// Verify updated history
				history, err = service.GetReferralHistory(referrer.ID)
				require.NoError(t, err)
				assert.Len(t, history, 2, "Referrer should have 2 referrals in history")
			}
		})
	}
}

// Property 3: Referral Statistics Accuracy
// **Feature: advanced-user-system, Property 3: Referral Statistics Accuracy**
// **Validates: Requirements 1.5**
func TestProperty_ReferralStatisticsAccuracy(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupReferralTestDB(t)
			service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

			// Generate random referrer
			referrer := generateRandomUser()
			result := db.Create(&referrer)
			require.NoError(t, result.Error)

			// Generate referral code for referrer
			referralCode, err := service.GenerateReferralCode(referrer.ID)
			require.NoError(t, err)

			// Generate random number of referrals (0-10)
			numReferrals := mathrand.IntN(11)

			for j := 0; j < numReferrals; j++ {
				referee := generateRandomUser()
				db.Create(&referee)

				err = service.ProcessReferral(referralCode, referee.ID)
				require.NoError(t, err)
			}

			// Test the property: For any user, their displayed referral count should always match
			// the actual number of users they have successfully referred
			stats, err := service.GetReferralStats(referrer.ID)
			require.NoError(t, err)

			// Assertions for referral statistics accuracy
			assert.Equal(t, referrer.ID, stats.UserID, "Stats should be for the correct user")
			assert.Equal(t, numReferrals, stats.TotalReferrals, "Total referrals should match actual count")
			assert.Equal(t, numReferrals, stats.ActiveReferrals, "Active referrals should match total (for now)")

			// Verify referral history matches statistics
			history, err := service.GetReferralHistory(referrer.ID)
			require.NoError(t, err)
			assert.Len(t, history, numReferrals, "Referral history length should match statistics")

			// Verify all referees in history are actually referred by this user
			for _, historyUser := range history {
				assert.Equal(t, referrer.ID, *historyUser.ReferredBy, "All users in history should be referred by the referrer")
			}

			// Cross-verify with direct database query
			var dbCount int64
			db.Model(&models.User{}).Where("referred_by = ?", referrer.ID).Count(&dbCount)
			assert.Equal(t, int64(numReferrals), dbCount, "Database count should match statistics")
		})
	}
}

// Property 4: Referral Notification Delivery
// **Feature: advanced-user-system, Property 4: Referral Notification Delivery**
// **Validates: Requirements 1.6**
func TestProperty_ReferralNotificationDelivery(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupReferralTestDB(t)
			service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

			// Generate random referrer
			referrer := generateRandomUser()
			result := db.Create(&referrer)
			require.NoError(t, result.Error)

			// Generate referral code for referrer
			referralCode, err := service.GenerateReferralCode(referrer.ID)
			require.NoError(t, err)

			// Generate random referee
			referee := generateRandomUser()
			result = db.Create(&referee)
			require.NoError(t, result.Error)

			// Test the property: For any successful referral, both the referrer and referee
			// should receive appropriate notifications
			err = service.ProcessReferral(referralCode, referee.ID)
			require.NoError(t, err)

			// Verify referrer notification
			referrerNotifications, err := service.GetReferralNotifications(referrer.ID, 10)
			require.NoError(t, err)
			assert.Len(t, referrerNotifications, 1, "Referrer should receive exactly one notification")

			referrerNotif := referrerNotifications[0]
			assert.Equal(t, "referral", referrerNotif.Type, "Referrer notification should be of type 'referral'")
			assert.Equal(t, referrer.ID, referrerNotif.UserID, "Notification should be for the referrer")
			assert.Equal(t, referee.ID, referrerNotif.ActorID, "Notification actor should be the referee")
			assert.Contains(t, referrerNotif.Message, referee.Firstname, "Notification should contain referee's name")
			assert.Contains(t, referrerNotif.Message, "joined using your referral code", "Notification should mention referral code usage")
			assert.False(t, referrerNotif.Read, "Notification should initially be unread")
			assert.Equal(t, referee.ID, *referrerNotif.RelatedID, "Related ID should point to referee")

			// Verify referee notification
			refereeNotifications, err := service.GetReferralNotifications(referee.ID, 10)
			require.NoError(t, err)
			assert.Len(t, refereeNotifications, 1, "Referee should receive exactly one notification")

			refereeNotif := refereeNotifications[0]
			assert.Equal(t, "referral", refereeNotif.Type, "Referee notification should be of type 'referral'")
			assert.Equal(t, referee.ID, refereeNotif.UserID, "Notification should be for the referee")
			assert.Equal(t, referrer.ID, refereeNotif.ActorID, "Notification actor should be the referrer")
			assert.Contains(t, refereeNotif.Message, referrer.Firstname, "Notification should contain referrer's name")
			assert.Contains(t, refereeNotif.Message, "Welcome", "Notification should contain welcome message")
			assert.False(t, refereeNotif.Read, "Notification should initially be unread")
			assert.Equal(t, referrer.ID, *refereeNotif.RelatedID, "Related ID should point to referrer")

			// Test notification read functionality
			if mathrand.IntN(2) == 0 { // Randomly test marking as read
				err = service.MarkNotificationAsRead(referrerNotif.ID)
				require.NoError(t, err)

				// Verify notification is marked as read
				var updatedNotif models.SocialNotification
				db.First(&updatedNotif, referrerNotif.ID)
				assert.True(t, updatedNotif.Read, "Notification should be marked as read")
			}
		})
	}
}

// Unit tests for specific functionality
func TestReferralService_GenerateReferralCode(t *testing.T) {
	db := setupReferralTestDB(t)
	service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

	user := generateRandomUser()
	result := db.Create(&user)
	require.NoError(t, result.Error)

	t.Run("generates valid referral code", func(t *testing.T) {
		code, err := service.GenerateReferralCode(user.ID)

		assert.NoError(t, err)
		assert.Len(t, code, 16)
		assert.Regexp(t, "^[A-F0-9]+$", code)

		// Verify code is saved to database
		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.Equal(t, code, updatedUser.ReferralCode)
	})

	t.Run("generates unique codes", func(t *testing.T) {
		codes := make(map[string]bool)
		for i := 0; i < 10; i++ {
			user := generateRandomUser()
			db.Create(&user)

			code, err := service.GenerateReferralCode(user.ID)
			assert.NoError(t, err)
			assert.False(t, codes[code], "Code should be unique")
			codes[code] = true
		}
	})
}

func TestReferralService_ProcessReferral(t *testing.T) {
	db := setupReferralTestDB(t)
	service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

	// Create referrer
	referrer := generateRandomUser()
	db.Create(&referrer)
	code, err := service.GenerateReferralCode(referrer.ID)
	require.NoError(t, err)

	// Create new user
	newUser := generateRandomUser()
	db.Create(&newUser)

	t.Run("processes valid referral", func(t *testing.T) {
		err := service.ProcessReferral(code, newUser.ID)
		assert.NoError(t, err)

		// Verify referral relationship
		var updatedUser models.User
		db.First(&updatedUser, newUser.ID)
		assert.Equal(t, referrer.ID, *updatedUser.ReferredBy)
	})

	t.Run("handles empty referral code", func(t *testing.T) {
		anotherUser := generateRandomUser()
		db.Create(&anotherUser)

		err := service.ProcessReferral("", anotherUser.ID)
		assert.NoError(t, err)

		// Verify no referral relationship
		var updatedUser models.User
		db.First(&updatedUser, anotherUser.ID)
		assert.Nil(t, updatedUser.ReferredBy)
	})

	t.Run("handles invalid referral code", func(t *testing.T) {
		anotherUser := generateRandomUser()
		db.Create(&anotherUser)

		err := service.ProcessReferral("INVALIDCODE123", anotherUser.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid referral code")
	})
}

func TestReferralService_GetReferralStats(t *testing.T) {
	db := setupReferralTestDB(t)
	service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

	// Create referrer
	referrer := generateRandomUser()
	db.Create(&referrer)
	code, err := service.GenerateReferralCode(referrer.ID)
	require.NoError(t, err)

	t.Run("returns correct stats with no referrals", func(t *testing.T) {
		stats, err := service.GetReferralStats(referrer.ID)
		assert.NoError(t, err)
		assert.Equal(t, referrer.ID, stats.UserID)
		assert.Equal(t, 0, stats.TotalReferrals)
		assert.Equal(t, 0, stats.ActiveReferrals)
	})

	t.Run("returns correct stats with referrals", func(t *testing.T) {
		// Create referred users
		for i := 0; i < 3; i++ {
			newUser := generateRandomUser()
			db.Create(&newUser)
			service.ProcessReferral(code, newUser.ID)
		}

		stats, err := service.GetReferralStats(referrer.ID)
		assert.NoError(t, err)
		assert.Equal(t, referrer.ID, stats.UserID)
		assert.Equal(t, 3, stats.TotalReferrals)
		assert.Equal(t, 3, stats.ActiveReferrals)
	})
}

func TestReferralService_NotificationSystem(t *testing.T) {
	db := setupReferralTestDB(t)
	service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

	// Create referrer
	referrer := generateRandomUser()
	db.Create(&referrer)
	code, err := service.GenerateReferralCode(referrer.ID)
	require.NoError(t, err)

	// Create new user (referee)
	referee := generateRandomUser()
	db.Create(&referee)

	t.Run("creates notifications for both referrer and referee", func(t *testing.T) {
		err := service.ProcessReferral(code, referee.ID)
		assert.NoError(t, err)

		// Check referrer notifications
		referrerNotifications, err := service.GetReferralNotifications(referrer.ID, 10)
		assert.NoError(t, err)
		assert.Len(t, referrerNotifications, 1)
		assert.Equal(t, "referral", referrerNotifications[0].Type)
		assert.Contains(t, referrerNotifications[0].Message, referee.Firstname)
		assert.Contains(t, referrerNotifications[0].Message, "joined using your referral code")
		assert.False(t, referrerNotifications[0].Read)

		// Check referee notifications
		refereeNotifications, err := service.GetReferralNotifications(referee.ID, 10)
		assert.NoError(t, err)
		assert.Len(t, refereeNotifications, 1)
		assert.Equal(t, "referral", refereeNotifications[0].Type)
		assert.Contains(t, refereeNotifications[0].Message, referrer.Firstname)
		assert.Contains(t, refereeNotifications[0].Message, "Welcome")
		assert.False(t, refereeNotifications[0].Read)
	})

	t.Run("marks notifications as read", func(t *testing.T) {
		// Get a notification
		notifications, err := service.GetReferralNotifications(referrer.ID, 1)
		require.NoError(t, err)
		require.Len(t, notifications, 1)

		// Mark as read
		err = service.MarkNotificationAsRead(notifications[0].ID)
		assert.NoError(t, err)

		// Verify it's marked as read
		var updatedNotification models.SocialNotification
		db.First(&updatedNotification, notifications[0].ID)
		assert.True(t, updatedNotification.Read)
	})

	t.Run("handles multiple referrals with separate notifications", func(t *testing.T) {
		// Create another referee
		secondReferee := generateRandomUser()
		db.Create(&secondReferee)

		err := service.ProcessReferral(code, secondReferee.ID)
		assert.NoError(t, err)

		// Check referrer now has 2 notifications
		referrerNotifications, err := service.GetReferralNotifications(referrer.ID, 10)
		assert.NoError(t, err)
		assert.Len(t, referrerNotifications, 2)

		// Verify both notifications are different
		assert.NotEqual(t, referrerNotifications[0].Message, referrerNotifications[1].Message)
	})
}

func TestReferralService_GetReferralHistory(t *testing.T) {
	db := setupReferralTestDB(t)
	service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

	// Create referrer
	referrer := generateRandomUser()
	db.Create(&referrer)
	code, err := service.GenerateReferralCode(referrer.ID)
	require.NoError(t, err)

	t.Run("returns empty history with no referrals", func(t *testing.T) {
		history, err := service.GetReferralHistory(referrer.ID)
		assert.NoError(t, err)
		assert.Empty(t, history)
	})

	t.Run("returns correct referral history", func(t *testing.T) {
		// Create referred users
		for i := 0; i < 2; i++ {
			newUser := generateRandomUser()
			db.Create(&newUser)
			service.ProcessReferral(code, newUser.ID)
		}

		history, err := service.GetReferralHistory(referrer.ID)
		assert.NoError(t, err)
		assert.Len(t, history, 2)

		// Verify all returned users were referred by the referrer
		for _, user := range history {
			assert.Equal(t, referrer.ID, *user.ReferredBy)
		}
	})
}

func TestReferralService_GenerateReferralMessage_EdgeCases(t *testing.T) {
	db := setupReferralTestDB(t)
	service := NewReferralService(db, "nyc", "https://nycdinnerclub.com")

	t.Run("handles user without existing referral code", func(t *testing.T) {
		user := generateRandomUser()
		user.ReferralCode = "" // Ensure no existing code
		db.Create(&user)

		message, err := service.GenerateReferralMessage(user.ID, "")
		assert.NoError(t, err)
		assert.NotEmpty(t, message)

		// Verify code was generated and saved
		var updatedUser models.User
		db.First(&updatedUser, user.ID)
		assert.NotEmpty(t, updatedUser.ReferralCode)
		assert.Contains(t, message, updatedUser.ReferralCode)
	})

	t.Run("handles user with existing referral code", func(t *testing.T) {
		user := generateRandomUser()
		existingCode := "EXISTINGCODE1234"
		user.ReferralCode = existingCode
		db.Create(&user)

		message, err := service.GenerateReferralMessage(user.ID, "")
		assert.NoError(t, err)
		assert.Contains(t, message, existingCode)
	})

	t.Run("handles non-existent user", func(t *testing.T) {
		_, err := service.GenerateReferralMessage(99999, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find user")
	})

	t.Run("handles empty names gracefully", func(t *testing.T) {
		user := generateRandomUser()
		user.Firstname = ""
		user.Lastname = ""
		db.Create(&user)

		message, err := service.GenerateReferralMessage(user.ID, "")
		assert.NoError(t, err)
		assert.NotEmpty(t, message)
		// Should still contain the space between empty names
		assert.Contains(t, message, " ")
	})
}
