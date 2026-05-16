// Package models defines the data models for the dinnerclub application
// including database schemas and validation rules.
package models

import (
	"log"

	"github.com/toozej/dinnerclub/pkg/database"
)

// MigrateSchema performs automatic database schema migration for all models
// and creates additional search indexes for improved query performance.
func MigrateSchema() {
	var schemaModels = []interface{}{
		Entry{},
		User{},
		Restaurant{},
		BackupCode{},
		UserFollow{},
		ProfileChange{},
		EmailForwardingRule{},
		SocialNotification{},
		PasswordReset{},
	}

	// Run auto-migration for all models
	for m := range schemaModels {
		if err := database.DB.AutoMigrate(&schemaModels[m]); err != nil {
			log.Fatalf("Failed to migrate database schema: %s", err)
		}
	}

	// Create additional composite indexes for search performance
	if err := createSearchIndexes(); err != nil {
		log.Fatalf("Failed to create search indexes: %s", err)
	}

	log.Printf("Successfully migrated all database schemas and created search indexes")
}

// createSearchIndexes creates composite indexes for better search performance
func createSearchIndexes() error {
	// Composite index for entries search by name and visited status
	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_entries_name_visited ON entries(name, visited)").Error; err != nil {
		return err
	}

	// Composite index for entries search by submitter and visited status
	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_entries_submitter_visited ON entries(submitter, visited)").Error; err != nil {
		return err
	}

	// Composite index for entries search by cuisine and location
	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_entries_cuisine_location ON entries(cuisine, location)").Error; err != nil {
		return err
	}

	// Full-text search index for restaurant names (database-specific)
	// Check database type to create appropriate index
	if database.DB.Dialector.Name() == "postgres" {
		// PostgreSQL: Use GIN index with full-text search
		if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_restaurants_name_fulltext ON restaurants USING gin(to_tsvector('english', name))").Error; err != nil {
			log.Printf("GIN index creation failed: %s", err)
		}
	} else {
		// SQLite and other databases: Use regular index
		if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_restaurants_name ON restaurants(name)").Error; err != nil {
			log.Printf("Regular index creation failed: %s", err)
		}
	}

	// Indexes for social features
	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_follows_follower ON user_follows(follower_id)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_follows_following ON user_follows(following_id)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_user_follows_unique ON user_follows(follower_id, following_id)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_social_notifications_user ON social_notifications(user_id)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_social_notifications_unread ON social_notifications(user_id, read)").Error; err != nil {
		return err
	}

	// Indexes for referrals
	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_referral_code ON users(referral_code)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_users_referred_by ON users(referred_by)").Error; err != nil {
		return err
	}

	// Indexes for authentication
	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_backup_codes_user ON backup_codes(user_id)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_backup_codes_user_unused ON backup_codes(user_id, used)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_profile_changes_token ON profile_changes(token)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_profile_changes_user_pending ON profile_changes(user_id, confirmed)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_password_resets_user_unused ON password_resets(user_id, used)").Error; err != nil {
		return err
	}

	// Indexes for email forwarding
	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_email_forwarding_username ON email_forwarding_rules(username)").Error; err != nil {
		return err
	}

	if err := database.DB.Exec("CREATE INDEX IF NOT EXISTS idx_email_forwarding_active ON email_forwarding_rules(active)").Error; err != nil {
		return err
	}

	return nil
}
