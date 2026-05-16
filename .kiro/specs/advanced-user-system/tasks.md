# Implementation Plan: Advanced User System

## Overview

This implementation plan breaks down the advanced user system and social profile functionality into discrete coding tasks. The plan follows an incremental approach, building core infrastructure first, then adding features layer by layer, with comprehensive testing throughout.

## Tasks

- [x] 1. Extend user model and database schema
  - Add new fields to User model: EmailVerified, EmailForwarding, Bio, AvatarURL, IsFounder, JoinDate, LastLoginAt, LocationEnabled, PrivacyLevel, TwoFactorEnabled, TwoFactorSecret, ReferredBy
  - Create database migration to add new columns to users table
  - Create new supporting tables: backup_codes, user_follows, profile_changes, email_forwarding_rules, social_notifications, password_resets
  - Add database indexes for performance optimization on social and referral queries
  - _Requirements: 1.4, 2.3, 3.2, 5.3, 6.1, 9.2, 10.1_

- [x] 2. Implement referral system service
  - [x] 2.1 Create ReferralService struct with database integration
    - Implement GenerateReferralCode method with secure random generation
    - Create ProcessReferral method to track referral relationships
    - Add GetReferralStats method for counting successful referrals
    - _Requirements: 1.1, 1.4, 1.5_

  - [x] 2.2 Write property test for referral code generation
    - **Property 1: Referral Message Generation**
    - **Validates: Requirements 1.2, 1.3**

  - [x] 2.3 Implement referral message generation
    - Create GenerateReferralMessage method with template formatting
    - Include referral code and personalized invitation text
    - Add city code and join URL to message template
    - _Requirements: 1.2, 1.3_

  - [x] 2.4 Write property test for referral tracking
    - **Property 2: Referral Tracking Consistency**
    - **Validates: Requirements 1.4**

  - [x] 2.5 Add referral notification system
    - Implement notification delivery for successful referrals
    - Send notifications to both referrer and referee
    - Create referral history tracking
    - _Requirements: 1.6_

  - [x] 2.6 Write property tests for referral statistics and notifications
    - **Property 3: Referral Statistics Accuracy**
    - **Property 4: Referral Notification Delivery**
    - **Validates: Requirements 1.5, 1.6**

- [-] 3. Create email service and forwarding system
  - [x] 3.1 Implement EmailService struct with SMTP configuration
    - Create SMTP client with TLS support
    - Implement SendConfirmationEmail method for profile changes
    - Add SendPasswordResetEmail method with temporary passwords
    - Add SendNotificationEmail method for general notifications
    - _Requirements: 2.1, 2.2, 4.1_

  - [ ] 3.2 Write property tests for email confirmation system
    - **Property 5: Email Change Confirmation**
    - **Property 6: Password Change Notification**
    - **Validates: Requirements 2.1, 2.2**

  - [x] 3.3 Implement email forwarding service
    - Create SetupEmailForwarding method to configure forwarding rules
    - Implement RemoveEmailForwarding method for cleanup
    - Add email forwarding rule management in database
    - Create SMTP relay functionality for username@${CITY_CODE}dinner.club addresses
    - _Requirements: 3.2, 3.3, 3.4, 3.5_

  - [ ] 3.4 Write property tests for email forwarding
    - **Property 12: Email Forwarding Rule Creation**
    - **Property 13: Email Forwarding Functionality**
    - **Property 14: Forwarding Rule Updates**
    - **Property 15: Forwarding Rule Cleanup**
    - **Validates: Requirements 3.2, 3.3, 3.4, 3.5**

  - [x] 3.5 Add email validation and error handling
    - Implement comprehensive email format validation
    - Add email deliverability checking
    - Create error logging and user notification for forwarding failures
    - _Requirements: 3.1, 3.6_

  - [x] 3.6 Write property tests for email validation and error handling
    - **Property 11: Email Validation Consistency**
    - **Property 16: Email Forwarding Error Handling**
    - **Validates: Requirements 3.1, 3.6**

- [x] 4. Implement profile change confirmation system
  - [x] 4.1 Create ProfileService struct for managing profile changes
    - Implement RequestEmailChange method with token generation
    - Create ConfirmEmailChange method with token validation
    - Add UpdateProfile method with confirmation requirements for sensitive changes
    - _Requirements: 2.3, 2.4, 2.5_

  - [x] 4.2 Write property tests for profile change confirmations
    - **Property 7: Sensitive Change Confirmation Requirement**
    - **Property 8: Confirmation Token Security**
    - **Property 9: Valid Token Processing**
    - **Validates: Requirements 2.3, 2.4, 2.5**

  - [x] 4.3 Implement confirmation token management
    - Create secure token generation with expiration
    - Add token validation and expiration checking
    - Implement token cleanup for expired confirmations
    - _Requirements: 2.4, 2.6_

  - [x] 4.4 Write property test for token expiration
    - **Property 10: Expired Token Rejection**
    - **Validates: Requirements 2.6**

