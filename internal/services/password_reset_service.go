// Package services provides business logic services for the dinnerclub application
package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// PasswordResetEmailer defines the interface for sending password reset emails
type PasswordResetEmailer interface {
	SendPasswordResetEmail(user *models.User, tempPassword string) error
}

// PasswordResetService handles password reset functionality
type PasswordResetService struct {
	db           *gorm.DB
	emailService PasswordResetEmailer
	rateLimiter  *ResetRateLimiter
}

// ResetRateLimiter implements rate limiting for password reset requests
type ResetRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int           // Maximum requests per window
	window   time.Duration // Time window for rate limiting
}

// NewPasswordResetService creates a new password reset service instance
func NewPasswordResetService(db *gorm.DB, emailService PasswordResetEmailer) *PasswordResetService {
	return &PasswordResetService{
		db:           db,
		emailService: emailService,
		rateLimiter:  NewResetRateLimiter(3, 1*time.Hour), // 3 requests per hour
	}
}

// NewResetRateLimiter creates a new rate limiter for password resets
func NewResetRateLimiter(limit int, window time.Duration) *ResetRateLimiter {
	return &ResetRateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RequestPasswordReset initiates a password reset request
func (p *PasswordResetService) RequestPasswordReset(email string) error {
	// Check rate limiting
	if !p.rateLimiter.Allow(email) {
		return fmt.Errorf("too many password reset requests. Please try again later")
	}

	// Find user by email
	var user models.User
	if err := p.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Don't reveal if email exists - security best practice
			return nil
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Generate temporary password
	tempPassword, err := p.GenerateTemporaryPassword()
	if err != nil {
		return fmt.Errorf("failed to generate temporary password: %w", err)
	}

	// Hash the temporary password for storage
	hashedTempPassword, err := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash temporary password: %w", err)
	}

	// Generate secure token
	token, err := p.GenerateResetToken()
	if err != nil {
		return fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Create password reset record with 1 hour expiration
	passwordReset := models.PasswordReset{
		UserID:       user.ID,
		Token:        token,
		TempPassword: string(hashedTempPassword),
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Used:         false,
	}

	if err := p.db.Create(&passwordReset).Error; err != nil {
		return fmt.Errorf("failed to create password reset record: %w", err)
	}

	// Send password reset email with temporary password
	if err := p.emailService.SendPasswordResetEmail(&user, tempPassword); err != nil {
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	return nil
}

// GenerateTemporaryPassword generates a cryptographically secure random temporary password
func (p *PasswordResetService) GenerateTemporaryPassword() (string, error) {
	// Generate 16 bytes of random data (128 bits)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Encode to base64 and take first 20 characters for a readable password
	// This gives us a password with uppercase, lowercase, numbers, and special chars
	password := base64.URLEncoding.EncodeToString(bytes)[:20]

	return password, nil
}

// GenerateResetToken generates a secure token for password reset
func (p *PasswordResetService) GenerateResetToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// ValidateResetToken validates a password reset token and returns the reset record
func (p *PasswordResetService) ValidateResetToken(token string) (*models.PasswordReset, error) {
	var passwordReset models.PasswordReset
	err := p.db.Where("token = ? AND used = ? AND expires_at > ?", token, false, time.Now()).
		First(&passwordReset).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("invalid or expired reset token")
		}
		return nil, fmt.Errorf("failed to validate reset token: %w", err)
	}

	return &passwordReset, nil
}

// ValidateTemporaryPassword validates a temporary password against the stored hash
func (p *PasswordResetService) ValidateTemporaryPassword(userID uint, tempPassword string) (*models.PasswordReset, error) {
	// Find active password reset for user
	var passwordReset models.PasswordReset
	err := p.db.Where("user_id = ? AND used = ? AND expires_at > ?", userID, false, time.Now()).
		Order("created_at DESC").
		First(&passwordReset).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active password reset found")
		}
		return nil, fmt.Errorf("failed to find password reset: %w", err)
	}

	// Validate the temporary password
	err = bcrypt.CompareHashAndPassword([]byte(passwordReset.TempPassword), []byte(tempPassword))
	if err != nil {
		return nil, fmt.Errorf("invalid temporary password")
	}

	return &passwordReset, nil
}

