// Package services provides business logic services for the dinnerclub application
package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// setupSocialUnitTestDB creates an in-memory SQLite database for unit testing
func setupSocialUnitTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(
		&models.User{},
		&models.UserFollow{},
		&models.Entry{},
		&models.SocialNotification{},
	)
	require.NoError(t, err)

	return db
}

func TestSocialService_FollowUser(t *testing.T) {
	db := setupSocialUnitTestDB(t)
	service := NewSocialService(db)

	// Create test users
	user1 := models.User{
		Username:  "user1",
		Password:  "password123456",
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john@example.com",
	}
	user2 := models.User{
		Username:  "user2",
		Password:  "password123456",
		Firstname: "Jane",
		Lastname:  "Smith",
		Email:     "jane@example.com",
	}

	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)

	// Test successful follow
	err := service.FollowUser(user1.ID, user2.ID)
	assert.NoError(t, err)

	// Verify relationship exists
	isFollowing, err := service.IsFollowing(user1.ID, user2.ID)
	require.NoError(t, err)
	assert.True(t, isFollowing)

	// Test duplicate follow
	err = service.FollowUser(user1.ID, user2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already following")

	// Test self-follow prevention
	err = service.FollowUser(user1.ID, user1.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot follow themselves")
}

func TestSocialService_UnfollowUser(t *testing.T) {
	db := setupSocialUnitTestDB(t)
	service := NewSocialService(db)

	// Create test users
	user1 := models.User{
		Username:  "user1",
		Password:  "password123456",
		Firstname: "John",
		Lastname:  "Doe",
		Email:     "john@example.com",
	}
	user2 := models.User{
		Username:  "user2",
		Password:  "password123456",
		Firstname: "Jane",
		Lastname:  "Smith",
		Email:     "jane@example.com",
	}

	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)

	// First follow the user
	err := service.FollowUser(user1.ID, user2.ID)
	require.NoError(t, err)

	// Test successful unfollow
	err = service.UnfollowUser(user1.ID, user2.ID)
	assert.NoError(t, err)

	// Verify relationship no longer exists
	isFollowing, err := service.IsFollowing(user1.ID, user2.ID)
	require.NoError(t, err)
	assert.False(t, isFollowing)

	// Test unfollow non-existent relationship
	err = service.UnfollowUser(user1.ID, user2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSocialService_GetFollowersAndFollowing(t *testing.T) {
	db := setupSocialUnitTestDB(t)
	service := NewSocialService(db)

	// Create test users
	user1 := models.User{Username: "user1", Password: "password123456", Firstname: "John", Lastname: "Doe", Email: "john@example.com"}
	user2 := models.User{Username: "user2", Password: "password123456", Firstname: "Jane", Lastname: "Smith", Email: "jane@example.com"}
	user3 := models.User{Username: "user3", Password: "password123456", Firstname: "Bob", Lastname: "Johnson", Email: "bob@example.com"}

	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)
	require.NoError(t, db.Create(&user3).Error)

	// Create follow relationships: user1 follows user2, user3 follows user2
	require.NoError(t, service.FollowUser(user1.ID, user2.ID))
	require.NoError(t, service.FollowUser(user3.ID, user2.ID))

	// Test GetFollowers
	followers, err := service.GetFollowers(user2.ID)
	require.NoError(t, err)
	assert.Len(t, followers, 2)

	followerIDs := make(map[uint]bool)
	for _, follower := range followers {
		followerIDs[follower.ID] = true
	}
	assert.True(t, followerIDs[user1.ID])
	assert.True(t, followerIDs[user3.ID])

	// Test GetFollowing
	following, err := service.GetFollowing(user1.ID)
	require.NoError(t, err)
	assert.Len(t, following, 1)
	assert.Equal(t, user2.ID, following[0].ID)
}

func TestSocialService_GetUserProfile(t *testing.T) {
	db := setupSocialUnitTestDB(t)
	service := NewSocialService(db)

	// Create test users
	user1 := models.User{Username: "user1", Password: "password123456", Firstname: "John", Lastname: "Doe", Email: "john@example.com"}
	user2 := models.User{Username: "user2", Password: "password123456", Firstname: "Jane", Lastname: "Smith", Email: "jane@example.com"}

	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)

	// Create some entries for user1
	entry1 := models.Entry{Submitter: user1.Username, Name: "Restaurant 1", Location: "Location 1", Cuisine: "Italian", Visited: true}
	entry2 := models.Entry{Submitter: user1.Username, Name: "Restaurant 2", Location: "Location 2", Cuisine: "French", Visited: false}
	require.NoError(t, db.Create(&entry1).Error)
	require.NoError(t, db.Create(&entry2).Error)

	// Create follow relationship
	require.NoError(t, service.FollowUser(user2.ID, user1.ID))

	// Test GetUserProfile
	profile, err := service.GetUserProfile(user1.ID, user2.ID)
	require.NoError(t, err)
	require.NotNil(t, profile)

	assert.Equal(t, user1.ID, profile.User.ID)
	assert.True(t, profile.IsFollowing) // user2 is following user1
	assert.Equal(t, 1, profile.FollowerCount)
	assert.Equal(t, 0, profile.FollowingCount)
	assert.Equal(t, 2, profile.EntryCount)
	assert.Len(t, profile.RecentEntries, 2)
}

