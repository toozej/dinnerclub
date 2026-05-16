# Design Document

## Overview

This design document outlines the comprehensive advanced user account system and social profile functionality for the Dinnerclub Golang web application. The design introduces sophisticated user management features including referral systems, email forwarding, two-factor authentication, location-based services, and social networking capabilities while maintaining the existing Gin web framework architecture.

The design emphasizes security, scalability, and user experience through modern authentication practices, efficient database design, and responsive user interfaces.

## Architecture

### Current Architecture
- **Web Framework**: Gin (Go HTTP web framework)
- **Database**: GORM with PostgreSQL
- **Templates**: Go HTML templates
- **Authentication**: JWT-based authentication system
- **Static Assets**: Embedded assets package

### Enhanced Architecture Components
- **Referral Service**: Referral code generation and tracking system
- **Email Service**: SMTP forwarding and notification system
- **Two-Factor Authentication**: TOTP-based 2FA with backup codes
- **Avatar Service**: Gravatar integration and custom avatar management
- **Location Service**: GPS-based restaurant discovery system
- **Social Service**: Following/follower relationship management
- **Profile Service**: Enhanced user profile management with confirmations

## Components and Interfaces

### 1. Enhanced User Model

#### Extended User Model
```go
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
    EmailVerified     bool      `gorm:"default:false" json:"email_verified"`
    EmailForwarding   bool      `gorm:"default:false" json:"email_forwarding"`
    Bio              string    `gorm:"size:500" form:"bio" json:"bio"`
    AvatarURL        string    `gorm:"size:255" json:"avatar_url"`
    IsFounder        bool      `gorm:"default:false" json:"is_founder"`
    JoinDate         time.Time `gorm:"autoCreateTime" json:"join_date"`
    LastLoginAt      *time.Time `json:"last_login_at"`
    LocationEnabled  bool      `gorm:"default:false" json:"location_enabled"`
    PrivacyLevel     int       `gorm:"default:0" json:"privacy_level"` // 0=public, 1=followers, 2=private
    
    // 2FA fields
    TwoFactorEnabled bool   `gorm:"default:false" json:"two_factor_enabled"`
    TwoFactorSecret  string `gorm:"size:255" json:"-"`
    
    // Relationships
    ReferredBy       *uint `gorm:"index" json:"referred_by"`
    Referrals        []User `gorm:"foreignKey:ReferredBy" json:"-"`
    BackupCodes      []BackupCode `gorm:"foreignKey:UserID" json:"-"`
    Following        []UserFollow `gorm:"foreignKey:FollowerID" json:"-"`
    Followers        []UserFollow `gorm:"foreignKey:FollowingID" json:"-"`
    PendingChanges   []ProfileChange `gorm:"foreignKey:UserID" json:"-"`
}
```

#### Supporting Models
```go
type BackupCode struct {
    gorm.Model
    UserID uint   `gorm:"not null;index" json:"user_id"`
    Code   string `gorm:"size:10;not null" json:"code"`
    Used   bool   `gorm:"default:false" json:"used"`
}

type UserFollow struct {
    gorm.Model
    FollowerID  uint `gorm:"not null;index" json:"follower_id"`
    FollowingID uint `gorm:"not null;index" json:"following_id"`
    Follower    User `gorm:"foreignKey:FollowerID" json:"follower"`
    Following   User `gorm:"foreignKey:FollowingID" json:"following"`
}

type ProfileChange struct {
    gorm.Model
    UserID       uint      `gorm:"not null;index" json:"user_id"`
    ChangeType   string    `gorm:"size:50;not null" json:"change_type"` // email, password, profile
    OldValue     string    `gorm:"size:255" json:"old_value"`
    NewValue     string    `gorm:"size:255" json:"new_value"`
    Token        string    `gorm:"size:255;unique;not null" json:"token"`
    ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
    Confirmed    bool      `gorm:"default:false" json:"confirmed"`
}

type ReferralStats struct {
    UserID          uint `json:"user_id"`
    TotalReferrals  int  `json:"total_referrals"`
    ActiveReferrals int  `json:"active_referrals"`
}
```