- [x] 5. Create password reset system
  - [x] 5.1 Implement password reset service
    - Create RequestPasswordReset method with temporary password generation
    - Implement secure random temporary password generation with expiration
    - Add password reset token management and validation
    - _Requirements: 4.1, 4.2, 4.4_

  - [x] 5.2 Write property tests for password reset functionality
    - **Property 17: Password Reset Email Delivery**
    - **Property 18: Temporary Password Security**
    - **Property 20: Temporary Password Expiration**
    - **Validates: Requirements 4.1, 4.2, 4.4**

  - [x] 5.3 Add password reset security features
    - Implement temporary password force change on first login
    - Add session invalidation on successful password reset
    - Create rate limiting for password reset requests
    - _Requirements: 4.3, 4.5, 4.6_

  - [x] 5.4 Write property tests for password reset security
    - **Property 19: Temporary Password Force Change**
    - **Property 21: Session Invalidation on Password Reset**
    - **Property 22: Password Reset Rate Limiting**
    - **Validates: Requirements 4.3, 4.5, 4.6**

- [x] 6. Implement two-factor authentication system
  - [x] 6.1 Create TOTPService struct for 2FA management
    - Implement GenerateSecret method with QR code generation
    - Add ValidateToken method for TOTP verification
    - Create backup code generation and validation methods
    - _Requirements: 5.1, 5.2, 5.3, 5.5_

  - [x] 6.2 Write property tests for 2FA core functionality
    - **Property 23: 2FA QR Code Generation**
    - **Property 24: 2FA Setup Verification Requirement**
    - **Property 25: Backup Code Generation**
    - **Property 27: Backup Code Single Use**
    - **Validates: Requirements 5.1, 5.2, 5.3, 5.5**

  - [x] 6.3 Enhance authentication flow for 2FA
    - Modify login process to require 2FA code when enabled
    - Add backup code authentication as alternative to TOTP
    - Implement 2FA disable functionality with security verification
    - _Requirements: 5.4, 5.7_

  - [x] 6.4 Write property tests for 2FA authentication flow
    - **Property 26: 2FA Login Requirements**
    - **Property 29: 2FA Disable Security**
    - **Validates: Requirements 5.4, 5.7**

  - [x] 6.5 Add 2FA backup code management
    - Implement backup code exhaustion detection and user prompting
    - Create backup code regeneration functionality
    - Add proper backup code storage and retrieval
    - _Requirements: 5.6_

  - [x] 6.6 Write property test for backup code management
    - **Property 28: Backup Code Exhaustion Handling**
    - **Validates: Requirements 5.6**

- [x] 7. Create founder status system
  - [x] 7.1 Implement founder status management
    - Add founder status assignment functionality
    - Create founder badge display logic across all UI components
    - Implement founder statistics tracking and display
    - _Requirements: 6.1, 6.2, 6.3, 6.6_

  - [x] 7.2 Write property tests for founder status
    - **Property 30: Founder Status Display**
    - **Property 31: Founder Profile Information**
    - **Property 33: Founder Statistics Accuracy**
    - **Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.6**

  - [x] 7.3 Add founder status notifications
    - Implement congratulatory notifications when founder status is granted
    - Create founder status tracking in user profiles
    - Add founder join date display functionality
    - _Requirements: 6.4, 6.5_

  - [x] 7.4 Write property test for founder notifications
    - **Property 32: Founder Status Notification**
    - **Validates: Requirements 6.5**

- [x] 8. Implement avatar and Gravatar system
  - [x] 8.1 Create AvatarService struct for avatar management
    - Implement GetGravatarURL method for email-based avatars
    - Add UploadCustomAvatar method with file validation and processing
    - Create image resizing and optimization functionality
    - _Requirements: 7.1, 7.2, 7.5_

  - [x] 8.2 Write property tests for avatar functionality
    - **Property 34: Gravatar Fallback**
    - **Property 35: Avatar Upload Processing**
    - **Property 38: Avatar Upload Validation**
    - **Validates: Requirements 7.1, 7.2, 7.5**

  - [x] 8.3 Add avatar display and management features
    - Implement consistent avatar display across all UI components
    - Add automatic Gravatar updates when email changes
    - Create error handling with default placeholder images
    - _Requirements: 7.3, 7.4, 7.6_

  - [x] 8.4 Write property tests for avatar display and updates
    - **Property 36: Avatar Display Consistency**
    - **Property 37: Gravatar Email Update**
    - **Property 39: Avatar Loading Error Handling**
    - **Validates: Requirements 7.3, 7.4, 7.6**

