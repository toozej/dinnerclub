// Package services provides business logic services for the dinnerclub application
package services

import (
	"fmt"
	mathrand "math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/toozej/dinnerclub/internal/models"
)

// setupSocialTestDB creates an in-memory SQLite database for testing
func setupSocialTestDB(t *testing.T) *gorm.DB {
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

// generateRandomSocialUser creates a random user for testing
func generateRandomSocialUser() models.User {
	id := mathrand.Uint32()
	return models.User{
		Username:  fmt.Sprintf("user%d", id),
		Password:  "password123456",
		Firstname: fmt.Sprintf("First%d", id),
		Lastname:  fmt.Sprintf("Last%d", id),
		Email:     fmt.Sprintf("user%d@example.com", id),
	}
}

// generateRandomEntry creates a random entry for testing
func generateRandomEntry(username string) models.Entry {
	id := mathrand.Uint32()
	return models.Entry{
		Submitter: username,
		Name:      fmt.Sprintf("Restaurant%d", id),
		Location:  fmt.Sprintf("Location%d", id),
		Cuisine:   fmt.Sprintf("Cuisine%d", id),
		Visited:   mathrand.IntN(2) == 0,
	}
}

// TestProperty_FollowButtonDisplay tests Property 44: Follow Button Display
// For any user profile viewed by another user, appropriate follow/unfollow buttons should be displayed
func TestProperty_FollowButtonDisplay(t *testing.T) {
	const iterations = 50

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupSocialTestDB(t)
			service := NewSocialService(db)

			// Create two random users
			user1 := generateRandomSocialUser()
			user2 := generateRandomSocialUser()

			require.NoError(t, db.Create(&user1).Error)
			require.NoError(t, db.Create(&user2).Error)

			// Test the property: GetUserProfile should indicate follow status correctly
			profile, err := service.GetUserProfile(user2.ID, user1.ID)
			require.NoError(t, err)
			require.NotNil(t, profile)

			// Initially, user1 should not be following user2
			assert.False(t, profile.IsFollowing, "Should not be following initially")

			// Follow user2
			err = service.FollowUser(user1.ID, user2.ID)
			require.NoError(t, err)

			// Now check the profile again
			profile, err = service.GetUserProfile(user2.ID, user1.ID)
			require.NoError(t, err)
			require.NotNil(t, profile)

			// Now user1 should be following user2
			assert.True(t, profile.IsFollowing, "Should be following after follow action")

			// Unfollow user2
			err = service.UnfollowUser(user1.ID, user2.ID)
			require.NoError(t, err)

			// Check the profile one more time
			profile, err = service.GetUserProfile(user2.ID, user1.ID)
			require.NoError(t, err)
			require.NotNil(t, profile)

			// Should not be following anymore
			assert.False(t, profile.IsFollowing, "Should not be following after unfollow action")
		})
	}
}

// TestProperty_FollowRelationshipManagement tests Property 45: Follow Relationship Management
// For any follow action, the user should be added to the follower's following list, and follow relationships should be managed correctly
func TestProperty_FollowRelationshipManagement(t *testing.T) {
	const iterations = 50

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupSocialTestDB(t)
			service := NewSocialService(db)

			// Create random users
			follower := generateRandomSocialUser()
			following := generateRandomSocialUser()

			require.NoError(t, db.Create(&follower).Error)
			require.NoError(t, db.Create(&following).Error)

			// Test the property: Follow relationships should be managed correctly

			// Initially, no relationship should exist
			isFollowing, err := service.IsFollowing(follower.ID, following.ID)
			require.NoError(t, err)
			assert.False(t, isFollowing, "Should not be following initially")

			// Follow the user
			err = service.FollowUser(follower.ID, following.ID)
			require.NoError(t, err)

			// Verify relationship exists
			isFollowing, err = service.IsFollowing(follower.ID, following.ID)
			require.NoError(t, err)
			assert.True(t, isFollowing, "Should be following after follow action")

			// Verify follower appears in following list
			followingList, err := service.GetFollowing(follower.ID)
			require.NoError(t, err)
			assert.Len(t, followingList, 1, "Follower should have 1 user in following list")
			assert.Equal(t, following.ID, followingList[0].ID, "Following list should contain the followed user")

			// Verify followed user appears in followers list
			followersList, err := service.GetFollowers(following.ID)
			require.NoError(t, err)
			assert.Len(t, followersList, 1, "Followed user should have 1 follower")
			assert.Equal(t, follower.ID, followersList[0].ID, "Followers list should contain the follower")

			// Unfollow the user
			err = service.UnfollowUser(follower.ID, following.ID)
			require.NoError(t, err)

			// Verify relationship no longer exists
			isFollowing, err = service.IsFollowing(follower.ID, following.ID)
			require.NoError(t, err)
			assert.False(t, isFollowing, "Should not be following after unfollow action")

			// Verify lists are empty
			followingList, err = service.GetFollowing(follower.ID)
			require.NoError(t, err)
			assert.Len(t, followingList, 0, "Following list should be empty after unfollow")

			followersList, err = service.GetFollowers(following.ID)
			require.NoError(t, err)
			assert.Len(t, followersList, 0, "Followers list should be empty after unfollow")
		})
	}
}