### 2. Referral System

#### Referral Service
```go
type ReferralService struct {
    db *gorm.DB
}

func (r *ReferralService) GenerateReferralCode(userID uint) (string, error)
func (r *ReferralService) GenerateReferralMessage(userID uint, customMessage string) (string, error)
func (r *ReferralService) ProcessReferral(referralCode string, newUserID uint) error
func (r *ReferralService) GetReferralStats(userID uint) (*ReferralStats, error)
func (r *ReferralService) GetReferralHistory(userID uint) ([]User, error)
```

#### Referral Message Template
```go
type ReferralMessage struct {
    ReferrerName string
    ReferralCode string
    CustomText   string
    JoinURL      string
}

const ReferralTemplate = `
Hi! {{.ReferrerName}} has invited you to join {{.CityCode}}dinner.club - 
a community for discovering and sharing great restaurants in {{.CityName}}.

{{.CustomText}}

Join using this link: {{.JoinURL}}?ref={{.ReferralCode}}

Happy dining!
`
```

### 3. Email Service and Forwarding

#### Email Service
```go
type EmailService struct {
    smtpHost     string
    smtpPort     int
    smtpUsername string
    smtpPassword string
    fromAddress  string
    cityCode     string
}

func (e *EmailService) SendConfirmationEmail(user *User, changeType string, token string) error
func (e *EmailService) SendPasswordResetEmail(user *User, tempPassword string) error
func (e *EmailService) SendNotificationEmail(user *User, subject, body string) error
func (e *EmailService) SetupEmailForwarding(username, email string) error
func (e *EmailService) RemoveEmailForwarding(username string) error
```

#### Email Forwarding Configuration
```go
type EmailForwardingRule struct {
    gorm.Model
    Username    string `gorm:"size:255;unique;not null" json:"username"`
    FromAddress string `gorm:"size:255;not null" json:"from_address"` // username@citycode.dinner.club
    ToAddress   string `gorm:"size:255;not null" json:"to_address"`   // user's actual email
    Active      bool   `gorm:"default:true" json:"active"`
}
```

### 4. Two-Factor Authentication System

#### TOTP Service
```go
type TOTPService struct {
    issuer string
}

func (t *TOTPService) GenerateSecret(username string) (string, string, error) // secret, qr_url
func (t *TOTPService) ValidateToken(secret, token string) bool
func (t *TOTPService) GenerateBackupCodes(userID uint) ([]string, error)
func (t *TOTPService) ValidateBackupCode(userID uint, code string) (bool, error)
func (t *TOTPService) RegenerateBackupCodes(userID uint) ([]string, error)
```

#### 2FA Authentication Flow
```go
type TwoFactorAuth struct {
    UserID      uint   `json:"user_id"`
    Secret      string `json:"-"`
    QRCodeURL   string `json:"qr_code_url"`
    BackupCodes []string `json:"backup_codes"`
    Verified    bool   `json:"verified"`
}

type LoginRequest struct {
    Username   string `json:"username" binding:"required"`
    Password   string `json:"password" binding:"required"`
    TOTPCode   string `json:"totp_code"`
    BackupCode string `json:"backup_code"`
}
```

### 5. Avatar and Gravatar System

#### Avatar Service
```go
type AvatarService struct {
    uploadPath   string
    maxFileSize  int64
    allowedTypes []string
}

func (a *AvatarService) GetGravatarURL(email string, size int) string
func (a *AvatarService) UploadCustomAvatar(userID uint, file multipart.File, header *multipart.FileHeader) (string, error)
func (a *AvatarService) ResizeImage(src image.Image, width, height int) image.Image
func (a *AvatarService) GetUserAvatarURL(user *User, size int) string
func (a *AvatarService) DeleteCustomAvatar(userID uint) error
```