- [-] 9. Create location-based services
  - [x] 9.1 Implement LocationService struct for GPS functionality
    - Create CalculateDistance method using Haversine formula
    - Implement FindNearbyRestaurants method with radius filtering
    - Add FindNearestWishlist and FindNearestUnvisited methods
    - _Requirements: 8.2, 8.3, 8.4, 8.6_

  - [x] 9.2 Write property tests for location calculations
    - **Property 40: Distance Calculation Accuracy**
    - **Property 41: Wishlist Distance Calculation**
    - **Property 42: Location Result Sorting**
    - **Property 43: Wishlist Location Search**
    - **Validates: Requirements 8.2, 8.3, 8.4, 8.6**

  - [x] 9.3 Add location-based UI components
    - Create GPS permission request functionality
    - Implement manual location entry options for disabled GPS
    - Add search radius expansion suggestions for empty results
    - _Requirements: 8.1, 8.5, 8.7_

- [-] 10. Implement social following system
  - [x] 10.1 Create SocialService struct for social features
    - Implement FollowUser and UnfollowUser methods with relationship management
    - Add GetFollowers and GetFollowing methods for relationship queries
    - Create GetUserProfile method with social statistics
    - _Requirements: 9.1, 9.2, 9.3, 9.6_

  - [x] 10.2 Write property tests for social relationships
    - **Property 44: Follow Button Display**
    - **Property 45: Follow Relationship Management**
    - **Property 48: Social Count Accuracy**
    - **Validates: Requirements 9.1, 9.2, 9.3, 9.6**

  - [x] 10.3 Implement social feed and notifications
    - Create GetFollowingFeed method for displaying followed users' content
    - Add notification system for new content from followed users
    - Implement follow notification delivery to followed users
    - _Requirements: 9.4, 9.5, 9.7_

  - [x] 10.4 Write property tests for social feeds and notifications
    - **Property 46: Social Feed Content**
    - **Property 47: Follower Notifications**
    - **Property 49: Follow Notification Delivery**
    - **Validates: Requirements 9.4, 9.5, 9.7**

- [x] 11. Create enhanced user profile system
  - [x] 11.1 Implement comprehensive user profiles
    - Create enhanced profile display with avatar, bio, join date, and founder status
    - Add profile statistics display for entries, followers, and referrals
    - Implement user name navigation to profile pages
    - _Requirements: 10.1, 10.2, 10.3_

  - [x] 11.2 Write property tests for profile display
    - **Property 50: Profile Information Display**
    - **Property 51: Profile Statistics Accuracy**
    - **Property 52: User Name Navigation**
    - **Validates: Requirements 10.1, 10.2, 10.3**

  - [x] 11.3 Add profile content and navigation features
    - Implement recent entries display on user profiles
    - Create navigation options to view all user entries
    - Add edit controls for users viewing their own profiles
    - _Requirements: 10.4, 10.5, 10.6_

  - [x] 11.4 Write property tests for profile content and navigation
    - **Property 53: Profile Entry Display**
    - **Property 54: Profile Navigation Options**
    - **Validates: Requirements 10.4, 10.5**

  - [x] 11.5 Implement privacy settings and controls
    - Create privacy level management for user profiles
    - Add privacy enforcement for profile information display
    - Implement access controls based on user relationships and privacy settings
    - _Requirements: 10.7_

  - [x] 11.6 Write property test for privacy settings
    - **Property 55: Privacy Settings Enforcement**
    - **Validates: Requirements 10.7**

- [x] 12. Create API endpoints and controllers
  - [x] 12.1 Implement referral and profile management endpoints
    - Create POST /api/referrals/generate for referral message generation
    - Add GET /api/referrals/stats for referral statistics
    - Implement POST /api/profile/change-email for email change requests
    - Add POST /api/profile/confirm for confirmation token processing
    - _Requirements: 1.2, 1.5, 2.1, 2.5_

  - [x] 12.2 Add authentication and security endpoints
    - Create POST /api/auth/reset-password for password reset requests
    - Implement POST /api/auth/2fa/setup for 2FA configuration
    - Add POST /api/auth/2fa/verify for 2FA verification
    - Create POST /api/auth/2fa/backup-codes for backup code management
    - _Requirements: 4.1, 5.1, 5.2, 5.6_

  - [x] 12.3 Implement social and location endpoints
    - Create POST /api/social/follow/:userID for follow actions
    - Add GET /api/social/feed for following feed
    - Implement GET /api/location/nearby for location-based searches
    - Create GET /api/users/:userID/profile for enhanced profile display
    - _Requirements: 9.2, 9.4, 8.4, 10.1_

