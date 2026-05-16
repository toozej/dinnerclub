package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFounderTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

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

func TestGrantFounderStatus(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create a test user
	user := models.User{
		Username:  "testuser",
		Password:  "password123",
		Email:     "test@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&user).Error)

	// Grant founder status
	err := service.GrantFounderStatus(user.ID)
	require.NoError(t, err)

	// Verify founder status was granted
	var updatedUser models.User
	require.NoError(t, db.First(&updatedUser, user.ID).Error)
	assert.True(t, updatedUser.IsFounder)

	// Granting again should not error
	err = service.GrantFounderStatus(user.ID)
	require.NoError(t, err)
}

func TestGrantFounderStatus_UserNotFound(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	err := service.GrantFounderStatus(999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestRevokeFounderStatus(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create a founder user
	user := models.User{
		Username:  "founder",
		Password:  "password123",
		Email:     "founder@example.com",
		IsFounder: true,
	}
	require.NoError(t, db.Create(&user).Error)

	// Revoke founder status
	err := service.RevokeFounderStatus(user.ID)
	require.NoError(t, err)

	// Verify founder status was revoked
	var updatedUser models.User
	require.NoError(t, db.First(&updatedUser, user.ID).Error)
	assert.False(t, updatedUser.IsFounder)

	// Revoking again should not error
	err = service.RevokeFounderStatus(user.ID)
	require.NoError(t, err)
}

func TestIsFounder(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create founder and non-founder users
	founder := models.User{
		Username:  "founder",
		Password:  "password123",
		Email:     "founder@example.com",
		IsFounder: true,
	}
	require.NoError(t, db.Create(&founder).Error)

	regular := models.User{
		Username:  "regular",
		Password:  "password123",
		Email:     "regular@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&regular).Error)

	// Check founder status
	isFounder, err := service.IsFounder(founder.ID)
	require.NoError(t, err)
	assert.True(t, isFounder)

	isFounder, err = service.IsFounder(regular.ID)
	require.NoError(t, err)
	assert.False(t, isFounder)

	// Check non-existent user
	_, err = service.IsFounder(999)
	assert.Error(t, err)
}

func TestGetFounderCount(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Initially no founders
	count, err := service.GetFounderCount()
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Create some founders
	for i := 0; i < 3; i++ {
		user := models.User{
			Username:  "founder" + string(rune('1'+i)),
			Password:  "password123",
			Email:     "founder" + string(rune('1'+i)) + "@example.com",
			IsFounder: true,
		}
		require.NoError(t, db.Create(&user).Error)
	}

	// Create a non-founder
	regular := models.User{
		Username:  "regular",
		Password:  "password123",
		Email:     "regular@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&regular).Error)

	// Check founder count
	count, err = service.GetFounderCount()
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestGetAllFounders(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create founders with different join dates
	now := time.Now()
	founders := []models.User{
		{
			Username:  "founder1",
			Password:  "password123",
			Email:     "founder1@example.com",
			IsFounder: true,
			JoinDate:  now.Add(-48 * time.Hour),
		},
		{
			Username:  "founder2",
			Password:  "password123",
			Email:     "founder2@example.com",
			IsFounder: true,
			JoinDate:  now.Add(-24 * time.Hour),
		},
		{
			Username:  "founder3",
			Password:  "password123",
			Email:     "founder3@example.com",
			IsFounder: true,
			JoinDate:  now,
		},
	}

	for _, founder := range founders {
		require.NoError(t, db.Create(&founder).Error)
	}

	// Create a non-founder
	regular := models.User{
		Username:  "regular",
		Password:  "password123",
		Email:     "regular@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&regular).Error)

	// Get all founders
	result, err := service.GetAllFounders()
	require.NoError(t, err)
	assert.Len(t, result, 3)

	// Verify they're sorted by join date (oldest first)
	assert.Equal(t, "founder1", result[0].Username)
	assert.Equal(t, "founder2", result[1].Username)
	assert.Equal(t, "founder3", result[2].Username)
}

func TestGetFounderInfo(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create a founder
	founder := models.User{
		Username:  "founder",
		Password:  "password123",
		Email:     "founder@example.com",
		Firstname: "John",
		Lastname:  "Doe",
		IsFounder: true,
		JoinDate:  time.Now(),
	}
	require.NoError(t, db.Create(&founder).Error)

	// Get founder info
	info, err := service.GetFounderInfo(founder.ID)
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, founder.ID, info.UserID)
	assert.Equal(t, founder.Username, info.Username)
	assert.True(t, info.IsFounder)
	assert.Equal(t, "John Doe", info.DisplayName)

	// Try to get info for non-founder
	regular := models.User{
		Username:  "regular",
		Password:  "password123",
		Email:     "regular@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&regular).Error)

	_, err = service.GetFounderInfo(regular.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a founder")
}

func TestShouldDisplayFounderBadge(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create founder and non-founder
	founder := models.User{
		Username:  "founder",
		Password:  "password123",
		Email:     "founder@example.com",
		IsFounder: true,
	}
	require.NoError(t, db.Create(&founder).Error)

	regular := models.User{
		Username:  "regular",
		Password:  "password123",
		Email:     "regular@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&regular).Error)

	// Check badge display
	assert.True(t, service.ShouldDisplayFounderBadge(founder.ID))
	assert.False(t, service.ShouldDisplayFounderBadge(regular.ID))
	assert.False(t, service.ShouldDisplayFounderBadge(999)) // Non-existent user
}

func TestGrantFounderStatus_SendsNotification(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create a test user
	user := models.User{
		Username:  "testuser",
		Password:  "password123",
		Email:     "test@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&user).Error)

	// Grant founder status
	err := service.GrantFounderStatus(user.ID)
	require.NoError(t, err)

	// Verify notification was created
	var notifications []models.SocialNotification
	err = db.Where("user_id = ? AND type = ?", user.ID, "founder_status").Find(&notifications).Error
	require.NoError(t, err)
	assert.Len(t, notifications, 1)
	assert.Equal(t, user.ID, notifications[0].UserID)
	assert.Equal(t, "founder_status", notifications[0].Type)
	assert.Contains(t, notifications[0].Message, "Congratulations")
	assert.False(t, notifications[0].Read)
}

func TestGetFounderNotifications(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create a founder user
	user := models.User{
		Username:  "founder",
		Password:  "password123",
		Email:     "founder@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&user).Error)

	// Grant founder status (creates notification)
	err := service.GrantFounderStatus(user.ID)
	require.NoError(t, err)

	// Get notifications
	notifications, err := service.GetFounderNotifications(user.ID, 10)
	require.NoError(t, err)
	assert.Len(t, notifications, 1)
	assert.Equal(t, "founder_status", notifications[0].Type)
}

func TestGetFounderNotifications_WithLimit(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create a user
	user := models.User{
		Username:  "testuser",
		Password:  "password123",
		Email:     "test@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&user).Error)

	// Create multiple notifications manually
	for i := 0; i < 5; i++ {
		notification := models.SocialNotification{
			UserID:  user.ID,
			ActorID: user.ID,
			Type:    "founder_status",
			Message: fmt.Sprintf("Notification %d", i),
		}
		require.NoError(t, db.Create(&notification).Error)
	}

	// Get notifications with limit
	notifications, err := service.GetFounderNotifications(user.ID, 3)
	require.NoError(t, err)
	assert.Len(t, notifications, 3)
}

func TestMarkNotificationAsRead(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create a user and notification
	user := models.User{
		Username:  "testuser",
		Password:  "password123",
		Email:     "test@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&user).Error)

	notification := models.SocialNotification{
		UserID:  user.ID,
		ActorID: user.ID,
		Type:    "founder_status",
		Message: "Test notification",
		Read:    false,
	}
	require.NoError(t, db.Create(&notification).Error)

	// Mark as read
	err := service.MarkNotificationAsRead(notification.ID)
	require.NoError(t, err)

	// Verify it's marked as read
	var updatedNotif models.SocialNotification
	require.NoError(t, db.First(&updatedNotif, notification.ID).Error)
	assert.True(t, updatedNotif.Read)
}

func TestGetUnreadFounderNotificationCount(t *testing.T) {
	db := setupFounderTestDB(t)
	service := NewFounderService(db)

	// Create a user
	user := models.User{
		Username:  "testuser",
		Password:  "password123",
		Email:     "test@example.com",
		IsFounder: false,
	}
	require.NoError(t, db.Create(&user).Error)

	// Initially no unread notifications
	count, err := service.GetUnreadFounderNotificationCount(user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Create unread notifications
	for i := 0; i < 3; i++ {
		notification := models.SocialNotification{
			UserID:  user.ID,
			ActorID: user.ID,
			Type:    "founder_status",
			Message: fmt.Sprintf("Notification %d", i),
			Read:    false,
		}
		require.NoError(t, db.Create(&notification).Error)
	}

	// Create a read notification
	readNotif := models.SocialNotification{
		UserID:  user.ID,
		ActorID: user.ID,
		Type:    "founder_status",
		Message: "Read notification",
		Read:    true,
	}
	require.NoError(t, db.Create(&readNotif).Error)

	// Check unread count
	count, err = service.GetUnreadFounderNotificationCount(user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}