func TestSocialService_GetFollowingFeed(t *testing.T) {
	db := setupSocialUnitTestDB(t)
	service := NewSocialService(db)

	// Create test users
	follower := models.User{Username: "follower", Password: "password123456", Firstname: "John", Lastname: "Doe", Email: "john@example.com"}
	followed1 := models.User{Username: "followed1", Password: "password123456", Firstname: "Jane", Lastname: "Smith", Email: "jane@example.com"}
	followed2 := models.User{Username: "followed2", Password: "password123456", Firstname: "Bob", Lastname: "Johnson", Email: "bob@example.com"}
	notFollowed := models.User{Username: "notfollowed", Password: "password123456", Firstname: "Alice", Lastname: "Brown", Email: "alice@example.com"}

	require.NoError(t, db.Create(&follower).Error)
	require.NoError(t, db.Create(&followed1).Error)
	require.NoError(t, db.Create(&followed2).Error)
	require.NoError(t, db.Create(&notFollowed).Error)

	// Create follow relationships
	require.NoError(t, service.FollowUser(follower.ID, followed1.ID))
	require.NoError(t, service.FollowUser(follower.ID, followed2.ID))

	// Create entries
	entry1 := models.Entry{Submitter: followed1.Username, Name: "Restaurant 1", Location: "Location 1", Cuisine: "Italian", Visited: true}
	entry2 := models.Entry{Submitter: followed2.Username, Name: "Restaurant 2", Location: "Location 2", Cuisine: "French", Visited: false}
	entry3 := models.Entry{Submitter: notFollowed.Username, Name: "Restaurant 3", Location: "Location 3", Cuisine: "Chinese", Visited: true}

	require.NoError(t, db.Create(&entry1).Error)
	require.NoError(t, db.Create(&entry2).Error)
	require.NoError(t, db.Create(&entry3).Error)

	// Test GetFollowingFeed
	feed, err := service.GetFollowingFeed(follower.ID, 0, 0)
	require.NoError(t, err)
	assert.Len(t, feed, 2) // Should only include entries from followed users

	// Verify entries are from followed users
	submitters := make(map[string]bool)
	for _, entry := range feed {
		submitters[entry.Submitter] = true
	}
	assert.True(t, submitters[followed1.Username])
	assert.True(t, submitters[followed2.Username])
	assert.False(t, submitters[notFollowed.Username])
}