- [x] 13. Update templates and UI components
  - [x] 13.1 Create enhanced profile templates
    - Update user profile template with avatar, bio, and social statistics
    - Add founder status display across all user interface components
    - Implement follow/unfollow buttons on profile pages
    - Create referral code display and message generation UI
    - _Requirements: 6.2, 9.1, 10.1, 1.1_

  - [x] 13.2 Add authentication and security UI
    - Create 2FA setup template with QR code display
    - Implement password reset request and confirmation forms
    - Add email change confirmation templates
    - Create backup code display and management interface
    - _Requirements: 5.1, 4.1, 2.1, 5.3_

  - [x] 13.3 Implement location and social UI components
    - Create location permission request interface
    - Add nearby restaurant display with distance information
    - Implement social feed template for followed users' content
    - Create notification display for social interactions
    - _Requirements: 8.1, 8.3, 9.4, 9.5_

- [ ] 14. Add comprehensive error handling and validation
  - [ ] 14.1 Implement service-level error handling
    - Create comprehensive error types for all user operations
    - Add validation for email addresses, referral codes, and user inputs
    - Implement rate limiting for sensitive operations
    - Create audit logging for security-related events
    - _Requirements: All requirements - error handling_

  - [ ] 14.2 Add API-level error responses
    - Implement consistent error response formats
    - Create user-friendly error messages for validation failures
    - Add proper HTTP status codes for all error conditions
    - Implement security-conscious error messages that don't leak information
    - _Requirements: All requirements - API error handling_

- [ ] 15. Write comprehensive unit tests
  - [ ] 15.1 Create service layer unit tests
    - Test all ReferralService methods with various input scenarios
    - Add comprehensive EmailService testing with SMTP mocking
    - Test TOTPService functionality including QR code generation
    - Create LocationService tests with known coordinate calculations
    - _Requirements: All requirements - unit testing coverage_

  - [ ] 15.2 Add model and validation unit tests
    - Test all new model validation rules and constraints
    - Create comprehensive database relationship testing
    - Add avatar upload and processing unit tests
    - Test social relationship management and statistics
    - _Requirements: All requirements - model testing coverage_

- [ ] 16. Write integration tests for new functionality
  - [ ] 16.1 Create API endpoint integration tests
    - Test complete referral generation and tracking flow
    - Add email confirmation and password reset integration tests
    - Test 2FA setup and authentication integration
    - Create social following and feed generation integration tests
    - _Requirements: All requirements - integration testing_

  - [ ] 16.2 Add database and email integration tests
    - Test email forwarding setup and delivery
    - Create location-based search integration tests
    - Add profile privacy and access control integration tests
    - Test notification delivery and social interaction flows
    - _Requirements: All requirements - system integration testing_

- [ ] 17. Update configuration and environment setup
  - [ ] 17.1 Add new configuration options
    - Update .env.sample with SMTP configuration for email services
    - Add 2FA issuer configuration and QR code settings
    - Create avatar upload path and size limit configurations
    - Add location service and distance calculation settings
    - _Requirements: All requirements - configuration management_

  - [ ] 17.2 Update database migrations and deployment
    - Create comprehensive database migration scripts for all new tables
    - Add index creation for performance optimization
    - Update deployment documentation for new email forwarding requirements
    - Create backup and recovery procedures for new user data
    - _Requirements: All requirements - deployment and operations_

- [ ] 18. Final integration and testing
  - [ ] 18.1 Ensure all tests pass and system integration works
    - Run comprehensive test suite and resolve any failures
    - Test complete user registration and onboarding flow with all new features
    - Verify email forwarding and notification delivery in staging environment
    - Test social features and location services end-to-end
    - _Requirements: All requirements - final validation_

  - [ ] 18.2 Performance testing and optimization
    - Test social feed generation performance with large datasets
    - Optimize location-based queries for fast response times
    - Test email service performance and delivery reliability
    - Verify 2FA and authentication performance under load
    - _Requirements: All requirements - performance validation_

## Notes

- All tasks are required for comprehensive implementation
- Each task references specific requirements for traceability
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Integration tests ensure end-to-end functionality works correctly
- The implementation follows a layered approach: models → services → controllers → UI → testing