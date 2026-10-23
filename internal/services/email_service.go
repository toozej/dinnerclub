// Package services provides business logic services for the dinnerclub application
package services

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// EmailService handles SMTP email sending and forwarding functionality
type EmailService struct {
	db           *gorm.DB
	smtpHost     string
	smtpPort     int
	smtpUsername string
	smtpPassword string
	fromAddress  string
	cityCode     string
}

// NewEmailService creates a new email service instance
func NewEmailService(db *gorm.DB, smtpHost string, smtpPort int, smtpUsername, smtpPassword, fromAddress, cityCode string) *EmailService {
	return &EmailService{
		db:           db,
		smtpHost:     smtpHost,
		smtpPort:     smtpPort,
		smtpUsername: smtpUsername,
		smtpPassword: smtpPassword,
		fromAddress:  fromAddress,
		cityCode:     cityCode,
	}
}

// EmailTemplate represents data for email templates
type EmailTemplate struct {
	ToEmail      string
	ToName       string
	Subject      string
	Body         string
	Token        string
	TempPassword string
	CityCode     string
	BaseURL      string
}

// SendConfirmationEmail sends a confirmation email for profile changes
func (e *EmailService) SendConfirmationEmail(user *models.User, changeType string, token string) error {
	var subject, body string
	baseURL := fmt.Sprintf("https://%s.dinner.club", e.cityCode)

	switch changeType {
	case "email":
		subject = fmt.Sprintf("[%s Dinner Club] Confirm Email Change", strings.ToUpper(e.cityCode))
		body = fmt.Sprintf(`Hi %s,

You recently requested to change your email address on %s Dinner Club.

To confirm this change, please click the link below:
%s/confirm-email?token=%s

This link will expire in 24 hours.

If you didn't request this change, please ignore this email.

Best regards,
The %s Dinner Club Team`,
			user.Firstname,
			strings.ToUpper(e.cityCode),
			baseURL,
			token,
			strings.ToUpper(e.cityCode))
	case "password":
		subject = fmt.Sprintf("[%s Dinner Club] Password Changed", strings.ToUpper(e.cityCode))
		body = fmt.Sprintf(`Hi %s,

Your password has been successfully changed on %s Dinner Club.

If you didn't make this change, please contact us immediately.

Best regards,
The %s Dinner Club Team`,
			user.Firstname,
			strings.ToUpper(e.cityCode),
			strings.ToUpper(e.cityCode))
	case "profile":
		subject = fmt.Sprintf("[%s Dinner Club] Confirm Profile Changes", strings.ToUpper(e.cityCode))
		body = fmt.Sprintf(`Hi %s,

You recently made changes to your profile on %s Dinner Club.

To confirm these changes, please click the link below:
%s/confirm-profile?token=%s

This link will expire in 24 hours.

If you didn't request these changes, please ignore this email.

Best regards,
The %s Dinner Club Team`,
			user.Firstname,
			strings.ToUpper(e.cityCode),
			baseURL,
			token,
			strings.ToUpper(e.cityCode))
	default:
		return fmt.Errorf("unsupported change type: %s", changeType)
	}

	return e.sendEmail(user.Email, user.Firstname+" "+user.Lastname, subject, body)
}

// SendPasswordResetEmail sends a password reset email with temporary password
func (e *EmailService) SendPasswordResetEmail(user *models.User, tempPassword string) error {
	subject := fmt.Sprintf("[%s Dinner Club] Password Reset", strings.ToUpper(e.cityCode))
	baseURL := fmt.Sprintf("https://%s.dinner.club", e.cityCode)

	body := fmt.Sprintf(`Hi %s,

You requested a password reset for your %s Dinner Club account.

Your temporary password is: %s

Please log in using this temporary password and change it immediately:
%s/login

This temporary password will expire in 1 hour.

If you didn't request this reset, please ignore this email.

Best regards,
The %s Dinner Club Team`,
		user.Firstname,
		strings.ToUpper(e.cityCode),
		tempPassword,
		baseURL,
		strings.ToUpper(e.cityCode))

	return e.sendEmail(user.Email, user.Firstname+" "+user.Lastname, subject, body)
}

// SendNotificationEmail sends a general notification email
func (e *EmailService) SendNotificationEmail(user *models.User, subject, body string) error {
	fullSubject := fmt.Sprintf("[%s Dinner Club] %s", strings.ToUpper(e.cityCode), subject)
	return e.sendEmail(user.Email, user.Firstname+" "+user.Lastname, fullSubject, body)
}

