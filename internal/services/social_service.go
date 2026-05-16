// Package services provides business logic services for the dinnerclub application
package services

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// SocialService handles social features including following, feeds, and notifications
type SocialService struct {
	db *gorm.DB
}

// NewSocialService creates a new social service instance
func NewSocialService(db *gorm.DB) *SocialService {
	return &SocialService{
		db: db,
	}
}

// UserProfile represents a user profile with social statistics
type UserProfile struct {
	User           models.User    `json:"user"`
	IsFollowing    bool           `json:"is_following"`
	FollowerCount  int            `json:"follower_count"`
	FollowingCount int            `json:"following_count"`
	EntryCount     int            `json:"entry_count"`
	RecentEntries  []models.Entry `json:"recent_entries"`
}

// FeedEntry represents an entry in a social feed with author information
type FeedEntry struct {
	Entry           models.Entry `json:"entry"`
	AuthorFirstname string       `json:"author_firstname"`
	AuthorLastname  string       `json:"author_lastname"`
	AuthorAvatarURL string       `json:"author_avatar_url"`
	AuthorIsFounder bool         `json:"author_is_founder"`
}

// FollowUser creates a following relationship between two users
func (s *SocialService) FollowUser(followerID, followingID uint) error {
	// Prevent users from following themselves
	if followerID == followingID {
		return fmt.Errorf("users cannot follow themselves")
	}

	// Check if both users exist
	var follower, following models.User
	if err := s.db.First(&follower, followerID).Error; err != nil {
		return fmt.Errorf("follower user not found: %w", err)
	}
	if err := s.db.First(&following, followingID).Error; err != nil {
		return fmt.Errorf("following user not found: %w", err)
	}

	// Check if relationship already exists
	var existingFollow models.UserFollow
	err := s.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&existingFollow).Error
	if err == nil {
		return fmt.Errorf("user is already following this user")
	}
	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing follow relationship: %w", err)
	}

	// Create the follow relationship
	follow := models.UserFollow{
		FollowerID:  followerID,
		FollowingID: followingID,
	}

	if err := s.db.Create(&follow).Error; err != nil {
		return fmt.Errorf("failed to create follow relationship: %w", err)
	}

	// Send notification to the followed user
	if err := s.sendFollowNotification(follower, following); err != nil {
		// Log error but don't fail the follow operation
		fmt.Printf("Warning: Failed to send follow notification: %v\n", err)
	}

	return nil
}

// UnfollowUser removes a following relationship between two users
func (s *SocialService) UnfollowUser(followerID, followingID uint) error {
	// Find and delete the follow relationship
	result := s.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&models.UserFollow{})
	if result.Error != nil {
		return fmt.Errorf("failed to unfollow user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("follow relationship not found")
	}

	return nil
}

// GetFollowers returns a list of users who follow the given user
func (s *SocialService) GetFollowers(userID uint) ([]models.User, error) {
	var followers []models.User

	if err := s.db.Table("users").
		Joins("JOIN user_follows ON users.id = user_follows.follower_id").
		Where("user_follows.following_id = ?", userID).
		Find(&followers).Error; err != nil {
		return nil, fmt.Errorf("failed to get followers: %w", err)
	}

	return followers, nil
}

// GetFollowing returns a list of users that the given user follows
func (s *SocialService) GetFollowing(userID uint) ([]models.User, error) {
	var following []models.User

	if err := s.db.Table("users").
		Joins("JOIN user_follows ON users.id = user_follows.following_id").
		Where("user_follows.follower_id = ?", userID).
		Find(&following).Error; err != nil {
		return nil, fmt.Errorf("failed to get following: %w", err)
	}

	return following, nil
}

// GetUserProfile returns a comprehensive user profile with social statistics
func (s *SocialService) GetUserProfile(userID, viewerID uint) (*UserProfile, error) {
	// Get the user
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if viewer is following this user
	var isFollowing bool
	if viewerID != 0 && viewerID != userID {
		var follow models.UserFollow
		err := s.db.Where("follower_id = ? AND following_id = ?", viewerID, userID).First(&follow).Error
		isFollowing = (err == nil)
	}

	// Get follower count
	var followerCount int64
	if err := s.db.Model(&models.UserFollow{}).Where("following_id = ?", userID).Count(&followerCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count followers: %w", err)
	}

	// Get following count
	var followingCount int64
	if err := s.db.Model(&models.UserFollow{}).Where("follower_id = ?", userID).Count(&followingCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count following: %w", err)
	}

	// Get entry count
	var entryCount int64
	if err := s.db.Model(&models.Entry{}).Where("submitter = ?", user.Username).Count(&entryCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count entries: %w", err)
	}

	// Get recent entries (limit to 5)
	var recentEntries []models.Entry
	if err := s.db.Where("submitter = ?", user.Username).
		Order("created_at DESC").
		Limit(5).
		Find(&recentEntries).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent entries: %w", err)
	}

	return &UserProfile{
		User:           user,
		IsFollowing:    isFollowing,
		FollowerCount:  int(followerCount),
		FollowingCount: int(followingCount),
		EntryCount:     int(entryCount),
		RecentEntries:  recentEntries,
	}, nil
}

// GetFollowingFeed returns entries from users that the given user follows
func (s *SocialService) GetFollowingFeed(userID uint, limit, offset int) ([]models.Entry, error) {
	var entries []models.Entry

	// Get entries from followed users
	query := s.db.Table("entries").
		Joins("JOIN users ON entries.submitter = users.username").
		Joins("JOIN user_follows ON users.id = user_follows.following_id").
		Where("user_follows.follower_id = ?", userID).
		Order("entries.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("failed to get following feed: %w", err)
	}

	return entries, nil
}