// TestProperty_SocialCountAccuracy tests Property 48: Social Count Accuracy
// For any user profile, the displayed follower and following counts should match the actual relationships
func TestProperty_SocialCountAccuracy(t *testing.T) {
	const iterations = 30

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupSocialTestDB(t)
			service := NewSocialService(db)

			// Create a target user
			targetUser := generateRandomSocialUser()
			require.NoError(t, db.Create(&targetUser).Error)

			// Create random number of followers (1-10)
			numFollowers := mathrand.IntN(10) + 1
			var followers []models.User
			for j := 0; j < numFollowers; j++ {
				follower := generateRandomSocialUser()
				require.NoError(t, db.Create(&follower).Error)
				followers = append(followers, follower)

				// Follow the target user
				err := service.FollowUser(follower.ID, targetUser.ID)
				require.NoError(t, err)
			}

			// Create random number of users that target user follows (1-10)
			numFollowing := mathrand.IntN(10) + 1
			var following []models.User
			for j := 0; j < numFollowing; j++ {
				followedUser := generateRandomSocialUser()
				require.NoError(t, db.Create(&followedUser).Error)
				following = append(following, followedUser)

				// Target user follows this user
				err := service.FollowUser(targetUser.ID, followedUser.ID)
				require.NoError(t, err)
			}

			// Create some entries for the target user
			numEntries := mathrand.IntN(5) + 1
			for j := 0; j < numEntries; j++ {
				entry := generateRandomEntry(targetUser.Username)
				require.NoError(t, db.Create(&entry).Error)
			}

			// Test the property: Profile counts should match actual relationships
			profile, err := service.GetUserProfile(targetUser.ID, 0)
			require.NoError(t, err)
			require.NotNil(t, profile)

			// Verify follower count accuracy
			assert.Equal(t, numFollowers, profile.FollowerCount,
				"Profile follower count should match actual number of followers")

			// Verify following count accuracy
			assert.Equal(t, numFollowing, profile.FollowingCount,
				"Profile following count should match actual number of users being followed")

			// Verify entry count accuracy
			assert.Equal(t, numEntries, profile.EntryCount,
				"Profile entry count should match actual number of entries")

			// Double-check by getting actual lists
			actualFollowers, err := service.GetFollowers(targetUser.ID)
			require.NoError(t, err)
			assert.Len(t, actualFollowers, numFollowers, "Actual followers list should match expected count")

			actualFollowing, err := service.GetFollowing(targetUser.ID)
			require.NoError(t, err)
			assert.Len(t, actualFollowing, numFollowing, "Actual following list should match expected count")
		})
	}
}

// TestProperty_FollowSelfPrevention tests that users cannot follow themselves
func TestProperty_FollowSelfPrevention(t *testing.T) {
	const iterations = 20

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupSocialTestDB(t)
			service := NewSocialService(db)

			// Create a random user
			user := generateRandomSocialUser()
			require.NoError(t, db.Create(&user).Error)

			// Test the property: Users should not be able to follow themselves
			err := service.FollowUser(user.ID, user.ID)
			assert.Error(t, err, "Following oneself should return an error")
			assert.Contains(t, err.Error(), "cannot follow themselves", "Error should indicate self-follow prevention")

			// Verify no self-follow relationship was created
			isFollowing, err := service.IsFollowing(user.ID, user.ID)
			require.NoError(t, err)
			assert.False(t, isFollowing, "User should not be following themselves")

			// Verify counts remain zero
			profile, err := service.GetUserProfile(user.ID, user.ID)
			require.NoError(t, err)
			assert.Equal(t, 0, profile.FollowerCount, "Self-follow should not affect follower count")
			assert.Equal(t, 0, profile.FollowingCount, "Self-follow should not affect following count")
		})
	}
}

