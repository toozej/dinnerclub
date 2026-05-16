package services

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAvatarTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate the User model
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 128, 255})
		}
	}
	return img
}

func createMultipartFile(t *testing.T, img image.Image, filename, contentType string) (multipart.File, *multipart.FileHeader) {
	// Encode image to buffer
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Failed to encode test image: %v", err)
	}

	// Create multipart header
	header := &multipart.FileHeader{
		Filename: filename,
		Size:     int64(buf.Len()),
		Header:   make(textproto.MIMEHeader),
	}
	header.Header.Set("Content-Type", contentType)

	// Create a reader that implements multipart.File
	file := &mockMultipartFile{
		Reader: bytes.NewReader(buf.Bytes()),
		data:   buf.Bytes(),
	}

	return file, header
}

type mockMultipartFile struct {
	*bytes.Reader
	data []byte
}

func (m *mockMultipartFile) Close() error {
	return nil
}

func (m *mockMultipartFile) Seek(offset int64, whence int) (int64, error) {
	return m.Reader.Seek(offset, whence)
}

func TestNewAvatarService(t *testing.T) {
	db := setupAvatarTestDB(t)
	uploadPath := "/tmp/avatars"
	maxFileSize := int64(5 * 1024 * 1024) // 5MB
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}

	service := NewAvatarService(db, uploadPath, maxFileSize, allowedTypes)

	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}

	if service.db != db {
		t.Error("Database not set correctly")
	}

	if service.uploadPath != uploadPath {
		t.Errorf("Expected uploadPath %s, got %s", uploadPath, service.uploadPath)
	}

	if service.maxFileSize != maxFileSize {
		t.Errorf("Expected maxFileSize %d, got %d", maxFileSize, service.maxFileSize)
	}

	if len(service.allowedTypes) != len(allowedTypes) {
		t.Errorf("Expected %d allowed types, got %d", len(allowedTypes), len(service.allowedTypes))
	}
}

func TestGetGravatarURL(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/jpeg", "image/png"})

	tests := []struct {
		name     string
		email    string
		size     int
		expected string
	}{
		{
			name:     "Standard email",
			email:    "test@example.com",
			size:     150,
			expected: "https://www.gravatar.com/avatar/55502f40dc8b7c769880b10874abc9d0?s=150&d=mp&r=pg",
		},
		{
			name:     "Email with uppercase",
			email:    "Test@Example.COM",
			size:     150,
			expected: "https://www.gravatar.com/avatar/55502f40dc8b7c769880b10874abc9d0?s=150&d=mp&r=pg",
		},
		{
			name:     "Email with whitespace",
			email:    "  test@example.com  ",
			size:     150,
			expected: "https://www.gravatar.com/avatar/55502f40dc8b7c769880b10874abc9d0?s=150&d=mp&r=pg",
		},
		{
			name:     "Different size",
			email:    "test@example.com",
			size:     300,
			expected: "https://www.gravatar.com/avatar/55502f40dc8b7c769880b10874abc9d0?s=300&d=mp&r=pg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetGravatarURL(tt.email, tt.size)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestResizeImage(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/jpeg", "image/png"})

	tests := []struct {
		name         string
		sourceWidth  int
		sourceHeight int
		targetWidth  int
		targetHeight int
	}{
		{
			name:         "Downscale large image",
			sourceWidth:  1000,
			sourceHeight: 1000,
			targetWidth:  300,
			targetHeight: 300,
		},
		{
			name:         "Upscale small image",
			sourceWidth:  50,
			sourceHeight: 50,
			targetWidth:  150,
			targetHeight: 150,
		},
		{
			name:         "Non-square to square",
			sourceWidth:  800,
			sourceHeight: 600,
			targetWidth:  300,
			targetHeight: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcImg := createTestImage(tt.sourceWidth, tt.sourceHeight)
			resized := service.ResizeImage(srcImg, tt.targetWidth, tt.targetHeight)

			bounds := resized.Bounds()
			if bounds.Dx() != tt.targetWidth {
				t.Errorf("Expected width %d, got %d", tt.targetWidth, bounds.Dx())
			}
			if bounds.Dy() != tt.targetHeight {
				t.Errorf("Expected height %d, got %d", tt.targetHeight, bounds.Dy())
			}
		})
	}
}