#### Avatar Configuration
```go
type AvatarConfig struct {
    MaxFileSize    int64    `json:"max_file_size"`    // 5MB
    AllowedTypes   []string `json:"allowed_types"`    // jpg, png, gif
    DefaultSize    int      `json:"default_size"`     // 150px
    ThumbnailSizes []int    `json:"thumbnail_sizes"`  // 50, 100, 150, 300
}
```

### 6. Location-Based Services

#### Location Service
```go
type LocationService struct {
    db *gorm.DB
}

type Location struct {
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
}

type RestaurantDistance struct {
    Restaurant models.Restaurant `json:"restaurant"`
    Distance   float64          `json:"distance"` // in kilometers
    Visited    bool             `json:"visited"`
    Wishlist   bool             `json:"wishlist"`
}

func (l *LocationService) CalculateDistance(loc1, loc2 Location) float64
func (l *LocationService) FindNearbyRestaurants(userLocation Location, radius float64) ([]RestaurantDistance, error)
func (l *LocationService) FindNearestWishlist(userID uint, userLocation Location) ([]RestaurantDistance, error)
func (l *LocationService) FindNearestUnvisited(userID uint, userLocation Location) ([]RestaurantDistance, error)
```

#### Haversine Distance Calculation
```go
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
    const R = 6371 // Earth's radius in kilometers
    
    dLat := (lat2 - lat1) * math.Pi / 180
    dLon := (lon2 - lon1) * math.Pi / 180
    
    a := math.Sin(dLat/2)*math.Sin(dLat/2) +
        math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
        math.Sin(dLon/2)*math.Sin(dLon/2)
    
    c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    
    return R * c
}
```

### 7. Social Following System

#### Social Service
```go
type SocialService struct {
    db *gorm.DB
}

type UserProfile struct {
    User         User   `json:"user"`
    IsFollowing  bool   `json:"is_following"`
    FollowerCount int   `json:"follower_count"`
    FollowingCount int  `json:"following_count"`
    EntryCount   int    `json:"entry_count"`
    RecentEntries []models.Entry `json:"recent_entries"`
}

func (s *SocialService) FollowUser(followerID, followingID uint) error
func (s *SocialService) UnfollowUser(followerID, followingID uint) error
func (s *SocialService) GetFollowers(userID uint) ([]User, error)
func (s *SocialService) GetFollowing(userID uint) ([]User, error)
func (s *SocialService) GetUserProfile(userID, viewerID uint) (*UserProfile, error)
func (s *SocialService) GetFollowingFeed(userID uint, limit, offset int) ([]models.Entry, error)
```

#### Social Notifications
```go
type SocialNotification struct {
    gorm.Model
    UserID       uint   `gorm:"not null;index" json:"user_id"`
    ActorID      uint   `gorm:"not null" json:"actor_id"`
    Type         string `gorm:"size:50;not null" json:"type"` // follow, entry, referral
    Message      string `gorm:"size:255" json:"message"`
    Read         bool   `gorm:"default:false" json:"read"`
    RelatedID    *uint  `json:"related_id"` // entry_id, etc.
}
```

### 8. Profile Management and Confirmations

#### Profile Service
```go
type ProfileService struct {
    db           *gorm.DB
    emailService *EmailService
}

func (p *ProfileService) RequestEmailChange(userID uint, newEmail string) (string, error)
func (p *ProfileService) ConfirmEmailChange(token string) error
func (p *ProfileService) RequestPasswordReset(email string) error
func (p *ProfileService) UpdateProfile(userID uint, updates map[string]interface{}) error
func (p *ProfileService) GenerateConfirmationToken() string
func (p *ProfileService) ValidateConfirmationToken(token string) (*ProfileChange, error)
```

#### Password Reset System
```go
type PasswordReset struct {
    gorm.Model
    UserID       uint      `gorm:"not null;index" json:"user_id"`
    Token        string    `gorm:"size:255;unique;not null" json:"token"`
    TempPassword string    `gorm:"size:255;not null" json:"-"`
    ExpiresAt    time.Time `gorm:"not null" json:"expires_at"`
    Used         bool      `gorm:"default:false" json:"used"`
}
```

## Data Models

