// Package models defines the data models for the dinnerclub application
// including database schemas and validation rules.
package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user account in the dinnerclub application with authentication
// and profile information.
type User struct {
	gorm.Model
	// Existing fields
	Username     string `gorm:"size:255;not null;unique" form:"username" json:"username" binding:"required,alphanum,lowercase"`
	Password     string `gorm:"size:255;not null;" form:"password" json:"-" binding:"required,min=10"`
	Firstname    string `gorm:"size:255" form:"firstname" json:"firstname"`
	Lastname     string `gorm:"size:255" form:"lastname" json:"lastname"`
	Email        string `gorm:"size:255;unique" form:"email" json:"email" binding:"required,email"`
	ReferralCode string `gorm:"size:255" form:"referralcode" json:"-"`

	// New enhanced fields
	EmailVerified   bool       `gorm:"default:false" json:"email_verified"`
	EmailForwarding bool       `gorm:"default:false" json:"email_forwarding"`
	Bio             string     `gorm:"size:500" form:"bio" json:"bio"`
	AvatarURL       string     `gorm:"size:255" json:"avatar_url"`
	IsFounder       bool       `gorm:"default:false" json:"is_founder"`
	JoinDate        time.Time  `gorm:"autoCreateTime" json:"join_date"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	LocationEnabled bool       `gorm:"default:false" json:"location_enabled"`
	PrivacyLevel    int        `gorm:"default:0" json:"privacy_level"` // 0=public, 1=followers, 2=private

	// 2FA fields
	TwoFactorEnabled bool   `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret  string `gorm:"size:255" json:"-"`

	// Referral relationship
	ReferredBy *uint `gorm:"index" json:"referred_by"`

	// Relationships
	Referrals      []User               `gorm:"foreignKey:ReferredBy" json:"-"`
	BackupCodes    []BackupCode         `gorm:"foreignKey:UserID" json:"-"`
	Following      []UserFollow         `gorm:"foreignKey:FollowerID" json:"-"`
	Followers      []UserFollow         `gorm:"foreignKey:FollowingID" json:"-"`
	PendingChanges []ProfileChange      `gorm:"foreignKey:UserID" json:"-"`
	Notifications  []SocialNotification `gorm:"foreignKey:UserID" json:"-"`
	PasswordResets []PasswordReset      `gorm:"foreignKey:UserID" json:"-"`
}

// BackupCode represents a single-use backup code for 2FA recovery
type BackupCode struct {
	gorm.Model
	UserID uint   `gorm:"not null;index" json:"user_id"`
	Code   string `gorm:"size:10;not null" json:"code"`
	Used   bool   `gorm:"default:false" json:"used"`
}

// UserFollow represents a following relationship between users
type UserFollow struct {
	gorm.Model
	FollowerID  uint `gorm:"not null;index" json:"follower_id"`
	FollowingID uint `gorm:"not null;index" json:"following_id"`
	Follower    User `gorm:"foreignKey:FollowerID" json:"follower"`
	Following   User `gorm:"foreignKey:FollowingID" json:"following"`
}

// ProfileChange represents a pending profile change requiring confirmation
type ProfileChange struct {
	gorm.Model
	UserID     uint      `gorm:"not null;index" json:"user_id"`
	ChangeType string    `gorm:"size:50;not null" json:"change_type"` // email, password, profile
	OldValue   string    `gorm:"size:255" json:"old_value"`
	NewValue   string    `gorm:"size:255" json:"new_value"`
	Token      string    `gorm:"size:255;unique;not null" json:"token"`
	ExpiresAt  time.Time `gorm:"not null" json:"expires_at"`
	Confirmed  bool      `gorm:"default:false" json:"confirmed"`
}

// EmailForwardingRule represents email forwarding configuration
type EmailForwardingRule struct {
	gorm.Model
	Username    string `gorm:"size:255;unique;not null" json:"username"`
	FromAddress string `gorm:"size:255;not null" json:"from_address"` // username@citycode.dinner.club
	ToAddress   string `gorm:"size:255;not null" json:"to_address"`   // user's actual email
	Active      bool   `gorm:"default:true" json:"active"`
}

// SocialNotification represents notifications for social interactions
type SocialNotification struct {
	gorm.Model
	UserID    uint   `gorm:"not null;index" json:"user_id"`
	ActorID   uint   `gorm:"not null" json:"actor_id"`
	Type      string `gorm:"size:50;not null" json:"type"` // follow, entry, referral
	Message   string `gorm:"size:255" json:"message"`
	Read      bool   `gorm:"default:false" json:"read"`
	RelatedID *uint  `json:"related_id"` // entry_id, etc.
}

// PasswordReset represents a password reset request with temporary password
type PasswordReset struct {
	gorm.Model
	UserID       uint      `gorm:"not null;index" json:"user_id"`
	Token        string    `gorm:"size:255;unique;not null" json:"token"`
	TempPassword string    `gorm:"size:255;not null" json:"-"`
	ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
	Used         bool      `gorm:"default:false" json:"used"`
}

// ReferralStats represents referral statistics for a user
type ReferralStats struct {
	UserID          uint `json:"user_id"`
	TotalReferrals  int  `json:"total_referrals"`
	ActiveReferrals int  `json:"active_referrals"`
}