func TestGetUserAvatarURL(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/jpeg", "image/png"})

	tests := []struct {
		name      string
		user      *models.User
		size      int
		expectURL string
	}{
		{
			name: "User with custom avatar",
			user: &models.User{
				Email:     "test@example.com",
				AvatarURL: "/uploads/avatars/avatar_1.png",
			},
			size:      150,
			expectURL: "/uploads/avatars/avatar_1.png",
		},
		{
			name: "User without custom avatar",
			user: &models.User{
				Email:     "test@example.com",
				AvatarURL: "",
			},
			size:      150,
			expectURL: "https://www.gravatar.com/avatar/55502f40dc8b7c769880b10874abc9d0?s=150&d=mp&r=pg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetUserAvatarURL(tt.user, tt.size)
			if result != tt.expectURL {
				t.Errorf("Expected %s, got %s", tt.expectURL, result)
			}
		})
	}
}

func TestValidateAvatarFile(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 1024*1024, []string{"image/png", "image/jpeg"})

	tests := []struct {
		name        string
		imageSize   int
		contentType string
		fileSize    int64
		expectError bool
	}{
		{
			name:        "Valid PNG file",
			imageSize:   200,
			contentType: "image/png",
			fileSize:    50000,
			expectError: false,
		},
		{
			name:        "File too large",
			imageSize:   200,
			contentType: "image/png",
			fileSize:    2 * 1024 * 1024,
			expectError: true,
		},
		{
			name:        "Invalid content type",
			imageSize:   200,
			contentType: "image/bmp",
			fileSize:    50000,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := createTestImage(tt.imageSize, tt.imageSize)
			file, header := createMultipartFile(t, img, "test.png", tt.contentType)
			header.Size = tt.fileSize

			err := service.ValidateAvatarFile(file, header)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestUploadCustomAvatar(t *testing.T) {
	db := setupAvatarTestDB(t)

	// Create temporary upload directory
	tmpDir := t.TempDir()

	service := NewAvatarService(db, tmpDir, 5*1024*1024, []string{"image/png", "image/jpeg"})

	// Create test user
	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create test image
	img := createTestImage(500, 500)
	file, header := createMultipartFile(t, img, "avatar.png", "image/png")

	// Upload avatar
	avatarURL, err := service.UploadCustomAvatar(user.ID, file, header)
	if err != nil {
		t.Fatalf("Failed to upload avatar: %v", err)
	}

	// Verify URL format
	if !strings.HasPrefix(avatarURL, "/uploads/avatars/") {
		t.Errorf("Expected URL to start with /uploads/avatars/, got %s", avatarURL)
	}

	// Verify file was created
	filename := filepath.Base(avatarURL)
	fullPath := filepath.Join(tmpDir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		t.Errorf("Avatar file was not created at %s", fullPath)
	}

	// Verify database was updated
	var updatedUser models.User
	if err := db.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve updated user: %v", err)
	}

	if updatedUser.AvatarURL != avatarURL {
		t.Errorf("Expected avatar URL %s, got %s", avatarURL, updatedUser.AvatarURL)
	}
}

func TestDeleteCustomAvatar(t *testing.T) {
	db := setupAvatarTestDB(t)

	// Create temporary upload directory
	tmpDir := t.TempDir()

	service := NewAvatarService(db, tmpDir, 5*1024*1024, []string{"image/png"})

	// Create test user with avatar
	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
		AvatarURL: "/uploads/avatars/avatar_1.png",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a dummy avatar file
	avatarPath := filepath.Join(tmpDir, "avatar_1.png")
	img := createTestImage(100, 100)
	file, err := os.Create(avatarPath)
	if err != nil {
		t.Fatalf("Failed to create test avatar file: %v", err)
	}
	png.Encode(file, img)
	file.Close()

	// Delete avatar
	err = service.DeleteCustomAvatar(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete avatar: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(avatarPath); !os.IsNotExist(err) {
		t.Error("Avatar file was not deleted")
	}

	// Verify database was updated
	var updatedUser models.User
	if err := db.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("Failed to retrieve updated user: %v", err)
	}

	if updatedUser.AvatarURL != "" {
		t.Errorf("Expected empty avatar URL, got %s", updatedUser.AvatarURL)
	}
}

func TestGetDefaultAvatarURL(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	defaultURL := service.GetDefaultAvatarURL()

	if !strings.Contains(defaultURL, "gravatar.com") {
		t.Error("Expected default URL to be a Gravatar URL")
	}

	if !strings.Contains(defaultURL, "d=mp") {
		t.Error("Expected default URL to use mystery person default")
	}
}

func TestGetAvatarUploadConfig(t *testing.T) {
	db := setupAvatarTestDB(t)
	maxFileSize := int64(5 * 1024 * 1024)
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
	uploadPath := "/tmp/avatars"

	service := NewAvatarService(db, uploadPath, maxFileSize, allowedTypes)

	config := service.GetAvatarUploadConfig()

	if config["max_file_size"] != maxFileSize {
		t.Errorf("Expected max_file_size %d, got %v", maxFileSize, config["max_file_size"])
	}

	if config["upload_path"] != uploadPath {
		t.Errorf("Expected upload_path %s, got %v", uploadPath, config["upload_path"])
	}

	configTypes, ok := config["allowed_types"].([]string)
	if !ok {
		t.Fatal("Expected allowed_types to be []string")
	}

	if len(configTypes) != len(allowedTypes) {
		t.Errorf("Expected %d allowed types, got %d", len(allowedTypes), len(configTypes))
	}
}

func TestUpdateAvatarOnEmailChange(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	// Create test user
	user := &models.User{
		Username:  "testuser",
		Email:     "old@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Update email
	newEmail := "new@example.com"
	err := service.UpdateAvatarOnEmailChange(user.ID, newEmail)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// For users without custom avatars, Gravatar is generated on-the-fly
	// so no database update is needed
}

func TestGetAvatarForDisplay(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	tests := []struct {
		name     string
		user     *models.User
		size     int
		contains string
	}{
		{
			name:     "Nil user returns default",
			user:     nil,
			size:     150,
			contains: "gravatar.com",
		},
		{
			name: "User with custom avatar",
			user: &models.User{
				Email:     "test@example.com",
				AvatarURL: "/uploads/avatars/custom.png",
			},
			size:     150,
			contains: "/uploads/avatars/custom.png",
		},
		{
			name: "User without avatar uses Gravatar",
			user: &models.User{
				Email:     "test@example.com",
				AvatarURL: "",
			},
			size:     150,
			contains: "gravatar.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetAvatarForDisplay(tt.user, tt.size)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %s, got %s", tt.contains, result)
			}
		})
	}
}

