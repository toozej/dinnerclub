// Package services provides business logic services for the dinnerclub application
// including profile display, user statistics, and privacy management.
package services

import (
	"fmt"

	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/gorm"
)

// ProfileDisplayService handles enhanced user profile display functionality
type ProfileDisplayService struct {
	db            *gorm.DB
	socialService *SocialService
	avatarService *AvatarService
}

// NewProfileDisplayService creates a new ProfileDisplayService instance
func NewProfileDisplayService(db *gorm.DB, socialService *SocialService, avatarService *AvatarService) *ProfileDisplayService {
	return &ProfileDisplayService{
		db:            db,
		socialService: socialService,
		avatarService: avatarService,
	}
}

// EnhancedUserProfile represents a comprehensive user profile with all display information
type EnhancedUserProfile struct {
	User             models.User    `json:"user"`
	IsFollowing      bool           `json:"is_following"`
	FollowerCount    int            `json:"follower_count"`
	FollowingCount   int            `json:"following_count"`
	EntryCount       int            `json:"entry_count"`
	ReferralCount    int            `json:"referral_count"`
	RecentEntries    []models.Entry `json:"recent_entries"`
	AvatarURL        string         `json:"avatar_url"`
	CanEdit          bool           `json:"can_edit"`
	IsVisible        bool           `json:"is_visible"`
	ShowFollowButton bool           `json:"show_follow_button"`
	ShowContactInfo  bool           `json:"show_contact_info"`
}

// GetEnhancedUserProfile returns a comprehensive user profile with privacy controls
func (p *ProfileDisplayService) GetEnhancedUserProfile(userID, viewerID uint) (*EnhancedUserProfile, error) {
	// Get basic profile from social service
	basicProfile, err := p.socialService.GetUserProfile(userID, viewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get basic profile: %w", err)
	}

	// Get referral count
	referralCount, err := p.getReferralCount(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get referral count: %w", err)
	}

	// Get avatar URL
	avatarURL := p.avatarService.GetUserAvatarURL(&basicProfile.User, 150)

	// Determine privacy settings and visibility
	canEdit := (viewerID == userID)
	isVisible, showFollowButton, showContactInfo := p.determineVisibility(&basicProfile.User, viewerID, basicProfile.IsFollowing)

	return &EnhancedUserProfile{
		User:             basicProfile.User,
		IsFollowing:      basicProfile.IsFollowing,
		FollowerCount:    basicProfile.FollowerCount,
		FollowingCount:   basicProfile.FollowingCount,
		EntryCount:       basicProfile.EntryCount,
		ReferralCount:    referralCount,
		RecentEntries:    basicProfile.RecentEntries,
		AvatarURL:        avatarURL,
		CanEdit:          canEdit,
		IsVisible:        isVisible,
		ShowFollowButton: showFollowButton,
		ShowContactInfo:  showContactInfo,
	}, nil
}

// GetUserProfileByUsername returns a user profile by username
func (p *ProfileDisplayService) GetUserProfileByUsername(username string, viewerID uint) (*EnhancedUserProfile, error) {
	var user models.User
	if err := p.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return p.GetEnhancedUserProfile(user.ID, viewerID)
}

// GetAllUserEntries returns all entries for a user with pagination
func (p *ProfileDisplayService) GetAllUserEntries(userID uint, limit, offset int) ([]models.Entry, int64, error) {
	var user models.User
	if err := p.db.First(&user, userID).Error; err != nil {
		return nil, 0, fmt.Errorf("user not found: %w", err)
	}

	var entries []models.Entry
	var total int64

	// Get total count
	if err := p.db.Model(&models.Entry{}).Where("submitter = ?", user.Username).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count entries: %w", err)
	}

	// Get entries with pagination
	if err := p.db.Where("submitter = ?", user.Username).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get entries: %w", err)
	}

	return entries, total, nil
}