// TestProperty_DuplicateFollowPrevention tests that duplicate follow relationships are prevented
func TestProperty_DuplicateFollowPrevention(t *testing.T) {
	const iterations = 20

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupSocialTestDB(t)
			service := NewSocialService(db)

			// Create two random users
			follower := generateRandomSocialUser()
			following := generateRandomSocialUser()

			require.NoError(t, db.Create(&follower).Error)
			require.NoError(t, db.Create(&following).Error)

			// First follow should succeed
			err := service.FollowUser(follower.ID, following.ID)
			require.NoError(t, err)

			// Test the property: Duplicate follows should be prevented
			err = service.FollowUser(follower.ID, following.ID)
			assert.Error(t, err, "Duplicate follow should return an error")
			assert.Contains(t, err.Error(), "already following", "Error should indicate duplicate follow")

			// Verify only one relationship exists
			followers, err := service.GetFollowers(following.ID)
			require.NoError(t, err)
			assert.Len(t, followers, 1, "Should have exactly one follower")

			following_list, err := service.GetFollowing(follower.ID)
			require.NoError(t, err)
			assert.Len(t, following_list, 1, "Should be following exactly one user")
		})
	}
}

// TestProperty_SocialFeedContent tests Property 46: Social Feed Content
// For any user's following feed, it should contain entries from users they follow, sorted by creation time
func TestProperty_SocialFeedContent(t *testing.T) {
	const iterations = 30

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupSocialTestDB(t)
			service := NewSocialService(db)

			// Create a follower user
			follower := generateRandomSocialUser()
			require.NoError(t, db.Create(&follower).Error)

			// Create random number of users to follow (1-5)
			numFollowing := mathrand.IntN(5) + 1
			var followedUsers []models.User
			var expectedEntries []models.Entry

			for j := 0; j < numFollowing; j++ {
				followedUser := generateRandomSocialUser()
				require.NoError(t, db.Create(&followedUser).Error)
				followedUsers = append(followedUsers, followedUser)

				// Follow this user
				err := service.FollowUser(follower.ID, followedUser.ID)
				require.NoError(t, err)

				// Create random entries for this followed user (1-3 entries)
				numEntries := mathrand.IntN(3) + 1
				for k := 0; k < numEntries; k++ {
					entry := generateRandomEntry(followedUser.Username)
					require.NoError(t, db.Create(&entry).Error)
					expectedEntries = append(expectedEntries, entry)
				}
			}

			// Create some entries from users not followed (should not appear in feed)
			nonFollowedUser := generateRandomSocialUser()
			require.NoError(t, db.Create(&nonFollowedUser).Error)
			nonFollowedEntry := generateRandomEntry(nonFollowedUser.Username)
			require.NoError(t, db.Create(&nonFollowedEntry).Error)

			// Test the property: Feed should contain only entries from followed users
			feedEntries, err := service.GetFollowingFeed(follower.ID, 0, 0)
			require.NoError(t, err)

			// Verify feed contains expected number of entries
			assert.Len(t, feedEntries, len(expectedEntries),
				"Feed should contain entries only from followed users")

			// Verify all entries in feed are from followed users
			followedUsernames := make(map[string]bool)
			for _, user := range followedUsers {
				followedUsernames[user.Username] = true
			}

			for _, entry := range feedEntries {
				assert.True(t, followedUsernames[entry.Submitter],
					"Feed entry should be from a followed user")
			}

			// Verify entries are sorted by creation time (newest first)
			for k := 1; k < len(feedEntries); k++ {
				assert.True(t, feedEntries[k-1].CreatedAt.After(feedEntries[k].CreatedAt) ||
					feedEntries[k-1].CreatedAt.Equal(feedEntries[k].CreatedAt),
					"Feed entries should be sorted by creation time (newest first)")
			}

			// Verify non-followed user's entry is not in feed
			for _, entry := range feedEntries {
				assert.NotEqual(t, nonFollowedUser.Username, entry.Submitter,
					"Feed should not contain entries from non-followed users")
			}
		})
	}
}

