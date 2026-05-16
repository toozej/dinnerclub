package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// ProfileService handles profile change confirmations and management
type ProfileService struct {
	db           *gorm.DB
	emailService *EmailService
}

// NewProfileService creates a new ProfileService instance
func NewProfileService(db *gorm.DB, emailService *EmailService) *ProfileService {
	return &ProfileService{
		db:           db,
		emailService: emailService,
	}
}

// RequestEmailChange initiates an email change request with token generation
func (p *ProfileService) RequestEmailChange(userID uint, newEmail string) (string, error) {
	// Validate the new email format
	if err := p.emailService.ValidateEmail(newEmail); err != nil {
		return "", fmt.Errorf("invalid email format: %w", err)
	}

	// Check if email is already in use
	var existingUser models.User
	if err := p.db.Where("email = ?", newEmail).First(&existingUser).Error; err == nil {
		return "", fmt.Errorf("email already in use")
	} else if err != gorm.ErrRecordNotFound {
		return "", fmt.Errorf("database error: %w", err)
	}

	// Get current user
	var user models.User
	if err := p.db.First(&user, userID).Error; err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	// Generate confirmation token
	token, err := p.GenerateConfirmationToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Create profile change record
	profileChange := &models.ProfileChange{
		UserID:     userID,
		ChangeType: "email",
		OldValue:   user.Email,
		NewValue:   newEmail,
		Token:      token,
		ExpiresAt:  time.Now().Add(24 * time.Hour), // 24 hour expiration
		Confirmed:  false,
	}

	if err := p.db.Create(profileChange).Error; err != nil {
		return "", fmt.Errorf("failed to create profile change record: %w", err)
	}

	// Send confirmation email to the new email address
	tempUser := user
	tempUser.Email = newEmail // Temporarily set new email for sending confirmation
	if err := p.emailService.SendConfirmationEmail(&tempUser, "email", token); err != nil {
		// Log error but don't fail the request - email sending issues shouldn't block the process
		fmt.Printf("Failed to send confirmation email: %v\n", err)
	}

	return token, nil
}

// ConfirmEmailChange processes an email change confirmation token
func (p *ProfileService) ConfirmEmailChange(token string) error {
	// Validate and get the profile change
	profileChange, err := p.ValidateConfirmationToken(token)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	if profileChange.ChangeType != "email" {
		return fmt.Errorf("token is not for email change")
	}

	// Update user email
	if err := p.db.Model(&models.User{}).Where("id = ?", profileChange.UserID).Update("email", profileChange.NewValue).Error; err != nil {
		return fmt.Errorf("failed to update user email: %w", err)
	}

	// Mark profile change as confirmed
	profileChange.Confirmed = true
	if err := p.db.Save(profileChange).Error; err != nil {
		return fmt.Errorf("failed to mark change as confirmed: %w", err)
	}

	// Send notification email to old address
	var user models.User
	if err := p.db.First(&user, profileChange.UserID).Error; err == nil {
		// Create a temporary user object with the old email for notification
		oldUser := models.User{
			Email:     profileChange.OldValue,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
		}
		if err := p.emailService.SendNotificationEmail(&oldUser, "Email Changed",
			fmt.Sprintf("Your email has been successfully changed to %s", profileChange.NewValue)); err != nil {
			// Log error but don't fail the confirmation
			fmt.Printf("Failed to send notification email: %v\n", err)
		}
	}

	return nil
}

