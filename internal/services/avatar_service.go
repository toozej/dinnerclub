// Package services provides business logic services for the dinnerclub application
package services

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
	"golang.org/x/image/draw"
)

// AvatarService handles user avatar management including Gravatar and custom uploads
type AvatarService struct {
	db           *gorm.DB
	uploadPath   string
	maxFileSize  int64
	allowedTypes []string
}

// NewAvatarService creates a new avatar service instance
func NewAvatarService(db *gorm.DB, uploadPath string, maxFileSize int64, allowedTypes []string) *AvatarService {
	return &AvatarService{
		db:           db,
		uploadPath:   uploadPath,
		maxFileSize:  maxFileSize,
		allowedTypes: allowedTypes,
	}
}

// GetGravatarURL generates a Gravatar URL for an email address
func (a *AvatarService) GetGravatarURL(email string, size int) string {
	// Normalize email: trim whitespace and convert to lowercase
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))

	// Generate MD5 hash of the email
	hash := md5.Sum([]byte(normalizedEmail))
	hashString := hex.EncodeToString(hash[:])

	// Construct Gravatar URL with size parameter
	// d=mp sets default to mystery person icon
	// r=pg sets rating to PG (family-friendly)
	return fmt.Sprintf("https://www.gravatar.com/avatar/%s?s=%d&d=mp&r=pg", hashString, size)
}