// GetUserEntriesByCategory returns entries for a user filtered by category
func (p *ProfileDisplayService) GetUserEntriesByCategory(userID uint, category, value string, limit, offset int) ([]models.Entry, int64, error) {
	var user models.User
	if err := p.db.First(&user, userID).Error; err != nil {
		return nil, 0, fmt.Errorf("user not found: %w", err)
	}

	var entries []models.Entry
	var total int64

	query := p.db.Model(&models.Entry{}).Where("submitter = ?", user.Username)

	// Apply category filter
	switch category {
	case "cuisine":
		query = query.Where("cuisine = ?", value)
	case "location":
		query = query.Where("location = ?", value)
	case "meal_service":
		query = query.Where("meal_service = ?", value)
	case "rating":
		query = query.Where("rating = ?", value)
	default:
		return nil, 0, fmt.Errorf("invalid category: %s", category)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count entries: %w", err)
	}

	// Get entries with pagination
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entries).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get entries: %w", err)
	}

	return entries, total, nil
}

// GetUserEntryCategories returns available categories and their counts for a user's entries
func (p *ProfileDisplayService) GetUserEntryCategories(userID uint) (map[string]map[string]int, error) {
	var user models.User
	if err := p.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	categories := make(map[string]map[string]int)

	// Get cuisine categories
	var cuisineResults []struct {
		Cuisine string
		Count   int
	}
	if err := p.db.Model(&models.Entry{}).
		Select("cuisine, COUNT(*) as count").
		Where("submitter = ? AND cuisine != ''", user.Username).
		Group("cuisine").
		Scan(&cuisineResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get cuisine categories: %w", err)
	}

	cuisines := make(map[string]int)
	for _, result := range cuisineResults {
		cuisines[result.Cuisine] = result.Count
	}
	categories["cuisine"] = cuisines

	// Get location categories
	var locationResults []struct {
		Location string
		Count    int
	}
	if err := p.db.Model(&models.Entry{}).
		Select("location, COUNT(*) as count").
		Where("submitter = ? AND location != ''", user.Username).
		Group("location").
		Scan(&locationResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get location categories: %w", err)
	}

	locations := make(map[string]int)
	for _, result := range locationResults {
		locations[result.Location] = result.Count
	}
	categories["location"] = locations

	// Get meal service categories
	var mealServiceResults []struct {
		MealService string
		Count       int
	}
	if err := p.db.Model(&models.Entry{}).
		Select("meal_service, COUNT(*) as count").
		Where("submitter = ? AND meal_service != ''", user.Username).
		Group("meal_service").
		Scan(&mealServiceResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get meal service categories: %w", err)
	}

	mealServices := make(map[string]int)
	for _, result := range mealServiceResults {
		mealServices[result.MealService] = result.Count
	}
	categories["meal_service"] = mealServices

	// Get rating categories
	var ratingResults []struct {
		Rating int
		Count  int
	}
	if err := p.db.Model(&models.Entry{}).
		Select("rating, COUNT(*) as count").
		Where("submitter = ? AND rating > 0", user.Username).
		Group("rating").
		Scan(&ratingResults).Error; err != nil {
		return nil, fmt.Errorf("failed to get rating categories: %w", err)
	}

	ratings := make(map[string]int)
	for _, result := range ratingResults {
		ratings[fmt.Sprintf("%d", result.Rating)] = result.Count
	}
	categories["rating"] = ratings

	return categories, nil
}