// UpdateProfile handles profile updates with confirmation requirements for sensitive changes
func (p *ProfileService) UpdateProfile(userID uint, updates map[string]interface{}) error {
	// Define sensitive fields that require confirmation
	sensitiveFields := map[string]bool{
		"email":    true,
		"password": true,
	}

	// Check if any sensitive fields are being updated
	requiresConfirmation := false
	for field := range updates {
		if sensitiveFields[field] {
			requiresConfirmation = true
			break
		}
	}

	if requiresConfirmation {
		return fmt.Errorf("sensitive profile changes require confirmation - use specific methods")
	}

	// Update non-sensitive fields directly
	allowedFields := []string{"bio", "avatar_url", "location_enabled", "privacy_level"}
	filteredUpdates := make(map[string]interface{})

	for _, field := range allowedFields {
		if value, exists := updates[field]; exists {
			filteredUpdates[field] = value
		}
	}

	if len(filteredUpdates) == 0 {
		return fmt.Errorf("no valid fields to update")
	}

	if err := p.db.Model(&models.User{}).Where("id = ?", userID).Updates(filteredUpdates).Error; err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

// GenerateConfirmationToken creates a secure random token for confirmations
func (p *ProfileService) GenerateConfirmationToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateConfirmationToken validates a confirmation token and returns the associated profile change
func (p *ProfileService) ValidateConfirmationToken(token string) (*models.ProfileChange, error) {
	var profileChange models.ProfileChange
	if err := p.db.Where("token = ?", token).First(&profileChange).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check if token has expired
	if time.Now().After(profileChange.ExpiresAt) {
		return nil, fmt.Errorf("token has expired")
	}

	// Check if token has already been used
	if profileChange.Confirmed {
		return nil, fmt.Errorf("token has already been used")
	}

	return &profileChange, nil
}

// CleanupExpiredTokens removes expired confirmation tokens from the database
func (p *ProfileService) CleanupExpiredTokens() error {
	result := p.db.Where("expires_at < ? AND confirmed = ?", time.Now(), false).
		Delete(&models.ProfileChange{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
	}

	fmt.Printf("Cleaned up %d expired confirmation tokens\n", result.RowsAffected)
	return nil
}

// GetPendingChanges returns all pending profile changes for a user
func (p *ProfileService) GetPendingChanges(userID uint) ([]models.ProfileChange, error) {
	var changes []models.ProfileChange
	err := p.db.Where("user_id = ? AND confirmed = ? AND expires_at > ?",
		userID, false, time.Now()).Find(&changes).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get pending changes: %w", err)
	}

	return changes, nil
}

// CancelPendingChange cancels a pending profile change
func (p *ProfileService) CancelPendingChange(userID uint, token string) error {
	result := p.db.Where("user_id = ? AND token = ? AND confirmed = ?",
		userID, token, false).Delete(&models.ProfileChange{})

	if result.Error != nil {
		return fmt.Errorf("failed to cancel pending change: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no pending change found with the given token")
	}

	return nil
}

// ConfirmProfileChange processes a general profile change confirmation
func (p *ProfileService) ConfirmProfileChange(token string) error {
	// Find and validate the profile change record
	profileChange, err := p.ValidateConfirmationToken(token)
	if err != nil {
		return fmt.Errorf("invalid confirmation token: %w", err)
	}

	// Handle different types of changes
	switch profileChange.ChangeType {
	case "email":
		return p.ConfirmEmailChange(token)
	case "profile":
		// Mark the change as confirmed
		if err := p.db.Model(profileChange).
			Update("confirmed", true).Error; err != nil {
			return fmt.Errorf("failed to mark change as confirmed: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported change type: %s", profileChange.ChangeType)
	}
}

// GetProfileChangeStats returns statistics about profile changes
func (p *ProfileService) GetProfileChangeStats() (map[string]interface{}, error) {
	var totalChanges int64
	var pendingChanges int64
	var confirmedChanges int64

	// Count total changes
	if err := p.db.Model(&models.ProfileChange{}).Count(&totalChanges).Error; err != nil {
		return nil, fmt.Errorf("failed to count total changes: %w", err)
	}

	// Count pending changes
	if err := p.db.Model(&models.ProfileChange{}).
		Where("confirmed = ? AND expires_at > ?", false, time.Now()).
		Count(&pendingChanges).Error; err != nil {
		return nil, fmt.Errorf("failed to count pending changes: %w", err)
	}

	// Count confirmed changes
	if err := p.db.Model(&models.ProfileChange{}).
		Where("confirmed = ?", true).
		Count(&confirmedChanges).Error; err != nil {
		return nil, fmt.Errorf("failed to count confirmed changes: %w", err)
	}

	stats := map[string]interface{}{
		"total_changes":     totalChanges,
		"pending_changes":   pendingChanges,
		"confirmed_changes": confirmedChanges,
	}

	return stats, nil
}