// UploadCustomAvatar handles custom avatar upload with validation and processing
func (a *AvatarService) UploadCustomAvatar(userID uint, file multipart.File, header *multipart.FileHeader) (string, error) {
	// Validate file size
	if header.Size > a.maxFileSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", a.maxFileSize)
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !a.isAllowedType(contentType) {
		return "", fmt.Errorf("file type %s not allowed. Allowed types: %v", contentType, a.allowedTypes)
	}

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Validate format matches content type
	if !a.validateFormat(format, contentType) {
		return "", fmt.Errorf("image format %s does not match content type %s", format, contentType)
	}

	// Resize image to standard size (300x300)
	resizedImg := a.ResizeImage(img, 300, 300)

	// Generate filename
	filename := fmt.Sprintf("avatar_%d.%s", userID, format)
	filepath := filepath.Join(a.uploadPath, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(a.uploadPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Save resized image
	outFile, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode and save based on format
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(outFile, resizedImg, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(outFile, resizedImg)
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}

	if err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	// Generate URL path for the avatar
	avatarURL := fmt.Sprintf("/uploads/avatars/%s", filename)

	// Update user's avatar URL in database
	if err := a.db.Model(&models.User{}).Where("id = ?", userID).Update("avatar_url", avatarURL).Error; err != nil {
		// Clean up uploaded file if database update fails
		os.Remove(filepath)
		return "", fmt.Errorf("failed to update user avatar URL: %w", err)
	}

	return avatarURL, nil
}

// ResizeImage resizes an image to the specified width and height
func (a *AvatarService) ResizeImage(src image.Image, width, height int) image.Image {
	// Create destination image
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Use high-quality scaling algorithm
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	return dst
}

// GetUserAvatarURL returns the appropriate avatar URL for a user
func (a *AvatarService) GetUserAvatarURL(user *models.User, size int) string {
	// If user has a custom avatar, return it
	if user.AvatarURL != "" {
		return user.AvatarURL
	}

	// Otherwise, return Gravatar URL
	return a.GetGravatarURL(user.Email, size)
}

// DeleteCustomAvatar removes a user's custom avatar
func (a *AvatarService) DeleteCustomAvatar(userID uint) error {
	// Get user to find avatar URL
	var user models.User
	if err := a.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// If no custom avatar, nothing to delete
	if user.AvatarURL == "" {
		return nil
	}

	// Extract filename from URL
	filename := filepath.Base(user.AvatarURL)
	filepath := filepath.Join(a.uploadPath, filename)

	// Delete file if it exists
	if _, err := os.Stat(filepath); err == nil {
		if err := os.Remove(filepath); err != nil {
			return fmt.Errorf("failed to delete avatar file: %w", err)
		}
	}

	// Clear avatar URL in database
	if err := a.db.Model(&user).Update("avatar_url", "").Error; err != nil {
		return fmt.Errorf("failed to clear avatar URL: %w", err)
	}

	return nil
}

// isAllowedType checks if a content type is in the allowed list
func (a *AvatarService) isAllowedType(contentType string) bool {
	for _, allowed := range a.allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

// validateFormat ensures the image format matches the content type
func (a *AvatarService) validateFormat(format, contentType string) bool {
	formatMap := map[string]string{
		"jpeg": "image/jpeg",
		"jpg":  "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
	}

	expectedContentType, ok := formatMap[format]
	if !ok {
		return false
	}

	return expectedContentType == contentType
}

// UpdateAvatarOnEmailChange updates Gravatar when user changes email
func (a *AvatarService) UpdateAvatarOnEmailChange(userID uint, newEmail string) error {
	// Get user
	var user models.User
	if err := a.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// If user has a custom avatar, no need to update
	if user.AvatarURL != "" && !strings.Contains(user.AvatarURL, "gravatar.com") {
		return nil
	}

	// Gravatar URLs are generated on-the-fly based on email
	// No database update needed as GetUserAvatarURL will use the new email
	return nil
}

// GetDefaultAvatarURL returns a default placeholder avatar URL
func (a *AvatarService) GetDefaultAvatarURL() string {
	// Return a default mystery person Gravatar
	return "https://www.gravatar.com/avatar/00000000000000000000000000000000?s=150&d=mp&r=pg"
}

// ValidateAvatarFile validates an avatar file before processing
func (a *AvatarService) ValidateAvatarFile(file multipart.File, header *multipart.FileHeader) error {
	// Check file size
	if header.Size > a.maxFileSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size of %d bytes", header.Size, a.maxFileSize)
	}

	// Check content type
	contentType := header.Header.Get("Content-Type")
	if !a.isAllowedType(contentType) {
		return fmt.Errorf("file type %s not allowed. Allowed types: %v", contentType, a.allowedTypes)
	}

	// Try to decode image to ensure it's valid
	_, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("invalid image file: %w", err)
	}

	// Reset file pointer for subsequent reads
	if seeker, ok := file.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	return nil
}

// GetAvatarUploadConfig returns the current avatar upload configuration
func (a *AvatarService) GetAvatarUploadConfig() map[string]interface{} {
	return map[string]interface{}{
		"max_file_size":   a.maxFileSize,
		"allowed_types":   a.allowedTypes,
		"upload_path":     a.uploadPath,
		"default_size":    150,
		"thumbnail_sizes": []int{50, 100, 150, 300},
	}
}

// GetAvatarForDisplay returns the appropriate avatar URL for display with error handling
// This method ensures consistent avatar display across all UI components
func (a *AvatarService) GetAvatarForDisplay(user *models.User, size int) string {
	if user == nil {
		return a.GetDefaultAvatarURL()
	}

	// Try to get user's avatar URL
	avatarURL := a.GetUserAvatarURL(user, size)

	// If empty or error, return default
	if avatarURL == "" {
		return a.GetDefaultAvatarURL()
	}

	return avatarURL
}

// GetAvatarWithFallback returns avatar URL with automatic fallback to default on error
func (a *AvatarService) GetAvatarWithFallback(userID uint, size int) string {
	var user models.User
	if err := a.db.First(&user, userID).Error; err != nil {
		// User not found or error - return default
		return a.GetDefaultAvatarURL()
	}

	return a.GetAvatarForDisplay(&user, size)
}

// UpdateGravatarOnEmailChange automatically updates Gravatar when email changes
// This is called after email change confirmation
func (a *AvatarService) UpdateGravatarOnEmailChange(userID uint, oldEmail, newEmail string) error {
	// Get user
	var user models.User
	if err := a.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// If user has a custom avatar (not Gravatar), no update needed
	if user.AvatarURL != "" && !strings.Contains(user.AvatarURL, "gravatar.com") {
		return nil
	}

	// For Gravatar users, the URL is generated on-the-fly based on current email
	// No database update needed - GetUserAvatarURL will automatically use new email
	// This is intentional - Gravatar URLs are computed, not stored

	return nil
}

// GetAvatarDisplayHTML generates HTML for consistent avatar display
func (a *AvatarService) GetAvatarDisplayHTML(user *models.User, size int, cssClass string) string {
	avatarURL := a.GetAvatarForDisplay(user, size)
	username := "User"
	if user != nil {
		username = user.Username
	}

	if cssClass == "" {
		cssClass = "avatar"
	}

	return fmt.Sprintf(`<img src="%s" alt="%s's avatar" class="%s" width="%d" height="%d" onerror="this.src='%s'">`,
		avatarURL, username, cssClass, size, size, a.GetDefaultAvatarURL())
}

// ValidateAndGetAvatar validates user exists and returns their avatar URL
func (a *AvatarService) ValidateAndGetAvatar(userID uint, size int) (string, error) {
	var user models.User
	if err := a.db.First(&user, userID).Error; err != nil {
		return "", fmt.Errorf("user not found: %w", err)
	}

	return a.GetAvatarForDisplay(&user, size), nil
}

// GetMultipleAvatars efficiently retrieves avatars for multiple users
func (a *AvatarService) GetMultipleAvatars(userIDs []uint, size int) (map[uint]string, error) {
	if len(userIDs) == 0 {
		return make(map[uint]string), nil
	}

	var users []models.User
	if err := a.db.Where("id IN ?", userIDs).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	avatars := make(map[uint]string)
	for _, user := range users {
		avatars[user.ID] = a.GetAvatarForDisplay(&user, size)
	}

	// Fill in defaults for any missing users
	for _, userID := range userIDs {
		if _, exists := avatars[userID]; !exists {
			avatars[userID] = a.GetDefaultAvatarURL()
		}
	}

	return avatars, nil
}

// RefreshAvatarCache can be used to refresh cached avatar data
// This is useful after email changes or avatar uploads
func (a *AvatarService) RefreshAvatarCache(userID uint) error {
	// In a real implementation with caching, this would invalidate the cache
	// For now, it's a no-op since we don't have caching yet
	// But it provides a hook for future caching implementation
	return nil
}

// GetAvatarMetadata returns metadata about a user's avatar
func (a *AvatarService) GetAvatarMetadata(userID uint) (map[string]interface{}, error) {
	var user models.User
	if err := a.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	metadata := map[string]interface{}{
		"user_id":       user.ID,
		"has_custom":    user.AvatarURL != "" && !strings.Contains(user.AvatarURL, "gravatar.com"),
		"uses_gravatar": user.AvatarURL == "" || strings.Contains(user.AvatarURL, "gravatar.com"),
		"avatar_url":    a.GetAvatarForDisplay(&user, 150),
		"email":         user.Email,
	}

	if user.AvatarURL != "" && !strings.Contains(user.AvatarURL, "gravatar.com") {
		metadata["custom_avatar_url"] = user.AvatarURL
	}

	return metadata, nil
}
