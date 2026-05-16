package services

import (
	"fmt"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/toozej/dinnerclub/internal/models"
)

// **Validates: Requirements 7.1**
// Property 34: Gravatar Fallback
// For any user without a custom avatar, their Gravatar should be displayed based on their email address
func TestProperty34_GravatarFallback(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 34: Users without custom avatars use Gravatar", prop.ForAll(
		func(username string, emailPrefix string, size uint) bool {
			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png", "image/jpeg"})

			// Construct valid email
			email := emailPrefix + "@example.com"

			// Create user without custom avatar
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     email,
				AvatarURL: "", // No custom avatar
			}
			if err := db.Create(&user).Error; err != nil {
				return true // Skip on error
			}

			// Get avatar URL
			avatarURL := service.GetUserAvatarURL(&user, int(size))

			// Should be a Gravatar URL
			if !strings.Contains(avatarURL, "gravatar.com") {
				return false
			}

			// Should contain the size parameter
			expectedSize := fmt.Sprintf("s=%d", size)
			if !strings.Contains(avatarURL, expectedSize) {
				return false
			}

			// Should use mystery person default
			if !strings.Contains(avatarURL, "d=mp") {
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.UIntRange(50, 500),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Validates: Requirements 7.2**
// Property 35: Avatar Upload Processing
// For any custom avatar upload, the image should be resized and optimized according to system specifications
func TestProperty35_AvatarUploadProcessing(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 35: Uploaded avatars are resized to standard dimensions", prop.ForAll(
		func(width, height uint) bool {
			// Limit to reasonable sizes for test performance
			if width < 50 || width > 2000 || height < 50 || height > 2000 {
				return true
			}

			db := setupAvatarTestDB(t)
			tmpDir := t.TempDir()
			service := NewAvatarService(db, tmpDir, 10*1024*1024, []string{"image/png"})

			// Create test image with random dimensions
			img := createTestImage(int(width), int(height))

			// Resize to standard size (300x300)
			resized := service.ResizeImage(img, 300, 300)

			// Verify dimensions are exactly 300x300
			bounds := resized.Bounds()
			return bounds.Dx() == 300 && bounds.Dy() == 300
		},
		gen.UIntRange(50, 2000),
		gen.UIntRange(50, 2000),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Validates: Requirements 7.5**
// Property 38: Avatar Upload Validation
// For any avatar upload attempt, the system should validate file type and size according to configured limits
func TestProperty38_AvatarUploadValidation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 38: Files exceeding size limit are rejected", prop.ForAll(
		func(fileSize uint) bool {
			// Test with various file sizes
			if fileSize == 0 {
				return true
			}

			db := setupAvatarTestDB(t)
			maxSize := int64(1024 * 1024) // 1MB limit
			service := NewAvatarService(db, "/tmp/avatars", maxSize, []string{"image/png"})

			// Create test image
			img := createTestImage(100, 100)
			file, header := createMultipartFile(t, img, "test.png", "image/png")
			header.Size = int64(fileSize)

			// Validate file
			err := service.ValidateAvatarFile(file, header)

			// If file size exceeds limit, should return error
			if int64(fileSize) > maxSize {
				return err != nil
			}

			// If file size is within limit, should not return size error
			if err != nil && strings.Contains(err.Error(), "exceeds maximum") {
				return false
			}

			return true
		},
		gen.UIntRange(1, 5*1024*1024),
	))

	properties.Property("Property 38: Invalid content types are rejected", prop.ForAll(
		func(contentType string) bool {
			if len(contentType) == 0 {
				return true
			}

			db := setupAvatarTestDB(t)
			allowedTypes := []string{"image/png", "image/jpeg"}
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, allowedTypes)

			// Create test image
			img := createTestImage(100, 100)
			file, header := createMultipartFile(t, img, "test.png", contentType)
			header.Size = 50000 // Within size limit

			// Validate file
			err := service.ValidateAvatarFile(file, header)

			// Check if content type is in allowed list
			isAllowed := false
			for _, allowed := range allowedTypes {
				if contentType == allowed {
					isAllowed = true
					break
				}
			}

			// If not allowed, should return error
			if !isAllowed {
				return err != nil && strings.Contains(err.Error(), "not allowed")
			}

			return true
		},
		gen.OneConstOf("image/png", "image/jpeg", "image/gif", "image/bmp", "image/webp", "text/plain", "application/pdf"),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Validates: Requirements 7.3**
// Property 36: Avatar Display Consistency
// For any user avatar, it should be displayed consistently across all interface locations
func TestProperty36_AvatarDisplayConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 36: Avatar URL is consistent for same user", prop.ForAll(
		func(username string, emailPrefix string, avatarURL string, size uint) bool {
			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			email := emailPrefix + "@example.com"

			// Create user
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     email,
				AvatarURL: avatarURL,
			}
			if err := db.Create(&user).Error; err != nil {
				return true
			}

			// Get avatar URL multiple times
			url1 := service.GetUserAvatarURL(&user, int(size))
			url2 := service.GetUserAvatarURL(&user, int(size))
			url3 := service.GetUserAvatarURL(&user, int(size))

			// All calls should return the same URL
			return url1 == url2 && url2 == url3
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.AlphaString(),
		gen.UIntRange(50, 500),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Validates: Requirements 7.4**
// Property 37: Gravatar Email Update
// For any user email change, their Gravatar should be automatically updated to reflect the new email address
func TestProperty37_GravatarEmailUpdate(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 37: Gravatar changes when email changes", prop.ForAll(
		func(username string, oldEmailPrefix string, newEmailPrefix string) bool {
			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			// Construct emails
			oldEmail := oldEmailPrefix + "@example.com"
			newEmail := newEmailPrefix + "@test.com"

			// Create user with old email (no custom avatar)
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     oldEmail,
				AvatarURL: "", // No custom avatar, will use Gravatar
			}
			if err := db.Create(&user).Error; err != nil {
				return true
			}

			// Get Gravatar URL with old email
			oldGravatarURL := service.GetUserAvatarURL(&user, 150)

			// Update email
			user.Email = newEmail
			if err := db.Save(&user).Error; err != nil {
				return true
			}

			// Get Gravatar URL with new email
			newGravatarURL := service.GetUserAvatarURL(&user, 150)

			// Gravatar URLs should be different (different email hashes)
			return oldGravatarURL != newGravatarURL &&
				strings.Contains(oldGravatarURL, "gravatar.com") &&
				strings.Contains(newGravatarURL, "gravatar.com")
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.Property("Custom avatars are not affected by email changes", prop.ForAll(
		func(username string, oldEmailPrefix string, newEmailPrefix string, customAvatarURL string) bool {
			// Skip empty custom avatar URLs
			if len(customAvatarURL) == 0 {
				return true
			}

			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			oldEmail := oldEmailPrefix + "@example.com"
			newEmail := newEmailPrefix + "@test.com"

			// Create user with custom avatar
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     oldEmail,
				AvatarURL: customAvatarURL,
			}
			if err := db.Create(&user).Error; err != nil {
				return true
			}

			// Get avatar URL with old email
			oldAvatarURL := service.GetUserAvatarURL(&user, 150)

			// Update email
			user.Email = newEmail
			if err := db.Save(&user).Error; err != nil {
				return true
			}

			// Get avatar URL with new email
			newAvatarURL := service.GetUserAvatarURL(&user, 150)

			// Custom avatar URL should remain the same
			return oldAvatarURL == newAvatarURL && oldAvatarURL == customAvatarURL
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// **Validates: Requirements 7.6**
// Property 39: Avatar Loading Error Handling
// For any avatar that fails to load, a default placeholder image should be displayed instead
func TestProperty39_AvatarLoadingErrorHandling(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Property 39: Default avatar is always available", prop.ForAll(
		func(seed uint) bool {
			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			// Get default avatar URL
			defaultURL := service.GetDefaultAvatarURL()

			// Should be a valid Gravatar URL
			if !strings.Contains(defaultURL, "gravatar.com") {
				return false
			}

			// Should use mystery person default
			if !strings.Contains(defaultURL, "d=mp") {
				return false
			}

			// Should be non-empty
			return len(defaultURL) > 0
		},
		gen.UInt(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional property test: Gravatar URL generation is deterministic
func TestProperty_GravatarDeterminism(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Gravatar URL generation is deterministic for same email", prop.ForAll(
		func(emailPrefix string, size uint) bool {
			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			email := emailPrefix + "@example.com"

			// Generate Gravatar URL multiple times
			url1 := service.GetGravatarURL(email, int(size))
			url2 := service.GetGravatarURL(email, int(size))
			url3 := service.GetGravatarURL(email, int(size))

			// All should be identical
			return url1 == url2 && url2 == url3
		},
		gen.Identifier(),
		gen.UIntRange(50, 500),
	))

	properties.Property("Gravatar URL is case-insensitive for email", prop.ForAll(
		func(emailPrefix string, size uint) bool {
			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			email := emailPrefix + "@example.com"

			// Generate URLs with different cases
			lowerURL := service.GetGravatarURL(strings.ToLower(email), int(size))
			upperURL := service.GetGravatarURL(strings.ToUpper(email), int(size))
			mixedURL := service.GetGravatarURL(email, int(size))

			// All should generate the same hash (case-insensitive)
			return lowerURL == upperURL && upperURL == mixedURL
		},
		gen.Identifier(),
		gen.UIntRange(50, 500),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional property test: Image resizing maintains aspect ratio handling
func TestProperty_ImageResizingQuality(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Resized images have correct dimensions", prop.ForAll(
		func(srcWidth, srcHeight, dstWidth, dstHeight uint) bool {
			// Limit to reasonable sizes
			if srcWidth < 10 || srcWidth > 1000 || srcHeight < 10 || srcHeight > 1000 {
				return true
			}
			if dstWidth < 10 || dstWidth > 500 || dstHeight < 10 || dstHeight > 500 {
				return true
			}

			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			// Create source image
			srcImg := createTestImage(int(srcWidth), int(srcHeight))

			// Resize
			resized := service.ResizeImage(srcImg, int(dstWidth), int(dstHeight))

			// Verify dimensions
			bounds := resized.Bounds()
			return bounds.Dx() == int(dstWidth) && bounds.Dy() == int(dstHeight)
		},
		gen.UIntRange(10, 1000),
		gen.UIntRange(10, 1000),
		gen.UIntRange(10, 500),
		gen.UIntRange(10, 500),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional property test: Avatar configuration is consistent
func TestProperty_AvatarConfigurationConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Avatar configuration reflects initialization parameters", prop.ForAll(
		func(maxSize uint, numTypes uint) bool {
			// Limit to reasonable values
			if maxSize == 0 || maxSize > 100*1024*1024 || numTypes == 0 || numTypes > 10 {
				return true
			}

			db := setupAvatarTestDB(t)

			// Create allowed types list
			allowedTypes := make([]string, numTypes)
			for i := uint(0); i < numTypes; i++ {
				allowedTypes[i] = fmt.Sprintf("image/type%d", i)
			}

			service := NewAvatarService(db, "/tmp/avatars", int64(maxSize), allowedTypes)

			// Get configuration
			config := service.GetAvatarUploadConfig()

			// Verify configuration matches initialization
			configMaxSize, ok := config["max_file_size"].(int64)
			if !ok || configMaxSize != int64(maxSize) {
				return false
			}

			configTypes, ok := config["allowed_types"].([]string)
			if !ok || len(configTypes) != int(numTypes) {
				return false
			}

			return true
		},
		gen.UIntRange(1024, 10*1024*1024),
		gen.UIntRange(1, 5),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional property test: Avatar deletion is idempotent
func TestProperty_AvatarDeletionIdempotence(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Deleting avatar multiple times is safe", prop.ForAll(
		func(username string, iterations uint) bool {
			if iterations == 0 || iterations > 5 {
				return true
			}

			db := setupAvatarTestDB(t)
			tmpDir := t.TempDir()
			service := NewAvatarService(db, tmpDir, 5*1024*1024, []string{"image/png"})

			// Create user
			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     username + "@example.com",
				AvatarURL: "",
			}
			if err := db.Create(&user).Error; err != nil {
				return true
			}

			// Delete avatar multiple times
			for i := uint(0); i < iterations; i++ {
				if err := service.DeleteCustomAvatar(user.ID); err != nil {
					return false
				}
			}

			// Verify user has no avatar URL
			var updatedUser models.User
			if err := db.First(&updatedUser, user.ID).Error; err != nil {
				return false
			}

			return updatedUser.AvatarURL == ""
		},
		gen.Identifier(),
		gen.UIntRange(1, 5),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Additional property tests for avatar display and management features

// Property test: GetAvatarForDisplay always returns a valid URL
func TestProperty_AvatarDisplayAlwaysValid(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("GetAvatarForDisplay always returns non-empty URL", prop.ForAll(
		func(username string, emailPrefix string, hasCustomAvatar bool) bool {
			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			email := emailPrefix + "@example.com"
			avatarURL := ""
			if hasCustomAvatar {
				avatarURL = "/custom/avatar.png"
			}

			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     email,
				AvatarURL: avatarURL,
			}
			if err := db.Create(&user).Error; err != nil {
				return true
			}

			// Get avatar for display
			displayURL := service.GetAvatarForDisplay(&user, 150)

			// Should always return a non-empty URL
			return len(displayURL) > 0
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Bool(),
	))

	properties.Property("GetAvatarForDisplay with nil user returns default", prop.ForAll(
		func(size uint) bool {
			if size == 0 || size > 1000 {
				return true
			}

			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			// Get avatar for nil user
			displayURL := service.GetAvatarForDisplay(nil, int(size))

			// Should return default Gravatar
			return strings.Contains(displayURL, "gravatar.com") && len(displayURL) > 0
		},
		gen.UIntRange(50, 500),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property test: GetAvatarWithFallback never fails
func TestProperty_AvatarFallbackNeverFails(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("GetAvatarWithFallback always returns a URL", prop.ForAll(
		func(userID uint, size uint) bool {
			if size == 0 || size > 1000 {
				return true
			}

			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			// Get avatar with fallback (even for non-existent user)
			avatarURL := service.GetAvatarWithFallback(userID, int(size))

			// Should always return a non-empty URL
			return len(avatarURL) > 0 && strings.Contains(avatarURL, "gravatar.com")
		},
		gen.UIntRange(1, 1000),
		gen.UIntRange(50, 500),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property test: GetMultipleAvatars returns correct count
func TestProperty_MultipleAvatarsCorrectCount(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("GetMultipleAvatars returns avatar for each requested user", prop.ForAll(
		func(numUsers uint) bool {
			// Limit to reasonable numbers
			if numUsers == 0 || numUsers > 20 {
				return true
			}

			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			// Create users
			userIDs := make([]uint, numUsers)
			for i := uint(0); i < numUsers; i++ {
				user := models.User{
					Username: fmt.Sprintf("user%d", i),
					Password: "password123",
					Email:    fmt.Sprintf("user%d@example.com", i),
				}
				if err := db.Create(&user).Error; err != nil {
					return true
				}
				userIDs[i] = user.ID
			}

			// Get multiple avatars
			avatars, err := service.GetMultipleAvatars(userIDs, 150)
			if err != nil {
				return false
			}

			// Should return exactly the requested number of avatars
			if len(avatars) != int(numUsers) {
				return false
			}

			// All avatars should be non-empty
			for _, avatarURL := range avatars {
				if len(avatarURL) == 0 {
					return false
				}
			}

			return true
		},
		gen.UIntRange(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property test: Avatar metadata is consistent
func TestProperty_AvatarMetadataConsistency(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Avatar metadata reflects actual user state", prop.ForAll(
		func(username string, emailPrefix string, hasCustomAvatar bool) bool {
			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			email := emailPrefix + "@example.com"
			avatarURL := ""
			if hasCustomAvatar {
				avatarURL = "/custom/avatar.png"
			}

			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     email,
				AvatarURL: avatarURL,
			}
			if err := db.Create(&user).Error; err != nil {
				return true
			}

			// Get metadata
			metadata, err := service.GetAvatarMetadata(user.ID)
			if err != nil {
				return false
			}

			// Verify metadata consistency
			hasCustomMeta, ok := metadata["has_custom"].(bool)
			if !ok {
				return false
			}

			usesGravatarMeta, ok := metadata["uses_gravatar"].(bool)
			if !ok {
				return false
			}

			// has_custom and uses_gravatar should be opposites
			if hasCustomAvatar {
				return hasCustomMeta && !usesGravatarMeta
			}
			return !hasCustomMeta && usesGravatarMeta
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Bool(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property test: Avatar display HTML is well-formed
func TestProperty_AvatarDisplayHTMLWellFormed(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("Avatar display HTML contains required elements", prop.ForAll(
		func(username string, emailPrefix string, size uint, cssClass string) bool {
			if size == 0 || size > 1000 {
				return true
			}

			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			email := emailPrefix + "@example.com"
			user := models.User{
				Username: username,
				Password: "password123",
				Email:    email,
			}

			// Generate HTML
			html := service.GetAvatarDisplayHTML(&user, int(size), cssClass)

			// Verify HTML structure
			if !strings.Contains(html, "<img") {
				return false
			}

			if !strings.Contains(html, "src=") {
				return false
			}

			if !strings.Contains(html, "alt=") {
				return false
			}

			if !strings.Contains(html, "onerror=") {
				return false
			}

			// Should contain size attributes
			sizeStr := fmt.Sprintf("width=\"%d\"", size)
			if !strings.Contains(html, sizeStr) {
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.UIntRange(50, 500),
		gen.AlphaString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property test: UpdateGravatarOnEmailChange is safe
func TestProperty_GravatarUpdateSafety(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("UpdateGravatarOnEmailChange never fails for valid users", prop.ForAll(
		func(username string, oldEmailPrefix string, newEmailPrefix string) bool {
			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			oldEmail := oldEmailPrefix + "@example.com"
			newEmail := newEmailPrefix + "@test.com"

			user := models.User{
				Username: username,
				Password: "password123",
				Email:    oldEmail,
			}
			if err := db.Create(&user).Error; err != nil {
				return true
			}

			// Update should not fail
			err := service.UpdateGravatarOnEmailChange(user.ID, oldEmail, newEmail)
			return err == nil
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
	))

	properties.Property("UpdateGravatarOnEmailChange preserves custom avatars", prop.ForAll(
		func(username string, oldEmailPrefix string, newEmailPrefix string, customAvatar string) bool {
			if len(customAvatar) == 0 {
				return true
			}

			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			oldEmail := oldEmailPrefix + "@example.com"
			newEmail := newEmailPrefix + "@test.com"

			user := models.User{
				Username:  username,
				Password:  "password123",
				Email:     oldEmail,
				AvatarURL: customAvatar,
			}
			if err := db.Create(&user).Error; err != nil {
				return true
			}

			// Update Gravatar
			err := service.UpdateGravatarOnEmailChange(user.ID, oldEmail, newEmail)
			if err != nil {
				return false
			}

			// Verify custom avatar is preserved
			var updatedUser models.User
			if err := db.First(&updatedUser, user.ID).Error; err != nil {
				return false
			}

			return updatedUser.AvatarURL == customAvatar
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property test: ValidateAndGetAvatar error handling
func TestProperty_ValidateAndGetAvatarErrorHandling(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("ValidateAndGetAvatar returns error for non-existent users", prop.ForAll(
		func(userID uint, size uint) bool {
			if size == 0 || size > 1000 {
				return true
			}

			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			// Try to get avatar for non-existent user
			_, err := service.ValidateAndGetAvatar(userID, int(size))

			// Should return an error
			return err != nil
		},
		gen.UIntRange(1000, 10000), // Use high IDs that won't exist
		gen.UIntRange(50, 500),
	))

	properties.Property("ValidateAndGetAvatar succeeds for existing users", prop.ForAll(
		func(username string, emailPrefix string, size uint) bool {
			if size == 0 || size > 1000 {
				return true
			}

			db := setupAvatarTestDB(t)
			service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

			email := emailPrefix + "@example.com"
			user := models.User{
				Username: username,
				Password: "password123",
				Email:    email,
			}
			if err := db.Create(&user).Error; err != nil {
				return true
			}

			// Should succeed for existing user
			avatarURL, err := service.ValidateAndGetAvatar(user.ID, int(size))
			if err != nil {
				return false
			}

			// Should return non-empty URL
			return len(avatarURL) > 0
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.UIntRange(50, 500),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