// sendEmail sends an email using SMTP with TLS
func (e *EmailService) sendEmail(toEmail, toName, subject, body string) error {
	// Create message
	message := fmt.Sprintf("To: %s <%s>\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", toName, toEmail, subject, body)

	// Connect to SMTP server
	serverName := fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort)

	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName: e.smtpHost,
	}

	// Connect with TLS
	conn, err := tls.Dial("tcp", serverName, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, e.smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	auth := smtp.PlainAuth("", e.smtpUsername, e.smtpPassword, e.smtpHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}

	// Set sender
	if err := client.Mail(e.fromAddress); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err := client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = writer.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close message writer: %w", err)
	}

	return nil
}

// SetupEmailForwarding creates an email forwarding rule for a user
func (e *EmailService) SetupEmailForwarding(username, email string) error {
	fromAddress := fmt.Sprintf("%s@%s.dinner.club", username, e.cityCode)

	// Check if forwarding rule already exists
	var existingRule models.EmailForwardingRule
	err := e.db.Where("username = ?", username).First(&existingRule).Error

	if err == gorm.ErrRecordNotFound {
		// Create new forwarding rule
		rule := models.EmailForwardingRule{
			Username:    username,
			FromAddress: fromAddress,
			ToAddress:   email,
			Active:      true,
		}

		if err := e.db.Create(&rule).Error; err != nil {
			return fmt.Errorf("failed to create email forwarding rule: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check existing forwarding rule: %w", err)
	} else {
		// Update existing rule
		existingRule.ToAddress = email
		existingRule.Active = true

		if err := e.db.Save(&existingRule).Error; err != nil {
			return fmt.Errorf("failed to update email forwarding rule: %w", err)
		}
	}

	return nil
}

// RemoveEmailForwarding removes an email forwarding rule for a user
func (e *EmailService) RemoveEmailForwarding(username string) error {
	result := e.db.Where("username = ?", username).Delete(&models.EmailForwardingRule{})
	if result.Error != nil {
		return fmt.Errorf("failed to remove email forwarding rule: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no forwarding rule found for username: %s", username)
	}

	return nil
}

// GenerateSecureToken generates a cryptographically secure random token
func (e *EmailService) GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// ValidateEmail performs comprehensive email format validation
func (e *EmailService) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email address is required")
	}

	// Check length limits
	if len(email) > 254 {
		return fmt.Errorf("email address too long (maximum 254 characters)")
	}

	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email format: missing @ symbol")
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email format: multiple @ symbols")
	}

	localPart := parts[0]
	domainPart := parts[1]

	// Validate local part
	if localPart == "" {
		return fmt.Errorf("invalid email format: missing local part")
	}

	if len(localPart) > 64 {
		return fmt.Errorf("invalid email format: local part too long (maximum 64 characters)")
	}

	// Check for invalid characters in local part
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return fmt.Errorf("invalid email format: local part cannot start or end with a dot")
	}

	if strings.Contains(localPart, "..") {
		return fmt.Errorf("invalid email format: consecutive dots not allowed in local part")
	}

	// Validate domain part
	if domainPart == "" {
		return fmt.Errorf("invalid email format: missing domain part")
	}

	if len(domainPart) > 253 {
		return fmt.Errorf("invalid email format: domain part too long (maximum 253 characters)")
	}

	if !strings.Contains(domainPart, ".") {
		return fmt.Errorf("invalid email format: domain missing TLD")
	}

	// Check for valid domain format
	if strings.HasPrefix(domainPart, ".") || strings.HasSuffix(domainPart, ".") {
		return fmt.Errorf("invalid email format: domain cannot start or end with a dot")
	}

	if strings.Contains(domainPart, "..") {
		return fmt.Errorf("invalid email format: consecutive dots not allowed in domain")
	}

	// Check for valid characters in domain
	domainLabels := strings.Split(domainPart, ".")
	for _, label := range domainLabels {
		if label == "" {
			return fmt.Errorf("invalid email format: empty domain label")
		}

		if len(label) > 63 {
			return fmt.Errorf("invalid email format: domain label too long (maximum 63 characters)")
		}

		if strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return fmt.Errorf("invalid email format: domain label cannot start or end with hyphen")
		}
	}

	return nil
}

// CheckEmailDeliverability performs basic deliverability checks
func (e *EmailService) CheckEmailDeliverability(email string) error {
	// Extract domain part first for blocked-domain checks
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		domain := parts[1]

		// Check if domain is in our blocked list before format validation
		// so the error message is consistently "not deliverable"
		blockedDomains := []string{
			"example.com",
			"test.com",
			"localhost",
			"invalid",
			"fake.com",
		}

		for _, blocked := range blockedDomains {
			if strings.EqualFold(domain, blocked) {
				return fmt.Errorf("email domain is not deliverable: %s", domain)
			}
		}
	}

	// Then validate the format for well-formed addresses
	if err := e.ValidateEmail(email); err != nil {
		return err
	}

	// In a real implementation, you might want to:
	// - Check MX records for the domain
	// - Verify the domain exists
	// - Check against known disposable email providers
	// For now, we'll do basic validation

	return nil
}