func TestGetAvatarWithFallback(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	// Test with non-existent user
	result := service.GetAvatarWithFallback(999, 150)
	if !strings.Contains(result, "gravatar.com") {
		t.Error("Expected default Gravatar for non-existent user")
	}

	// Create test user
	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
		AvatarURL: "/custom/avatar.png",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test with existing user
	result = service.GetAvatarWithFallback(user.ID, 150)
	if !strings.Contains(result, "/custom/avatar.png") {
		t.Errorf("Expected custom avatar, got %s", result)
	}
}

func TestUpdateGravatarOnEmailChange(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	// Create user with Gravatar
	user := &models.User{
		Username:  "testuser",
		Email:     "old@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
		AvatarURL: "",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Update Gravatar on email change
	err := service.UpdateGravatarOnEmailChange(user.ID, "old@example.com", "new@example.com")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test with custom avatar user
	user2 := &models.User{
		Username:  "testuser2",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
		AvatarURL: "/custom/avatar.png",
	}
	if err := db.Create(user2).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	err = service.UpdateGravatarOnEmailChange(user2.ID, "test@example.com", "new@example.com")
	if err != nil {
		t.Errorf("Expected no error for custom avatar user, got: %v", err)
	}
}

func TestGetAvatarDisplayHTML(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		AvatarURL: "/custom/avatar.png",
	}

	html := service.GetAvatarDisplayHTML(user, 150, "user-avatar")

	// Check HTML contains expected elements
	if !strings.Contains(html, "<img") {
		t.Error("Expected HTML to contain img tag")
	}
	if !strings.Contains(html, "/custom/avatar.png") {
		t.Error("Expected HTML to contain avatar URL")
	}
	if !strings.Contains(html, "user-avatar") {
		t.Error("Expected HTML to contain CSS class")
	}
	if !strings.Contains(html, "testuser") {
		t.Error("Expected HTML to contain username in alt text")
	}
	if !strings.Contains(html, "onerror") {
		t.Error("Expected HTML to contain onerror fallback")
	}
}

