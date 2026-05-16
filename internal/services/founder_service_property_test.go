package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/toozej/dinnerclub/internal/models"
)

// **Validates: Requirements 6.1, 6.2, 6.3**
// Property 30: Founder Status Display
// For any user with founder status, a star icon should be displayed next to their username across all interfaces
func TestProperty30_FounderStatusDisplay(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 30: Founder status display is consistent", prop.ForAll(
		func(username string, isFounder bool) bool {
			if len(username) == 0 {
				return true // Skip invalid inputs
			}

			db := setupFounderTestDB(t)
			service := NewFounderService(db)

			// Create test user with specified founder status
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     username + "@example.com",
				IsFounder: isFounder,
			}
			if err := db.Create(&user).Error; err != nil {
				return true // Skip on error
			}

			// Verify ShouldDisplayFounderBadge returns correct value
			shouldDisplay := service.ShouldDisplayFounderBadge(user.ID)

			// The badge display should match the founder status
			return shouldDisplay == isFounder
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.Bool(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Validates: Requirements 6.4**
// Property 31: Founder Profile Information
// For any founder user's profile, it should display their founder badge and join date prominently
func TestProperty31_FounderProfileInformation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 31: Founder profile contains required information", prop.ForAll(
		func(username, firstname, lastname string) bool {
			if len(username) == 0 {
				return true // Skip invalid inputs
			}

			db := setupFounderTestDB(t)
			service := NewFounderService(db)

			// Create founder user
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     username + "@example.com",
				Firstname: firstname,
				Lastname:  lastname,
				IsFounder: true,
				JoinDate:  time.Now().Add(-30 * 24 * time.Hour),
			}
			if err := db.Create(&user).Error; err != nil {
				return true // Skip on error
			}

			// Get founder info
			info, err := service.GetFounderInfo(user.ID)
			if err != nil {
				return false
			}

			// Verify all required information is present
			return info.UserID == user.ID &&
				info.Username == user.Username &&
				info.IsFounder == true &&
				!info.JoinDate.IsZero() &&
				len(info.DisplayName) > 0
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Validates: Requirements 6.6**
// Property 33: Founder Statistics Accuracy
// For any founder statistics display, the total count should match the actual number of users with founder status
func TestProperty33_FounderStatisticsAccuracy(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 33: Founder count is accurate", prop.ForAll(
		func(numFounders, numRegular uint) bool {
			// Limit to reasonable numbers for test performance
			if numFounders > 20 || numRegular > 20 {
				return true
			}

			db := setupFounderTestDB(t)
			service := NewFounderService(db)

			// Create founder users
			for i := uint(0); i < numFounders; i++ {
				user := models.User{
					Username:  fmt.Sprintf("founder%d", i),
					Password:  "password123",
					Email:     fmt.Sprintf("founder%d@example.com", i),
					IsFounder: true,
				}
				if err := db.Create(&user).Error; err != nil {
					return true // Skip on error
				}
			}

			// Create regular users
			for i := uint(0); i < numRegular; i++ {
				user := models.User{
					Username:  fmt.Sprintf("regular%d", i),
					Password:  "password123",
					Email:     fmt.Sprintf("regular%d@example.com", i),
					IsFounder: false,
				}
				if err := db.Create(&user).Error; err != nil {
					return true // Skip on error
				}
			}

			// Get founder count
			count, err := service.GetFounderCount()
			if err != nil {
				return false
			}

			// Count should match the number of founders created
			return count == int64(numFounders)
		},
		gen.UIntRange(0, 20),
		gen.UIntRange(0, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional property test: Founder status transitions are idempotent
func TestProperty_FounderStatusIdempotence(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Granting founder status multiple times is idempotent", prop.ForAll(
		func(username string, iterations uint) bool {
			if len(username) == 0 || iterations == 0 || iterations > 10 {
				return true // Skip invalid inputs
			}

			db := setupFounderTestDB(t)
			service := NewFounderService(db)

			// Create regular user
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     username + "@example.com",
				IsFounder: false,
			}
			if err := db.Create(&user).Error; err != nil {
				return true // Skip on error
			}

			// Grant founder status multiple times
			for i := uint(0); i < iterations; i++ {
				if err := service.GrantFounderStatus(user.ID); err != nil {
					return false
				}
			}

			// Verify user is still a founder
			isFounder, err := service.IsFounder(user.ID)
			if err != nil {
				return false
			}

			return isFounder
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.UIntRange(1, 10),
	))

	properties.Property("Revoking founder status multiple times is idempotent", prop.ForAll(
		func(username string, iterations uint) bool {
			if len(username) == 0 || iterations == 0 || iterations > 10 {
				return true // Skip invalid inputs
			}

			db := setupFounderTestDB(t)
			service := NewFounderService(db)

			// Create founder user
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     username + "@example.com",
				IsFounder: true,
			}
			if err := db.Create(&user).Error; err != nil {
				return true // Skip on error
			}

			// Revoke founder status multiple times
			for i := uint(0); i < iterations; i++ {
				if err := service.RevokeFounderStatus(user.ID); err != nil {
					return false
				}
			}

			// Verify user is not a founder
			isFounder, err := service.IsFounder(user.ID)
			if err != nil {
				return false
			}

			return !isFounder
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
		gen.UIntRange(1, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional property test: GetAllFounders returns only founders sorted by join date
func TestProperty_GetAllFoundersCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("GetAllFounders returns only founders sorted by join date", prop.ForAll(
		func(numFounders uint) bool {
			// Limit to reasonable numbers for test performance
			if numFounders == 0 || numFounders > 15 {
				return true
			}

			db := setupFounderTestDB(t)
			service := NewFounderService(db)

			baseTime := time.Now().Add(-365 * 24 * time.Hour)

			// Create founders with different join dates
			for i := uint(0); i < numFounders; i++ {
				user := models.User{
					Username:  fmt.Sprintf("founder%d", i),
					Password:  "password123",
					Email:     fmt.Sprintf("founder%d@example.com", i),
					IsFounder: true,
					JoinDate:  baseTime.Add(time.Duration(i) * 24 * time.Hour),
				}
				if err := db.Create(&user).Error; err != nil {
					return true // Skip on error
				}
			}

			// Create some non-founders
			for i := uint(0); i < 3; i++ {
				user := models.User{
					Username:  fmt.Sprintf("regular%d", i),
					Password:  "password123",
					Email:     fmt.Sprintf("regular%d@example.com", i),
					IsFounder: false,
				}
				if err := db.Create(&user).Error; err != nil {
					return true // Skip on error
				}
			}

			// Get all founders
			founders, err := service.GetAllFounders()
			if err != nil {
				return false
			}

			// Verify count matches
			if len(founders) != int(numFounders) {
				return false
			}

			// Verify all are founders
			for _, founder := range founders {
				if !founder.IsFounder {
					return false
				}
			}

			// Verify sorted by join date (oldest first)
			for i := 1; i < len(founders); i++ {
				if founders[i].JoinDate.Before(founders[i-1].JoinDate) {
					return false
				}
			}

			return true
		},
		gen.UIntRange(1, 15),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Validates: Requirements 6.5**
// Property 32: Founder Status Notification
// For any user granted founder status, they should receive a congratulatory notification
func TestProperty32_FounderStatusNotification(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 32: Founder status grants send notifications", prop.ForAll(
		func(username string) bool {
			if len(username) == 0 {
				return true // Skip invalid inputs
			}

			db := setupFounderTestDB(t)
			service := NewFounderService(db)

			// Create regular user
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     username + "@example.com",
				IsFounder: false,
			}
			if err := db.Create(&user).Error; err != nil {
				return true // Skip on error
			}

			// Grant founder status
			if err := service.GrantFounderStatus(user.ID); err != nil {
				return false
			}

			// Verify notification was created
			var notifications []models.SocialNotification
			if err := db.Where("user_id = ? AND type = ?", user.ID, "founder_status").
				Find(&notifications).Error; err != nil {
				return false
			}

			// Should have exactly one notification
			if len(notifications) != 1 {
				return false
			}

			// Verify notification properties
			notif := notifications[0]
			return notif.UserID == user.ID &&
				notif.Type == "founder_status" &&
				len(notif.Message) > 0 &&
				!notif.Read
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	properties.Property("Granting founder status twice doesn't create duplicate notifications", prop.ForAll(
		func(username string) bool {
			if len(username) == 0 {
				return true // Skip invalid inputs
			}

			db := setupFounderTestDB(t)
			service := NewFounderService(db)

			// Create regular user
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     username + "@example.com",
				IsFounder: false,
			}
			if err := db.Create(&user).Error; err != nil {
				return true // Skip on error
			}

			// Grant founder status twice
			if err := service.GrantFounderStatus(user.ID); err != nil {
				return false
			}
			if err := service.GrantFounderStatus(user.ID); err != nil {
				return false
			}

			// Verify only one notification was created
			var notifications []models.SocialNotification
			if err := db.Where("user_id = ? AND type = ?", user.ID, "founder_status").
				Find(&notifications).Error; err != nil {
				return false
			}

			// Should still have exactly one notification
			return len(notifications) == 1
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional property test: Notification retrieval and management
func TestProperty_FounderNotificationManagement(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Unread notification count is accurate", prop.ForAll(
		func(username string) bool {
			if len(username) == 0 {
				return true // Skip invalid inputs
			}

			db := setupFounderTestDB(t)
			service := NewFounderService(db)

			// Create user
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     username + "@example.com",
				IsFounder: false,
			}
			if err := db.Create(&user).Error; err != nil {
				return true // Skip on error
			}

			// Initially no unread notifications
			count, err := service.GetUnreadFounderNotificationCount(user.ID)
			if err != nil || count != 0 {
				return false
			}

			// Grant founder status (creates notification)
			if err := service.GrantFounderStatus(user.ID); err != nil {
				return false
			}

			// Should have 1 unread notification
			count, err = service.GetUnreadFounderNotificationCount(user.ID)
			if err != nil || count != 1 {
				return false
			}

			// Get notifications
			notifications, err := service.GetFounderNotifications(user.ID, 10)
			if err != nil || len(notifications) != 1 {
				return false
			}

			// Mark as read
			if err := service.MarkNotificationAsRead(notifications[0].ID); err != nil {
				return false
			}

			// Should have 0 unread notifications
			count, err = service.GetUnreadFounderNotificationCount(user.ID)
			if err != nil || count != 0 {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