// sendFollowNotification sends a notification when someone follows a user
func (s *SocialService) sendFollowNotification(follower, following models.User) error {
	notification := models.SocialNotification{
		UserID:    following.ID,
		ActorID:   follower.ID,
		Type:      "follow",
		Message:   fmt.Sprintf("%s %s started following you", follower.Firstname, follower.Lastname),
		Read:      false,
		RelatedID: &follower.ID,
	}

	if err := s.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create follow notification: %w", err)
	}

	return nil
}

// GetSocialNotifications returns social notifications for a user
func (s *SocialService) GetSocialNotifications(userID uint, limit int) ([]models.SocialNotification, error) {
	var notifications []models.SocialNotification
	query := s.db.Where("user_id = ?", userID).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get social notifications: %w", err)
	}
	return notifications, nil
}

// MarkNotificationAsRead marks a notification as read
func (s *SocialService) MarkNotificationAsRead(notificationID uint) error {
	if err := s.db.Model(&models.SocialNotification{}).
		Where("id = ?", notificationID).
		Update("read", true).Error; err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

// IsFollowing checks if one user follows another
func (s *SocialService) IsFollowing(followerID, followingID uint) (bool, error) {
	var follow models.UserFollow
	err := s.db.Where("follower_id = ? AND following_id = ?", followerID, followingID).First(&follow).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check follow relationship: %w", err)
	}
	return true, nil
}

// NotifyFollowersOfNewEntry sends notifications to followers when a user creates a new entry
func (s *SocialService) NotifyFollowersOfNewEntry(userID uint, entryID uint) error {
	// Get the user who created the entry
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return fmt.Errorf("failed to find user: %w", err)
	}

	// Get the entry details
	var entry models.Entry
	if err := s.db.First(&entry, entryID).Error; err != nil {
		return fmt.Errorf("failed to find entry: %w", err)
	}

	// Get all followers of this user
	followers, err := s.GetFollowers(userID)
	if err != nil {
		return fmt.Errorf("failed to get followers: %w", err)
	}

	// Create notifications for each follower
	for _, follower := range followers {
		notification := models.SocialNotification{
			UserID:    follower.ID,
			ActorID:   userID,
			Type:      "entry",
			Message:   fmt.Sprintf("%s %s added a new entry: %s", user.Firstname, user.Lastname, entry.Name),
			Read:      false,
			RelatedID: &entryID,
		}

		if err := s.db.Create(&notification).Error; err != nil {
			// Log error but continue with other notifications
			fmt.Printf("Warning: Failed to create entry notification for user %d: %v\n", follower.ID, err)
		}
	}

	return nil
}

// GetFollowingFeedWithDetails returns entries from followed users with additional metadata
func (s *SocialService) GetFollowingFeedWithDetails(userID uint, limit, offset int) ([]FeedEntry, error) {
	var feedEntries []FeedEntry

	// Get entries from followed users with user information
	query := s.db.Table("entries").
		Select("entries.*, users.firstname, users.lastname, users.avatar_url, users.is_founder").
		Joins("JOIN users ON entries.submitter = users.username").
		Joins("JOIN user_follows ON users.id = user_follows.following_id").
		Where("user_follows.follower_id = ?", userID).
		Order("entries.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get following feed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var feedEntry FeedEntry
		var entry models.Entry
		var firstname, lastname, avatarURL string
		var isFounder bool

		err := rows.Scan(
			&entry.ID, &entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt,
			&entry.Submitter, &entry.Name, &entry.Location, &entry.Cuisine, &entry.Visited,
			&entry.LookingForwardTo, &entry.Closed, &entry.MealService, &entry.Ordered,
			&entry.OrderNext, &entry.DoNotOrder, &entry.FoodRating, &entry.AmbienceRating,
			&entry.ValueRating, &entry.Comments,
			&firstname, &lastname, &avatarURL, &isFounder,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feed entry: %w", err)
		}

		feedEntry.Entry = entry
		feedEntry.AuthorFirstname = firstname
		feedEntry.AuthorLastname = lastname
		feedEntry.AuthorAvatarURL = avatarURL
		feedEntry.AuthorIsFounder = isFounder

		feedEntries = append(feedEntries, feedEntry)
	}

	return feedEntries, nil
}

// GetUnreadNotificationCount returns the count of unread notifications for a user
func (s *SocialService) GetUnreadNotificationCount(userID uint) (int, error) {
	var count int64
	if err := s.db.Model(&models.SocialNotification{}).
		Where("user_id = ? AND read = ?", userID, false).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unread notifications: %w", err)
	}
	return int(count), nil
}

// MarkAllNotificationsAsRead marks all notifications for a user as read
func (s *SocialService) MarkAllNotificationsAsRead(userID uint) error {
	if err := s.db.Model(&models.SocialNotification{}).
		Where("user_id = ?", userID).
		Update("read", true).Error; err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

// GetNotificationsByType returns notifications of a specific type for a user
func (s *SocialService) GetNotificationsByType(userID uint, notificationType string, limit int) ([]models.SocialNotification, error) {
	var notifications []models.SocialNotification
	query := s.db.Where("user_id = ? AND type = ?", userID, notificationType).
		Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get notifications by type: %w", err)
	}
	return notifications, nil
}