// ValidateEmailForForwarding validates email specifically for forwarding setup
func (e *EmailService) ValidateEmailForForwarding(email string) error {
	// First check basic deliverability
	if err := e.CheckEmailDeliverability(email); err != nil {
		return err
	}

	// Then check forwarding-specific rules
	return e.ValidateForwardingDomain(email)
}

// LogEmailError logs email-related errors for monitoring with structured logging
func (e *EmailService) LogEmailError(operation, email string, err error) {
	// In a real application, this would use a proper structured logging system
	// like logrus, zap, or the standard log/slog package
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] EMAIL ERROR [%s] for %s: %v\n", timestamp, operation, email, err)

	// You might also want to:
	// - Send alerts for critical errors
	// - Store error metrics for monitoring
	// - Log to external systems like Elasticsearch or CloudWatch
}

// HandleEmailError provides centralized error handling for email operations
func (e *EmailService) HandleEmailError(operation, email string, err error) error {
	// Log the error
	e.LogEmailError(operation, email, err)

	// Determine if this is a retryable error
	if e.isRetryableError(err) {
		return fmt.Errorf("temporary email error for %s: %w (operation: %s)", email, err, operation)
	}

	// For permanent errors, return a user-friendly message
	return fmt.Errorf("email operation failed for %s: %s", email, e.getUserFriendlyErrorMessage(err))
}

// isRetryableError determines if an error is temporary and worth retrying
func (e *EmailService) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := strings.ToLower(err.Error())

	// Network-related errors that might be temporary
	retryableErrors := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"network unreachable",
		"no such host", // DNS might be temporarily down
		"connection reset",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errorStr, retryable) {
			return true
		}
	}

	return false
}

// getUserFriendlyErrorMessage converts technical errors to user-friendly messages
func (e *EmailService) getUserFriendlyErrorMessage(err error) string {
	if err == nil {
		return "Unknown error"
	}

	errorStr := strings.ToLower(err.Error())

	// Map technical errors to user-friendly messages
	if strings.Contains(errorStr, "authentication failed") {
		return "Email service authentication failed. Please contact support."
	}

	if strings.Contains(errorStr, "connection refused") || strings.Contains(errorStr, "timeout") {
		return "Email service is temporarily unavailable. Please try again later."
	}

	if strings.Contains(errorStr, "invalid email") || strings.Contains(errorStr, "email format") {
		return "Please check your email address format."
	}

	if strings.Contains(errorStr, "not deliverable") || strings.Contains(errorStr, "disposable") {
		return "This email address cannot be used. Please use a different email address."
	}

	if strings.Contains(errorStr, "rate limit") {
		return "Too many email requests. Please wait a moment and try again."
	}

	// Default message for unknown errors
	return "An error occurred while processing your email. Please try again or contact support."
}

// NotifyForwardingFailure sends a notification to the user about forwarding failures
func (e *EmailService) NotifyForwardingFailure(username, originalEmail string, err error) error {
	// Get the user's information to send notification
	var user models.User
	if dbErr := e.db.Where("username = ?", username).First(&user).Error; dbErr != nil {
		e.LogEmailError("forwarding_failure_notification", username, fmt.Errorf("failed to find user for notification: %w", dbErr))
		return fmt.Errorf("failed to find user for forwarding failure notification: %w", dbErr)
	}

	subject := "Email Forwarding Issue"
	body := fmt.Sprintf(`Hi %s,

We encountered an issue forwarding an email to your address (%s).

Original email details:
- Intended recipient: %s@%s.dinner.club
- Error: %s

Please check your email address in your profile settings or contact support if this issue persists.

Best regards,
The %s Dinner Club Team`,
		user.Firstname,
		user.Email,
		username,
		e.cityCode,
		e.getUserFriendlyErrorMessage(err),
		strings.ToUpper(e.cityCode))

	// Send the notification email
	notifyErr := e.SendNotificationEmail(&user, subject, body)
	if notifyErr != nil {
		e.LogEmailError("forwarding_failure_notification", user.Email, notifyErr)
		return fmt.Errorf("failed to send forwarding failure notification: %w", notifyErr)
	}

	return nil
}

// GetEmailForwardingRule retrieves the forwarding rule for a username
func (e *EmailService) GetEmailForwardingRule(username string) (*models.EmailForwardingRule, error) {
	var rule models.EmailForwardingRule
	err := e.db.Where("username = ? AND active = ?", username, true).First(&rule).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("no active forwarding rule found for username: %s", username)
		}
		return nil, fmt.Errorf("failed to get forwarding rule: %w", err)
	}
	return &rule, nil
}