// CompletePasswordReset completes the password reset process
func (p *PasswordResetService) CompletePasswordReset(resetID uint, newPassword string) error {
	// Get the password reset record
	var passwordReset models.PasswordReset
	if err := p.db.First(&passwordReset, resetID).Error; err != nil {
		return fmt.Errorf("failed to find password reset: %w", err)
	}

	// Verify it hasn't been used and hasn't expired
	if passwordReset.Used {
		return fmt.Errorf("password reset has already been used")
	}

	if time.Now().After(passwordReset.ExpiresAt) {
		return fmt.Errorf("password reset has expired")
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update user's password
	if err := p.db.Model(&models.User{}).Where("id = ?", passwordReset.UserID).
		Update("password", string(hashedPassword)).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark reset as used
	passwordReset.Used = true
	if err := p.db.Save(&passwordReset).Error; err != nil {
		return fmt.Errorf("failed to mark reset as used: %w", err)
	}

	return nil
}

// LoginWithTemporaryPassword handles login with temporary password and enforces password change
func (p *PasswordResetService) LoginWithTemporaryPassword(userID uint, tempPassword string) (*models.PasswordReset, bool, error) {
	// Validate the temporary password
	passwordReset, err := p.ValidateTemporaryPassword(userID, tempPassword)
	if err != nil {
		return nil, false, err
	}

	// Return the reset record and flag indicating password change is required
	return passwordReset, true, nil
}

// RequiresPasswordChange checks if a user needs to change their password
func (p *PasswordResetService) RequiresPasswordChange(userID uint) (bool, error) {
	// Check if there's an active, unused password reset for this user
	var count int64
	err := p.db.Model(&models.PasswordReset{}).
		Where("user_id = ? AND used = ? AND expires_at > ?", userID, false, time.Now()).
		Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check password reset status: %w", err)
	}

	return count > 0, nil
}

// InvalidateAllSessions invalidates all sessions for a user after password reset
// This method should be called by the authentication service
func (p *PasswordResetService) InvalidateAllSessions(userID uint) error {
	// In a real implementation, this would:
	// 1. Delete all JWT tokens from a token blacklist/whitelist
	// 2. Update a session version number in the user record
	// 3. Clear any session storage (Redis, database, etc.)

	// For now, we'll update the user's LastLoginAt to force re-authentication
	// The authentication middleware should check this timestamp
	now := time.Now()
	if err := p.db.Model(&models.User{}).Where("id = ?", userID).
		Update("last_login_at", now).Error; err != nil {
		return fmt.Errorf("failed to invalidate sessions: %w", err)
	}

	return nil
}

// CompletePasswordResetWithSessionInvalidation completes password reset and invalidates sessions
func (p *PasswordResetService) CompletePasswordResetWithSessionInvalidation(resetID uint, newPassword string) error {
	// Get the password reset record first to get user ID
	var passwordReset models.PasswordReset
	if err := p.db.First(&passwordReset, resetID).Error; err != nil {
		return fmt.Errorf("failed to find password reset: %w", err)
	}

	// Complete the password reset
	if err := p.CompletePasswordReset(resetID, newPassword); err != nil {
		return err
	}

	// Invalidate all existing sessions for security
	if err := p.InvalidateAllSessions(passwordReset.UserID); err != nil {
		// Log the error but don't fail the password reset
		// The password has been changed successfully
		return fmt.Errorf("password reset completed but session invalidation failed: %w", err)
	}

	return nil
}

// CleanupExpiredResets removes expired password reset records
func (p *PasswordResetService) CleanupExpiredResets() error {
	result := p.db.Where("expires_at < ?", time.Now()).Delete(&models.PasswordReset{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired resets: %w", result.Error)
	}
	return nil
}

// Allow checks if a request is allowed under rate limiting
func (r *ResetRateLimiter) Allow(identifier string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Get existing requests for this identifier
	requests := r.requests[identifier]

	// Filter out requests outside the time window
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if limit is exceeded
	if len(validRequests) >= r.limit {
		r.requests[identifier] = validRequests
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	r.requests[identifier] = validRequests

	return true
}

// Reset clears rate limiting data for an identifier
func (r *ResetRateLimiter) Reset(identifier string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.requests, identifier)
}

// GetRemainingAttempts returns the number of remaining attempts for an identifier
func (r *ResetRateLimiter) GetRemainingAttempts(identifier string) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	requests := r.requests[identifier]
	var validRequests int
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests++
		}
	}

	remaining := r.limit - validRequests
	if remaining < 0 {
		return 0
	}
	return remaining
}
