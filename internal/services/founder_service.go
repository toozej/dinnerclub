// Package services provides business logic for the dinnerclub application
package services

import (
	"fmt"
	"time"

	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/gorm"
)

// FounderService handles founder status management and statistics
type FounderService struct {
	db *gorm.DB
}

// NewFounderService creates a new founder service instance
func NewFounderService(db *gorm.DB) *FounderService {
	return &FounderService{db: db}
}

// GrantFounderStatus assigns founder status to a user
func (f *FounderService) GrantFounderStatus(userID uint) error {
	var user models.User
	if err := f.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.IsFounder {
		return nil // Already a founder
	}

	// Update user to founder status
	if err := f.db.Model(&user).Update("is_founder", true).Error; err != nil {
		return fmt.Errorf("failed to grant founder status: %w", err)
	}

	// Send congratulatory notification
	if err := f.sendFounderNotification(user); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("failed to send founder notification: %v\n", err)
	}

	return nil
}

// RevokeFounderStatus removes founder status from a user
func (f *FounderService) RevokeFounderStatus(userID uint) error {
	var user models.User
	if err := f.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !user.IsFounder {
		return nil // Not a founder
	}

	// Remove founder status
	if err := f.db.Model(&user).Update("is_founder", false).Error; err != nil {
		return fmt.Errorf("failed to revoke founder status: %w", err)
	}

	return nil
}

// IsFounder checks if a user has founder status
func (f *FounderService) IsFounder(userID uint) (bool, error) {
	var user models.User
	if err := f.db.Select("is_founder").First(&user, userID).Error; err != nil {
		return false, fmt.Errorf("user not found: %w", err)
	}

	return user.IsFounder, nil
}

// GetFounderCount returns the total number of founders
func (f *FounderService) GetFounderCount() (int64, error) {
	var count int64
	if err := f.db.Model(&models.User{}).Where("is_founder = ?", true).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count founders: %w", err)
	}

	return count, nil
}

// GetAllFounders returns all users with founder status
func (f *FounderService) GetAllFounders() ([]models.User, error) {
	var founders []models.User
	if err := f.db.Where("is_founder = ?", true).Order("join_date ASC").Find(&founders).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve founders: %w", err)
	}

	return founders, nil
}

// GetFounderInfo returns founder-specific information for a user
func (f *FounderService) GetFounderInfo(userID uint) (*FounderInfo, error) {
	var user models.User
	if err := f.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !user.IsFounder {
		return nil, fmt.Errorf("user is not a founder")
	}

	return &FounderInfo{
		UserID:      user.ID,
		Username:    user.Username,
		IsFounder:   user.IsFounder,
		JoinDate:    user.JoinDate,
		DisplayName: fmt.Sprintf("%s %s", user.Firstname, user.Lastname),
	}, nil
}

// ShouldDisplayFounderBadge determines if founder badge should be shown for a user
func (f *FounderService) ShouldDisplayFounderBadge(userID uint) bool {
	isFounder, err := f.IsFounder(userID)
	if err != nil {
		return false
	}
	return isFounder
}

// FounderInfo contains founder-specific information
type FounderInfo struct {
	UserID      uint      `json:"user_id"`
	Username    string    `json:"username"`
	IsFounder   bool      `json:"is_founder"`
	JoinDate    time.Time `json:"join_date"`
	DisplayName string    `json:"display_name"`
}

// sendFounderNotification sends a congratulatory notification when founder status is granted
func (f *FounderService) sendFounderNotification(user models.User) error {
	notification := models.SocialNotification{
		UserID:  user.ID,
		ActorID: user.ID, // System notification, actor is the user themselves
		Type:    "founder_status",
		Message: "Congratulations! You've been granted founder status. Welcome to the founding members of dinner.club!",
		Read:    false,
	}

	if err := f.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create founder notification: %w", err)
	}

	return nil
}

// GetFounderNotifications returns founder-related notifications for a user
func (f *FounderService) GetFounderNotifications(userID uint, limit int) ([]models.SocialNotification, error) {
	var notifications []models.SocialNotification

	query := f.db.Where("user_id = ? AND type = ?", userID, "founder_status").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve founder notifications: %w", err)
	}

	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func (f *FounderService) MarkNotificationAsRead(notificationID uint) error {
	if err := f.db.Model(&models.SocialNotification{}).
		Where("id = ?", notificationID).
		Update("read", true).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	return nil
}

// GetUnreadFounderNotificationCount returns the count of unread founder notifications
func (f *FounderService) GetUnreadFounderNotificationCount(userID uint) (int64, error) {
	var count int64
	if err := f.db.Model(&models.SocialNotification{}).
		Where("user_id = ? AND type = ? AND read = ?", userID, "founder_status", false).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}

	return count, nil
}