// UpdateEmailForwardingRule updates an existing forwarding rule
func (e *EmailService) UpdateEmailForwardingRule(username, newEmail string) error {
	result := e.db.Model(&models.EmailForwardingRule{}).
		Where("username = ?", username).
		Updates(map[string]interface{}{
			"to_address": newEmail,
			"active":     true,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update forwarding rule: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no forwarding rule found to update for username: %s", username)
	}

	return nil
}

// DeactivateEmailForwarding deactivates a forwarding rule without deleting it
func (e *EmailService) DeactivateEmailForwarding(username string) error {
	result := e.db.Model(&models.EmailForwardingRule{}).
		Where("username = ?", username).
		Update("active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to deactivate forwarding rule: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no forwarding rule found for username: %s", username)
	}

	return nil
}

// ListActiveForwardingRules returns all active forwarding rules
func (e *EmailService) ListActiveForwardingRules() ([]models.EmailForwardingRule, error) {
	var rules []models.EmailForwardingRule
	err := e.db.Where("active = ?", true).Find(&rules).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list active forwarding rules: %w", err)
	}
	return rules, nil
}

// ProcessIncomingEmail processes an incoming email for forwarding
// This would be called by an SMTP relay service
func (e *EmailService) ProcessIncomingEmail(toAddress, fromAddress, subject, body string) error {
	// Extract username from the to address (username@citycode.dinner.club)
	parts := strings.Split(toAddress, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email address format: %s", toAddress)
	}

	username := parts[0]
	domain := parts[1]

	// Verify this is for our domain
	expectedDomain := fmt.Sprintf("%s.dinner.club", e.cityCode)
	if domain != expectedDomain {
		return fmt.Errorf("email not for our domain: %s (expected %s)", domain, expectedDomain)
	}

	// Get forwarding rule
	rule, err := e.GetEmailForwardingRule(username)
	if err != nil {
		e.LogEmailError("forwarding_lookup", toAddress, err)
		return fmt.Errorf("no forwarding rule found for %s: %w", username, err)
	}

	// Forward the email
	err = e.forwardEmail(rule.ToAddress, fromAddress, subject, body, toAddress)
	if err != nil {
		// Enhanced error handling: notify user of forwarding failure
		notifyErr := e.NotifyForwardingFailure(username, rule.ToAddress, err)
		if notifyErr != nil {
			e.LogEmailError("forwarding_failure_notification", rule.ToAddress, notifyErr)
		}

		handledErr := e.HandleEmailError("forwarding_send", rule.ToAddress, err)
		return fmt.Errorf("failed to forward email: %w", handledErr)
	}

	return nil
}

// forwardEmail forwards an email to the target address
func (e *EmailService) forwardEmail(toAddress, originalFrom, subject, body, originalTo string) error {
	// Create forwarded message with headers indicating it's forwarded
	forwardedSubject := fmt.Sprintf("Fwd: %s", subject)
	forwardedBody := fmt.Sprintf(`---------- Forwarded message ----------
From: %s
To: %s
Subject: %s

%s`, originalFrom, originalTo, subject, body)

	// Send the forwarded email
	return e.sendEmail(toAddress, "", forwardedSubject, forwardedBody)
}

// ValidateForwardingDomain checks if an email domain is allowed for forwarding
func (e *EmailService) ValidateForwardingDomain(email string) error {
	if err := e.ValidateEmail(email); err != nil {
		return err
	}

	// Extract domain
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email format")
	}

	domain := strings.ToLower(parts[1])

	// Block forwarding to our own domain to prevent loops
	ourDomain := fmt.Sprintf("%s.dinner.club", e.cityCode)
	if domain == ourDomain {
		return fmt.Errorf("cannot forward to the same domain: %s", domain)
	}

	// Block common temporary/disposable email domains
	blockedDomains := []string{
		"10minutemail.com",
		"tempmail.org",
		"guerrillamail.com",
		"mailinator.com",
		"throwaway.email",
	}

	for _, blocked := range blockedDomains {
		if domain == blocked {
			return fmt.Errorf("forwarding to disposable email domains is not allowed: %s", domain)
		}
	}

	return nil
}

// GetForwardingStats returns statistics about email forwarding
func (e *EmailService) GetForwardingStats() (map[string]interface{}, error) {
	var totalRules int64
	var activeRules int64

	// Count total rules
	if err := e.db.Model(&models.EmailForwardingRule{}).Count(&totalRules).Error; err != nil {
		return nil, fmt.Errorf("failed to count total rules: %w", err)
	}

	// Count active rules
	if err := e.db.Model(&models.EmailForwardingRule{}).Where("active = ?", true).Count(&activeRules).Error; err != nil {
		return nil, fmt.Errorf("failed to count active rules: %w", err)
	}

	stats := map[string]interface{}{
		"total_rules":  totalRules,
		"active_rules": activeRules,
		"city_code":    e.cityCode,
		"domain":       fmt.Sprintf("%s.dinner.club", e.cityCode),
	}

	return stats, nil
}