func TestSocialService_NotifyFollowersOfNewEntry(t *testing.T) {
	db := setupSocialUnitTestDB(t)
	service := NewSocialService(db)

	// Create test users
	creator := models.User{Username: "creator", Password: "password123456", Firstname: "John", Lastname: "Doe", Email: "john@example.com"}
	follower1 := models.User{Username: "follower1", Password: "password123456", Firstname: "Jane", Lastname: "Smith", Email: "jane@example.com"}
	follower2 := models.User{Username: "follower2", Password: "password123456", Firstname: "Bob", Lastname: "Johnson", Email: "bob@example.com"}

	require.NoError(t, db.Create(&creator).Error)
	require.NoError(t, db.Create(&follower1).Error)
	require.NoError(t, db.Create(&follower2).Error)

	// Create follow relationships
	require.NoError(t, service.FollowUser(follower1.ID, creator.ID))
	require.NoError(t, service.FollowUser(follower2.ID, creator.ID))

	// Create an entry
	entry := models.Entry{Submitter: creator.Username, Name: "New Restaurant", Location: "Location", Cuisine: "Italian", Visited: true}
	require.NoError(t, db.Create(&entry).Error)

	// Test NotifyFollowersOfNewEntry
	err := service.NotifyFollowersOfNewEntry(creator.ID, entry.ID)
	require.NoError(t, err)

	// Verify notifications were created
	notifications1, err := service.GetSocialNotifications(follower1.ID, 10)
	require.NoError(t, err)
	assert.Len(t, notifications1, 1)
	assert.Equal(t, "entry", notifications1[0].Type)
	assert.Equal(t, creator.ID, notifications1[0].ActorID)

	notifications2, err := service.GetSocialNotifications(follower2.ID, 10)
	require.NoError(t, err)
	assert.Len(t, notifications2, 1)
	assert.Equal(t, "entry", notifications2[0].Type)
	assert.Equal(t, creator.ID, notifications2[0].ActorID)
}

func TestSocialService_NotificationManagement(t *testing.T) {
	db := setupSocialUnitTestDB(t)
	service := NewSocialService(db)

	// Create test users
	user1 := models.User{Username: "user1", Password: "password123456", Firstname: "John", Lastname: "Doe", Email: "john@example.com"}
	user2 := models.User{Username: "user2", Password: "password123456", Firstname: "Jane", Lastname: "Smith", Email: "jane@example.com"}

	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)

	// Create follow relationship (generates notification)
	require.NoError(t, service.FollowUser(user1.ID, user2.ID))

	// Test GetUnreadNotificationCount
	unreadCount, err := service.GetUnreadNotificationCount(user2.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, unreadCount)

	// Get notifications
	notifications, err := service.GetSocialNotifications(user2.ID, 10)
	require.NoError(t, err)
	require.Len(t, notifications, 1)

	// Test MarkNotificationAsRead
	err = service.MarkNotificationAsRead(notifications[0].ID)
	require.NoError(t, err)

	// Verify unread count decreased
	unreadCount, err = service.GetUnreadNotificationCount(user2.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, unreadCount)

	// Create another follow relationship
	user3 := models.User{Username: "user3", Password: "password123456", Firstname: "Bob", Lastname: "Johnson", Email: "bob@example.com"}
	require.NoError(t, db.Create(&user3).Error)
	require.NoError(t, service.FollowUser(user3.ID, user2.ID))

	// Test MarkAllNotificationsAsRead
	err = service.MarkAllNotificationsAsRead(user2.ID)
	require.NoError(t, err)

	unreadCount, err = service.GetUnreadNotificationCount(user2.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, unreadCount)
}

func TestSocialService_GetNotificationsByType(t *testing.T) {
	db := setupSocialUnitTestDB(t)
	service := NewSocialService(db)

	// Create test users
	user1 := models.User{Username: "user1", Password: "password123456", Firstname: "John", Lastname: "Doe", Email: "john@example.com"}
	user2 := models.User{Username: "user2", Password: "password123456", Firstname: "Jane", Lastname: "Smith", Email: "jane@example.com"}

	require.NoError(t, db.Create(&user1).Error)
	require.NoError(t, db.Create(&user2).Error)

	// Create follow relationship (generates follow notification)
	require.NoError(t, service.FollowUser(user1.ID, user2.ID))

	// Create an entry and notify followers (generates entry notification)
	entry := models.Entry{Submitter: user2.Username, Name: "Restaurant", Location: "Location", Cuisine: "Italian", Visited: true}
	require.NoError(t, db.Create(&entry).Error)
	require.NoError(t, service.NotifyFollowersOfNewEntry(user2.ID, entry.ID))

	// Test GetNotificationsByType for follow notifications
	followNotifications, err := service.GetNotificationsByType(user2.ID, "follow", 10)
	require.NoError(t, err)
	assert.Len(t, followNotifications, 1)
	assert.Equal(t, "follow", followNotifications[0].Type)

	// Test GetNotificationsByType for entry notifications
	entryNotifications, err := service.GetNotificationsByType(user1.ID, "entry", 10)
	require.NoError(t, err)
	assert.Len(t, entryNotifications, 1)
	assert.Equal(t, "entry", entryNotifications[0].Type)
}