### Database Schema Changes

#### New Tables
```sql
-- Backup codes for 2FA
CREATE TABLE backup_codes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code VARCHAR(10) NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- User following relationships
CREATE TABLE user_follows (
    id SERIAL PRIMARY KEY,
    follower_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(follower_id, following_id)
);

-- Profile change confirmations
CREATE TABLE profile_changes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    change_type VARCHAR(50) NOT NULL,
    old_value VARCHAR(255),
    new_value VARCHAR(255),
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    confirmed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Email forwarding rules
CREATE TABLE email_forwarding_rules (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    from_address VARCHAR(255) NOT NULL,
    to_address VARCHAR(255) NOT NULL,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Social notifications
CREATE TABLE social_notifications (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    message VARCHAR(255),
    read BOOLEAN DEFAULT FALSE,
    related_id INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Password reset tokens
CREATE TABLE password_resets (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    temp_password VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

#### Enhanced User Table
```sql
-- Add new columns to existing users table
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN email_forwarding BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN bio VARCHAR(500);
ALTER TABLE users ADD COLUMN avatar_url VARCHAR(255);
ALTER TABLE users ADD COLUMN is_founder BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN join_date TIMESTAMP DEFAULT NOW();
ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP;
ALTER TABLE users ADD COLUMN location_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN privacy_level INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN two_factor_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN two_factor_secret VARCHAR(255);
ALTER TABLE users ADD COLUMN referred_by INTEGER REFERENCES users(id);
```

#### Indexes for Performance
```sql
-- Indexes for social features
CREATE INDEX idx_user_follows_follower ON user_follows(follower_id);
CREATE INDEX idx_user_follows_following ON user_follows(following_id);
CREATE INDEX idx_social_notifications_user ON social_notifications(user_id);
CREATE INDEX idx_social_notifications_unread ON social_notifications(user_id, read);

-- Indexes for referrals
CREATE INDEX idx_users_referral_code ON users(referral_code);
CREATE INDEX idx_users_referred_by ON users(referred_by);

-- Indexes for authentication
CREATE INDEX idx_backup_codes_user ON backup_codes(user_id);
CREATE INDEX idx_profile_changes_token ON profile_changes(token);
CREATE INDEX idx_password_resets_token ON password_resets(token);
```

## Error Handling

### Comprehensive Error Types
```go
type UserError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Field   string `json:"field,omitempty"`
}

