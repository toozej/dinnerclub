// Package services provides business logic services for the dinnerclub application
package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"text/template"
	"unicode"

	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// ReferralService handles referral code generation, tracking, and statistics
type ReferralService struct {
	db       *gorm.DB
	cityCode string
	baseURL  string
}

// NewReferralService creates a new referral service instance
func NewReferralService(db *gorm.DB, cityCode, baseURL string) *ReferralService {
	return &ReferralService{
		db:       db,
		cityCode: cityCode,
		baseURL:  baseURL,
	}
}

// capitalizeFirst capitalizes only the first letter of a string.
// Unlike strings.Title (deprecated), this handles simple ASCII strings correctly.
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// GenerateReferralCode generates a secure random referral code for a user
func (r *ReferralService) GenerateReferralCode(userID uint) (string, error) {
	// Generate 8 bytes of random data
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string and make uppercase
	code := strings.ToUpper(hex.EncodeToString(bytes))

	// Update user's referral code in database
	if err := r.db.Model(&models.User{}).Where("id = ?", userID).Update("referral_code", code).Error; err != nil {
		return "", fmt.Errorf("failed to update user referral code: %w", err)
	}

	return code, nil
}

// ProcessReferral processes a referral when a new user signs up with a referral code
func (r *ReferralService) ProcessReferral(referralCode string, newUserID uint) error {
	if referralCode == "" {
		return nil // No referral code provided
	}

	// Find the referrer by referral code
	var referrer models.User
	if err := r.db.Where("referral_code = ?", referralCode).First(&referrer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("invalid referral code: %s", referralCode)
		}
		return fmt.Errorf("failed to find referrer: %w", err)
	}

	// Get the new user (referee) information
	var referee models.User
	if err := r.db.First(&referee, newUserID).Error; err != nil {
		return fmt.Errorf("failed to find referee: %w", err)
	}

	// Update the new user's ReferredBy field
	if err := r.db.Model(&models.User{}).Where("id = ?", newUserID).Update("referred_by", referrer.ID).Error; err != nil {
		return fmt.Errorf("failed to update referral relationship: %w", err)
	}

	// Send notifications to both referrer and referee
	if err := r.sendReferralNotifications(referrer, referee); err != nil {
		// Log error but don't fail the referral process
		// In a real application, you might want to use a proper logger
		fmt.Printf("Warning: Failed to send referral notifications: %v\n", err)
	}

	return nil
}

// sendReferralNotifications sends notifications to both referrer and referee
func (r *ReferralService) sendReferralNotifications(referrer, referee models.User) error {
	// Create notification for referrer
	referrerNotification := models.SocialNotification{
		UserID:    referrer.ID,
		ActorID:   referee.ID,
		Type:      "referral",
		Message:   fmt.Sprintf("%s %s joined using your referral code!", referee.Firstname, referee.Lastname),
		Read:      false,
		RelatedID: &referee.ID,
	}

	if err := r.db.Create(&referrerNotification).Error; err != nil {
		return fmt.Errorf("failed to create referrer notification: %w", err)
	}

	// Create notification for referee
	refereeNotification := models.SocialNotification{
		UserID:    referee.ID,
		ActorID:   referrer.ID,
		Type:      "referral",
		Message:   fmt.Sprintf("Welcome! You were referred by %s %s", referrer.Firstname, referrer.Lastname),
		Read:      false,
		RelatedID: &referrer.ID,
	}

	if err := r.db.Create(&refereeNotification).Error; err != nil {
		return fmt.Errorf("failed to create referee notification: %w", err)
	}

	return nil
}

// GetReferralStats returns referral statistics for a user
func (r *ReferralService) GetReferralStats(userID uint) (*models.ReferralStats, error) {
	var totalReferrals int64
	var activeReferrals int64

	// Count total referrals
	if err := r.db.Model(&models.User{}).Where("referred_by = ?", userID).Count(&totalReferrals).Error; err != nil {
		return nil, fmt.Errorf("failed to count total referrals: %w", err)
	}

	// Count active referrals (users who have logged in recently or have entries)
	// For now, we'll consider all referrals as active
	activeReferrals = totalReferrals

	return &models.ReferralStats{
		UserID:          userID,
		TotalReferrals:  int(totalReferrals),
		ActiveReferrals: int(activeReferrals),
	}, nil
}

// GetReferralHistory returns a list of users referred by the given user
func (r *ReferralService) GetReferralHistory(userID uint) ([]models.User, error) {
	var referrals []models.User
	if err := r.db.Where("referred_by = ?", userID).Find(&referrals).Error; err != nil {
		return nil, fmt.Errorf("failed to get referral history: %w", err)
	}
	return referrals, nil
}

// GetReferralNotifications returns referral-related notifications for a user
func (r *ReferralService) GetReferralNotifications(userID uint, limit int) ([]models.SocialNotification, error) {
	var notifications []models.SocialNotification
	query := r.db.Where("user_id = ? AND type = ?", userID, "referral").
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get referral notifications: %w", err)
	}
	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func (r *ReferralService) MarkNotificationAsRead(notificationID uint) error {
	if err := r.db.Model(&models.SocialNotification{}).
		Where("id = ?", notificationID).
		Update("read", true).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

// ReferralMessage represents the data for generating referral messages
type ReferralMessage struct {
	ReferrerName string
	ReferralCode string
	CustomText   string
	JoinURL      string
	CityCode     string
	CityName     string
}

// GenerateReferralMessage generates a formatted referral message for a user
func (r *ReferralService) GenerateReferralMessage(userID uint, customMessage string) (string, error) {
	// Get user information
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return "", fmt.Errorf("failed to find user: %w", err)
	}

	// Ensure user has a referral code
	if user.ReferralCode == "" {
		code, err := r.GenerateReferralCode(userID)
		if err != nil {
			return "", fmt.Errorf("failed to generate referral code: %w", err)
		}
		user.ReferralCode = code
	}

	// Create referral message data
	messageData := ReferralMessage{
		ReferrerName: user.Firstname + " " + user.Lastname,
		ReferralCode: user.ReferralCode,
		CustomText:   customMessage,
		JoinURL:      fmt.Sprintf("%s/register", r.baseURL),
		CityCode:     r.cityCode,
		CityName:     capitalizeFirst(r.cityCode), // Simple capitalization
	}

	// If no custom message provided, use default
	if messageData.CustomText == "" {
		messageData.CustomText = "I've been using this great platform to discover amazing restaurants and thought you'd love it too!"
	}

	// Template for referral message
	const referralTemplate = `Hi! {{.ReferrerName}} has invited you to join {{.CityCode}}dinner.club - a community for discovering and sharing great restaurants in {{.CityName}}.

{{.CustomText}}

Join using this link: {{.JoinURL}}?ref={{.ReferralCode}}

Happy dining!`

	// Parse and execute template
	tmpl, err := template.New("referral").Parse(referralTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse referral template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, messageData); err != nil {
		return "", fmt.Errorf("failed to execute referral template: %w", err)
	}

	return result.String(), nil
}