// TestProperty_FollowerNotifications tests Property 47: Follower Notifications
// For any new content from followed users, followers should receive notifications
func TestProperty_FollowerNotifications(t *testing.T) {
	const iterations = 30

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupSocialTestDB(t)
			service := NewSocialService(db)

			// Create a content creator
			creator := generateRandomSocialUser()
			require.NoError(t, db.Create(&creator).Error)

			// Create random number of followers (1-5)
			numFollowers := mathrand.IntN(5) + 1
			var followers []models.User

			for j := 0; j < numFollowers; j++ {
				follower := generateRandomSocialUser()
				require.NoError(t, db.Create(&follower).Error)
				followers = append(followers, follower)

				// Follow the creator
				err := service.FollowUser(follower.ID, creator.ID)
				require.NoError(t, err)
			}

			// Create an entry for the creator
			entry := generateRandomEntry(creator.Username)
			require.NoError(t, db.Create(&entry).Error)

			// Test the property: Notify followers of new entry
			err := service.NotifyFollowersOfNewEntry(creator.ID, entry.ID)
			require.NoError(t, err)

			// Verify each follower received a notification
			for _, follower := range followers {
				notifications, err := service.GetSocialNotifications(follower.ID, 10)
				require.NoError(t, err)

				// Should have at least one notification
				assert.GreaterOrEqual(t, len(notifications), 1,
					"Follower should receive notification for new entry")

				// Find the entry notification
				var foundEntryNotification bool
				for _, notification := range notifications {
					if notification.Type == "entry" &&
						notification.ActorID == creator.ID &&
						notification.RelatedID != nil &&
						*notification.RelatedID == entry.ID {
						foundEntryNotification = true
						assert.Contains(t, notification.Message, creator.Firstname,
							"Notification should contain creator's name")
						assert.Contains(t, notification.Message, entry.Name,
							"Notification should contain entry name")
						assert.False(t, notification.Read,
							"New notification should be unread")
						break
					}
				}
				assert.True(t, foundEntryNotification,
					"Should find entry notification for follower")
			}
		})
	}
}

// TestProperty_FollowNotificationDelivery tests Property 49: Follow Notification Delivery
// For any follow action, the followed user should receive a notification
func TestProperty_FollowNotificationDelivery(t *testing.T) {
	const iterations = 30

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupSocialTestDB(t)
			service := NewSocialService(db)

			// Create two random users
			follower := generateRandomSocialUser()
			following := generateRandomSocialUser()

			require.NoError(t, db.Create(&follower).Error)
			require.NoError(t, db.Create(&following).Error)

			// Test the property: Following a user should send notification to followed user
			err := service.FollowUser(follower.ID, following.ID)
			require.NoError(t, err)

			// Verify the followed user received a notification
			notifications, err := service.GetSocialNotifications(following.ID, 10)
			require.NoError(t, err)

			// Should have at least one notification
			assert.GreaterOrEqual(t, len(notifications), 1,
				"Followed user should receive follow notification")

			// Find the follow notification
			var foundFollowNotification bool
			for _, notification := range notifications {
				if notification.Type == "follow" &&
					notification.ActorID == follower.ID {
					foundFollowNotification = true
					assert.Contains(t, notification.Message, follower.Firstname,
						"Notification should contain follower's name")
					assert.Contains(t, notification.Message, "following",
						"Notification should indicate follow action")
					assert.False(t, notification.Read,
						"New notification should be unread")
					assert.Equal(t, follower.ID, *notification.RelatedID,
						"Notification should reference the follower")
					break
				}
			}
			assert.True(t, foundFollowNotification,
				"Should find follow notification for followed user")
		})
	}
}

// TestProperty_NotificationManagement tests notification read/unread functionality
func TestProperty_NotificationManagement(t *testing.T) {
	const iterations = 20

	for i := 0; i < iterations; i++ {
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			// Setup
			db := setupSocialTestDB(t)
			service := NewSocialService(db)

			// Create a user
			user := generateRandomSocialUser()
			require.NoError(t, db.Create(&user).Error)

			// Create some followers to generate notifications
			numFollowers := mathrand.IntN(5) + 1
			for j := 0; j < numFollowers; j++ {
				follower := generateRandomSocialUser()
				require.NoError(t, db.Create(&follower).Error)

				// Follow the user (generates notification)
				err := service.FollowUser(follower.ID, user.ID)
				require.NoError(t, err)
			}

			// Test the property: Unread count should match actual unread notifications
			unreadCount, err := service.GetUnreadNotificationCount(user.ID)
			require.NoError(t, err)
			assert.Equal(t, numFollowers, unreadCount,
				"Unread count should match number of follow notifications")

			// Mark all notifications as read
			err = service.MarkAllNotificationsAsRead(user.ID)
			require.NoError(t, err)

			// Verify unread count is now zero
			unreadCount, err = service.GetUnreadNotificationCount(user.ID)
			require.NoError(t, err)
			assert.Equal(t, 0, unreadCount,
				"Unread count should be zero after marking all as read")

			// Verify all notifications are marked as read
			notifications, err := service.GetSocialNotifications(user.ID, 0)
			require.NoError(t, err)
			for _, notification := range notifications {
				assert.True(t, notification.Read,
					"All notifications should be marked as read")
			}
		})
	}
}