var (
    ErrUserNotFound        = &UserError{"USER_NOT_FOUND", "User not found", ""}
    ErrInvalidCredentials  = &UserError{"INVALID_CREDENTIALS", "Invalid username or password", ""}
    ErrEmailAlreadyExists  = &UserError{"EMAIL_EXISTS", "Email address already registered", "email"}
    ErrInvalidTOTP         = &UserError{"INVALID_TOTP", "Invalid 2FA code", "totp_code"}
    ErrInvalidBackupCode   = &UserError{"INVALID_BACKUP", "Invalid backup code", "backup_code"}
    ErrEmailNotVerified    = &UserError{"EMAIL_NOT_VERIFIED", "Email address not verified", "email"}
    ErrFollowSelf          = &UserError{"FOLLOW_SELF", "Cannot follow yourself", ""}
    ErrAlreadyFollowing    = &UserError{"ALREADY_FOLLOWING", "Already following this user", ""}
    ErrLocationDisabled    = &UserError{"LOCATION_DISABLED", "Location services disabled", ""}
    ErrInvalidToken        = &UserError{"INVALID_TOKEN", "Invalid or expired token", "token"}
)
```

### Security Error Handling
- **Rate Limiting**: Implement rate limiting for sensitive operations
- **Secure Error Messages**: Avoid leaking information in error responses
- **Audit Logging**: Log security-related events for monitoring
- **Input Validation**: Comprehensive validation with sanitization

## Testing Strategy

### Unit Tests
- **Service Layer Tests**: Comprehensive testing of all service methods
- **Model Validation**: Test all validation rules and constraints
- **Authentication Tests**: JWT, 2FA, and password reset functionality
- **Location Calculation**: Test distance calculations and nearby searches
- **Email Service**: Mock SMTP testing for email functionality

### Integration Tests
- **API Endpoint Tests**: Full request/response cycle testing
- **Database Integration**: Test all CRUD operations and relationships
- **Authentication Flow**: Complete login/logout with 2FA testing
- **Social Features**: Follow/unfollow and feed generation testing
- **Email Forwarding**: Test email routing and delivery

### Security Tests
- **2FA Security**: Test TOTP generation and validation
- **Token Security**: Test JWT and confirmation token handling
- **Rate Limiting**: Test abuse prevention mechanisms
- **Input Validation**: Test against injection and XSS attacks

## Performance Considerations

### Database Optimization
- **Query Optimization**: Efficient queries for social feeds and location searches
- **Index Strategy**: Comprehensive indexing for common access patterns
- **Connection Pooling**: Optimal database connection management
- **Caching Strategy**: Cache user profiles and social relationships

### Location Service Performance
- **Spatial Indexing**: Use PostGIS for efficient location queries
- **Distance Calculation**: Optimize haversine calculations
- **Result Caching**: Cache location-based search results
- **Batch Processing**: Process multiple location queries efficiently

### Social Feed Performance
- **Feed Generation**: Efficient algorithms for generating user feeds
- **Pagination**: Implement cursor-based pagination for large datasets
- **Real-time Updates**: Consider WebSocket connections for live updates
- **Content Caching**: Cache frequently accessed user content

## Security Considerations

### Authentication Security
- **2FA Implementation**: Secure TOTP with proper secret management
- **Backup Codes**: Secure generation and storage of recovery codes
- **Session Management**: Secure JWT handling with proper expiration
- **Password Security**: Strong password requirements and secure hashing

### Data Protection
- **Email Security**: Secure email forwarding and SMTP configuration
- **Profile Privacy**: Respect user privacy settings and data access
- **Location Privacy**: Secure handling of GPS coordinates
- **Social Privacy**: Implement proper access controls for social features

### Communication Security
- **Email Encryption**: Use TLS for all email communications
- **API Security**: Implement proper authentication for all endpoints
- **Token Security**: Secure generation and validation of all tokens
- **Rate Limiting**: Prevent abuse of sensitive operations

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property-Based Testing Overview

Property-based testing (PBT) validates software correctness by testing universal properties across many generated inputs. Each property is a formal specification that should hold for all valid inputs.

### Correctness Properties

#### Referral System Properties

**Property 1: Referral Message Generation**
*For any* user with a valid referral code, generating a referral message should always include their referral code and proper invitation formatting
**Validates: Requirements 1.2, 1.3**

**Property 2: Referral Tracking Consistency**
*For any* valid referral code used during signup, the system should correctly establish and track the referral relationship between referrer and referee
**Validates: Requirements 1.4**

**Property 3: Referral Statistics Accuracy**
*For any* user, their displayed referral count should always match the actual number of users they have successfully referred
**Validates: Requirements 1.5**

**Property 4: Referral Notification Delivery**
*For any* successful referral, both the referrer and referee should receive appropriate notifications
**Validates: Requirements 1.6**

#### Profile Change Confirmation Properties

**Property 5: Email Change Confirmation**
*For any* email address change request, a confirmation email should be sent to the new email address with a valid verification token
**Validates: Requirements 2.1**

**Property 6: Password Change Notification**
*For any* password change, a notification email should be sent to the user's current email address
**Validates: Requirements 2.2**

**Property 7: Sensitive Change Confirmation Requirement**
*For any* sensitive profile change, the system should require email confirmation before applying the changes
**Validates: Requirements 2.3**

**Property 8: Confirmation Token Security**
*For any* confirmation email sent, it should contain a secure verification link with proper expiration time
**Validates: Requirements 2.4**

**Property 9: Valid Token Processing**
*For any* valid confirmation token, clicking the link should verify the token and successfully apply the pending changes
**Validates: Requirements 2.5**

**Property 10: Expired Token Rejection**
*For any* expired confirmation token, the system should reject the change request and require a new confirmation
**Validates: Requirements 2.6**

#### Email Management Properties

**Property 11: Email Validation Consistency**
*For any* email address input, the system should consistently validate format and deliverability according to standard email validation rules
**Validates: Requirements 3.1**

**Property 12: Email Forwarding Rule Creation**
*For any* verified user email, the system should create a corresponding forwarding rule from username@${CITY_CODE}dinner.club to their registered email
**Validates: Requirements 3.2**

**Property 13: Email Forwarding Functionality**
*For any* email sent to username@${CITY_CODE}dinner.club where forwarding is enabled, it should be forwarded to the user's registered email address
**Validates: Requirements 3.3**

**Property 14: Forwarding Rule Updates**
*For any* confirmed email change, the email forwarding rule should be updated to reflect the new email address
**Validates: Requirements 3.4**

**Property 15: Forwarding Rule Cleanup**
*For any* account deactivation, the associated email forwarding rule should be removed from the system
**Validates: Requirements 3.5**

**Property 16: Email Forwarding Error Handling**
*For any* email forwarding failure, the system should log the error and notify the affected user
**Validates: Requirements 3.6**

#### Password Reset Properties

**Property 17: Password Reset Email Delivery**
*For any* valid password reset request, a temporary password should be sent to the user's registered email address
**Validates: Requirements 4.1**

**Property 18: Temporary Password Security**
*For any* generated temporary password, it should be cryptographically secure, random, and have a proper expiration time
**Validates: Requirements 4.2**

**Property 19: Temporary Password Force Change**
*For any* login using a temporary password, the system should require the user to change their password before proceeding
**Validates: Requirements 4.3**

**Property 20: Temporary Password Expiration**
*For any* expired temporary password, login attempts should be rejected and require a new reset request
**Validates: Requirements 4.4**

**Property 21: Session Invalidation on Password Reset**
*For any* successful password reset, all existing user sessions should be invalidated for security
**Validates: Requirements 4.5**

**Property 22: Password Reset Rate Limiting**
*For any* series of password reset requests from the same source, the system should enforce rate limiting to prevent abuse
**Validates: Requirements 4.6**

#### Two-Factor Authentication Properties

**Property 23: 2FA QR Code Generation**
*For any* user enabling 2FA, the system should generate a valid QR code for authenticator app setup
**Validates: Requirements 5.1**

**Property 24: 2FA Setup Verification Requirement**
*For any* 2FA setup attempt, the system should require successful verification of the authenticator before activation
**Validates: Requirements 5.2**

**Property 25: Backup Code Generation**
*For any* 2FA activation, the system should generate exactly 10 unique single-use backup codes
**Validates: Requirements 5.3**

**Property 26: 2FA Login Requirements**
*For any* login attempt by a user with 2FA enabled, both password and valid 2FA code should be required
**Validates: Requirements 5.4**

**Property 27: Backup Code Single Use**
*For any* backup code used for authentication, that specific code should be invalidated and cannot be reused
**Validates: Requirements 5.5**

**Property 28: Backup Code Exhaustion Handling**
*For any* user who has used all backup codes, the system should prompt them to generate new backup codes
**Validates: Requirements 5.6**

**Property 29: 2FA Disable Security**
*For any* 2FA disable request, the system should require both current password and valid 2FA verification
**Validates: Requirements 5.7**

#### Founder Status Properties

**Property 30: Founder Status Display**
*For any* user with founder status, a star icon should be displayed next to their username across all interfaces
**Validates: Requirements 6.1, 6.2, 6.3**

**Property 31: Founder Profile Information**
*For any* founder user's profile, it should display their founder badge and join date prominently
**Validates: Requirements 6.4**

**Property 32: Founder Status Notification**
*For any* user granted founder status, they should receive a congratulatory notification
**Validates: Requirements 6.5**

**Property 33: Founder Statistics Accuracy**
*For any* founder statistics display, the total count should match the actual number of users with founder status
**Validates: Requirements 6.6**

#### Avatar System Properties

**Property 34: Gravatar Fallback**
*For any* user without a custom avatar, their Gravatar should be displayed based on their email address
**Validates: Requirements 7.1**

**Property 35: Avatar Upload Processing**
*For any* custom avatar upload, the image should be resized and optimized according to system specifications
**Validates: Requirements 7.2**

**Property 36: Avatar Display Consistency**
*For any* user avatar, it should be displayed consistently across all interface locations
**Validates: Requirements 7.3**

**Property 37: Gravatar Email Update**
*For any* user email change, their Gravatar should be automatically updated to reflect the new email address
**Validates: Requirements 7.4**

**Property 38: Avatar Upload Validation**
*For any* avatar upload attempt, the system should validate file type and size according to configured limits
**Validates: Requirements 7.5**

**Property 39: Avatar Loading Error Handling**
*For any* avatar that fails to load, a default placeholder image should be displayed instead
**Validates: Requirements 7.6**

#### Location Service Properties

**Property 40: Distance Calculation Accuracy**
*For any* two geographic coordinates, the calculated distance should be mathematically accurate using the Haversine formula
**Validates: Requirements 8.6**

**Property 41: Wishlist Distance Calculation**
*For any* user location and wishlist, distances to all wishlist restaurants should be calculated correctly
**Validates: Requirements 8.2**

**Property 42: Location Result Sorting**
*For any* location-based search results, unvisited restaurants should be sorted by distance with closest first
**Validates: Requirements 8.3**

**Property 43: Wishlist Location Search**
*For any* "closest wishlist" search, results should only include restaurants the user wants to visit, sorted by distance
**Validates: Requirements 8.4**

#### Social System Properties

**Property 44: Follow Button Display**
*For any* user profile viewed by another user, appropriate follow/unfollow buttons should be displayed
**Validates: Requirements 9.1**

**Property 45: Follow Relationship Management**
*For any* follow action, the user should be added to the follower's following list, and for any unfollow action, they should be removed
**Validates: Requirements 9.2, 9.3**

**Property 46: Social Feed Content**
*For any* user's social feed, it should display recent entries from all users they follow
**Validates: Requirements 9.4**

**Property 47: Follower Notifications**
*For any* new content posted by a user, all their followers should receive appropriate notifications
**Validates: Requirements 9.5**

**Property 48: Social Count Accuracy**
*For any* user profile, the displayed follower and following counts should match the actual relationship counts
**Validates: Requirements 9.6**

**Property 49: Follow Notification Delivery**
*For any* follow action, the followed user should receive a notification about the new follower
**Validates: Requirements 9.7**

#### Profile System Properties

**Property 50: Profile Information Display**
*For any* user profile page, it should display avatar, bio, join date, and founder status (if applicable)
**Validates: Requirements 10.1**

**Property 51: Profile Statistics Accuracy**
*For any* user profile, the displayed entry count, follower count, and referral count should match actual data
**Validates: Requirements 10.2**

**Property 52: User Name Navigation**
*For any* clickable user name, it should navigate to that user's profile page
**Validates: Requirements 10.3**

**Property 53: Profile Entry Display**
*For any* user profile, it should display their recent restaurant entries
**Validates: Requirements 10.4**

**Property 54: Profile Navigation Options**
*For any* profile page, it should provide navigation options to view all of that user's entries
**Validates: Requirements 10.5**

**Property 55: Privacy Settings Enforcement**
*For any* private user profile, only authorized information should be displayed based on the viewer's relationship and privacy settings
**Validates: Requirements 10.7**

## Error Handling

### Comprehensive Error Types
```go
type UserError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Field   string `json:"field,omitempty"`
}