func TestValidateAndGetAvatar(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	// Test with non-existent user
	_, err := service.ValidateAndGetAvatar(999, 150)
	if err == nil {
		t.Error("Expected error for non-existent user")
	}

	// Create test user
	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
		AvatarURL: "/custom/avatar.png",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test with existing user
	avatarURL, err := service.ValidateAndGetAvatar(user.ID, 150)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !strings.Contains(avatarURL, "/custom/avatar.png") {
		t.Errorf("Expected custom avatar URL, got %s", avatarURL)
	}
}

func TestGetMultipleAvatars(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	// Create test users
	user1 := &models.User{
		Username:  "user1",
		Email:     "user1@example.com",
		Password:  "hashedpassword",
		Firstname: "User",
		Lastname:  "One",
		AvatarURL: "/avatar1.png",
	}
	user2 := &models.User{
		Username:  "user2",
		Email:     "user2@example.com",
		Password:  "hashedpassword",
		Firstname: "User",
		Lastname:  "Two",
		AvatarURL: "",
	}
	if err := db.Create(user1).Error; err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}
	if err := db.Create(user2).Error; err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Get multiple avatars
	userIDs := []uint{user1.ID, user2.ID, 999} // Include non-existent user
	avatars, err := service.GetMultipleAvatars(userIDs, 150)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify results
	if len(avatars) != 3 {
		t.Errorf("Expected 3 avatars, got %d", len(avatars))
	}

	if !strings.Contains(avatars[user1.ID], "/avatar1.png") {
		t.Error("Expected user1 to have custom avatar")
	}

	if !strings.Contains(avatars[user2.ID], "gravatar.com") {
		t.Error("Expected user2 to have Gravatar")
	}

	if !strings.Contains(avatars[999], "gravatar.com") {
		t.Error("Expected non-existent user to have default avatar")
	}
}

func TestGetAvatarMetadata(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	// Test with custom avatar
	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
		AvatarURL: "/custom/avatar.png",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	metadata, err := service.GetAvatarMetadata(user.ID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if metadata["user_id"] != user.ID {
		t.Error("Expected metadata to contain user ID")
	}

	hasCustom, ok := metadata["has_custom"].(bool)
	if !ok || !hasCustom {
		t.Error("Expected has_custom to be true")
	}

	usesGravatar, ok := metadata["uses_gravatar"].(bool)
	if !ok || usesGravatar {
		t.Error("Expected uses_gravatar to be false for custom avatar")
	}

	// Test with Gravatar user
	user2 := &models.User{
		Username:  "testuser2",
		Email:     "test2@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
		AvatarURL: "",
	}
	if err := db.Create(user2).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	metadata2, err := service.GetAvatarMetadata(user2.ID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	hasCustom2, ok := metadata2["has_custom"].(bool)
	if !ok || hasCustom2 {
		t.Error("Expected has_custom to be false for Gravatar user")
	}

	usesGravatar2, ok := metadata2["uses_gravatar"].(bool)
	if !ok || !usesGravatar2 {
		t.Error("Expected uses_gravatar to be true for Gravatar user")
	}
}

func TestRefreshAvatarCache(t *testing.T) {
	db := setupAvatarTestDB(t)
	service := NewAvatarService(db, "/tmp/avatars", 5*1024*1024, []string{"image/png"})

	// Create test user
	user := &models.User{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Firstname: "Test",
		Lastname:  "User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test refresh (currently a no-op but should not error)
	err := service.RefreshAvatarCache(user.ID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