// UpdatePrivacySettings updates a user's privacy level
func (p *ProfileDisplayService) UpdatePrivacySettings(userID uint, privacyLevel int) error {
	if privacyLevel < 0 || privacyLevel > 2 {
		return fmt.Errorf("invalid privacy level: must be 0 (public), 1 (followers), or 2 (private)")
	}

	result := p.db.Model(&models.User{}).Where("id = ?", userID).Update("privacy_level", privacyLevel)
	if result.Error != nil {
		return fmt.Errorf("failed to update privacy settings: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetPrivacySettings returns the current privacy settings for a user
func (p *ProfileDisplayService) GetPrivacySettings(userID uint) (map[string]interface{}, error) {
	var user models.User
	if err := p.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	privacyInfo := map[string]interface{}{
		"privacy_level":       user.PrivacyLevel,
		"privacy_description": p.getPrivacyDescription(user.PrivacyLevel),
		"can_be_found":        user.PrivacyLevel == 0, // Only public profiles can be found in search
		"show_in_directory":   user.PrivacyLevel == 0, // Only public profiles show in user directory
	}

	return privacyInfo, nil
}

// CheckProfileAccess determines if a viewer can access a user's profile
func (p *ProfileDisplayService) CheckProfileAccess(userID, viewerID uint) (bool, string, error) {
	if userID == viewerID {
		return true, "own_profile", nil
	}

	var user models.User
	if err := p.db.First(&user, userID).Error; err != nil {
		return false, "user_not_found", fmt.Errorf("user not found: %w", err)
	}

	switch user.PrivacyLevel {
	case 0: // Public
		return true, "public", nil
	case 1: // Followers only
		if viewerID == 0 {
			return false, "login_required", nil
		}

		// Check if viewer is following the user
		var follow models.UserFollow
		err := p.db.Where("follower_id = ? AND following_id = ?", viewerID, userID).First(&follow).Error
		if err == nil {
			return true, "follower", nil
		}
		return false, "followers_only", nil
	case 2: // Private
		return false, "private", nil
	default:
		return false, "invalid_privacy", fmt.Errorf("invalid privacy level: %d", user.PrivacyLevel)
	}
}

// GetVisibleUsers returns users that are visible to the viewer based on privacy settings
func (p *ProfileDisplayService) GetVisibleUsers(viewerID uint, limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := p.db.Model(&models.User{})

	if viewerID == 0 {
		// Not logged in - only show public profiles
		query = query.Where("privacy_level = ?", 0)
	} else {
		// Logged in - show public profiles and profiles of users they follow
		query = query.Where(`
			privacy_level = 0 OR 
			id = ? OR 
			(privacy_level = 1 AND id IN (
				SELECT following_id FROM user_follows WHERE follower_id = ?
			))
		`, viewerID, viewerID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count visible users: %w", err)
	}

	// Get users with pagination
	if err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get visible users: %w", err)
	}

	return users, total, nil
}

// getPrivacyDescription returns a human-readable description of the privacy level
func (p *ProfileDisplayService) getPrivacyDescription(privacyLevel int) string {
	switch privacyLevel {
	case 0:
		return "Public - Anyone can see your profile and entries"
	case 1:
		return "Followers Only - Only people you follow back can see your full profile"
	case 2:
		return "Private - Your profile is completely private"
	default:
		return "Unknown privacy level"
	}
}

// getReferralCount gets the number of successful referrals for a user
func (p *ProfileDisplayService) getReferralCount(userID uint) (int, error) {
	var count int64
	if err := p.db.Model(&models.User{}).Where("referred_by = ?", userID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count referrals: %w", err)
	}
	return int(count), nil
}

// determineVisibility determines what information should be visible based on privacy settings
func (p *ProfileDisplayService) determineVisibility(user *models.User, viewerID uint, isFollowing bool) (isVisible, showFollowButton, showContactInfo bool) {
	// User viewing their own profile
	if viewerID == user.ID {
		return true, false, true
	}

	// Not logged in viewer
	if viewerID == 0 {
		switch user.PrivacyLevel {
		case 0: // Public
			return true, false, false
		case 1, 2: // Followers only or Private
			return false, false, false
		}
	}

	// Logged in viewer
	switch user.PrivacyLevel {
	case 0: // Public
		return true, true, false
	case 1: // Followers only
		if isFollowing {
			return true, true, true
		}
		return false, true, false
	case 2: // Private
		return false, false, false
	}

	return true, true, false
}

// GetProfileStatistics returns comprehensive statistics for a user profile
func (p *ProfileDisplayService) GetProfileStatistics(userID uint) (map[string]interface{}, error) {
	profile, err := p.GetEnhancedUserProfile(userID, userID) // User viewing their own stats
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Get additional statistics
	var totalLikes int64
	if err := p.db.Raw(`
		SELECT COUNT(*) FROM entry_likes el 
		JOIN entries e ON el.entry_id = e.id 
		WHERE e.submitter = ?
	`, profile.User.Username).Scan(&totalLikes).Error; err != nil {
		return nil, fmt.Errorf("failed to count likes: %w", err)
	}

	stats := map[string]interface{}{
		"entries":     profile.EntryCount,
		"followers":   profile.FollowerCount,
		"following":   profile.FollowingCount,
		"referrals":   profile.ReferralCount,
		"total_likes": int(totalLikes),
		"join_date":   profile.User.JoinDate,
		"is_founder":  profile.User.IsFounder,
	}

	return stats, nil
}