var (
    ErrUserNotFound        = &UserError{"USER_NOT_FOUND", "User not found", ""}
    ErrInvalidCredentials  = &UserError{"INVALID_CREDENTIALS", "Invalid username or password", ""}
    ErrEmailAlreadyExists  = &UserError{"EMAIL_EXISTS", "Email address already registered", "email"}
    ErrInvalidTOTP         = &UserError{"INVALID_TOTP", "Invalid 2FA code", "totp_code"}
    ErrInvalidBackupCode   = &UserError{"INVALID_BACKUP", "Invalid backup code", "backup_code"}
    ErrEmailNotVerified    = &UserError{"EMAIL_NOT_VERIFIED", "Email address not verified", "email"}
    ErrFollowSelf          = &UserError{"FOLLOW_SELF", "Cannot follow yourself", ""}
    ErrAlreadyFollowing    = &UserError{"ALREADY_FOLLOWING", "Already following this user", ""}
    ErrLocationDisabled    = &UserError{"LOCATION_DISABLED", "Location services disabled", ""}
    ErrInvalidToken        = &UserError{"INVALID_TOKEN", "Invalid or expired token", "token"}
)
```

### Security Error Handling
- **Rate Limiting**: Implement rate limiting for sensitive operations
- **Secure Error Messages**: Avoid leaking information in error responses
- **Audit Logging**: Log security-related events for monitoring
- **Input Validation**: Comprehensive validation with sanitization

## Testing Strategy

### Unit Tests
- **Service Layer Tests**: Comprehensive testing of all service methods
- **Model Validation**: Test all validation rules and constraints
- **Authentication Tests**: JWT, 2FA, and password reset functionality
- **Location Calculation**: Test distance calculations and nearby searches
- **Email Service**: Mock SMTP testing for email functionality

### Property-Based Tests
- **Universal Properties**: Test all correctness properties with randomized inputs
- **Security Properties**: Test authentication and authorization across many scenarios
- **Data Integrity**: Test referral tracking and social relationships
- **Location Accuracy**: Test distance calculations with known coordinate pairs
- **Email Functionality**: Test email generation and forwarding with various inputs

### Integration Tests
- **API Endpoint Tests**: Full request/response cycle testing
- **Database Integration**: Test all CRUD operations and relationships
- **Authentication Flow**: Complete login/logout with 2FA testing
- **Social Features**: Follow/unfollow and feed generation testing
- **Email Forwarding**: Test email routing and delivery

### Security Tests
- **2FA Security**: Test TOTP generation and validation
- **Token Security**: Test JWT and confirmation token handling
- **Rate Limiting**: Test abuse prevention mechanisms
- **Input Validation**: Test against injection and XSS attacks

## Performance Considerations

### Database Optimization
- **Query Optimization**: Efficient queries for social feeds and location searches
- **Index Strategy**: Comprehensive indexing for common access patterns
- **Connection Pooling**: Optimal database connection management
- **Caching Strategy**: Cache user profiles and social relationships

### Location Service Performance
- **Spatial Indexing**: Use PostGIS for efficient location queries
- **Distance Calculation**: Optimize haversine calculations
- **Result Caching**: Cache location-based search results
- **Batch Processing**: Process multiple location queries efficiently

### Social Feed Performance
- **Feed Generation**: Efficient algorithms for generating user feeds
- **Pagination**: Implement cursor-based pagination for large datasets
- **Real-time Updates**: Consider WebSocket connections for live updates
- **Content Caching**: Cache frequently accessed user content

## Security Considerations

### Authentication Security
- **2FA Implementation**: Secure TOTP with proper secret management
- **Backup Codes**: Secure generation and storage of recovery codes
- **Session Management**: Secure JWT handling with proper expiration
- **Password Security**: Strong password requirements and secure hashing

### Data Protection
- **Email Security**: Secure email forwarding and SMTP configuration
- **Profile Privacy**: Respect user privacy settings and data access
- **Location Privacy**: Secure handling of GPS coordinates
- **Social Privacy**: Implement proper access controls for social features

### Communication Security
- **Email Encryption**: Use TLS for all email communications
- **API Security**: Implement proper authentication for all endpoints
- **Token Security**: Secure generation and validation of all tokens
- **Rate Limiting**: Prevent abuse of sensitive operations