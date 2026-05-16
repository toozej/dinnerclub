# Requirements Document

## Introduction

This feature encompasses comprehensive enhancements to the Dinnerclub Golang web application, focusing on improving user experience, search functionality, authentication security, and overall consistency. The enhancements include UI improvements for entry forms, advanced search capabilities, restaurant page reorganization, authentication middleware refactoring, and system-wide consistency improvements.

## Requirements

### Requirement 1: Entry Form Improvements

**User Story:** As a user creating or editing restaurant entries, I want an improved form experience with conditional fields and better organization, so that I can efficiently capture relevant information based on whether I've visited the restaurant.

#### Acceptance Criteria

1. WHEN a user accesses the new entry form THEN the system SHALL display conditional fields based on visit status
2. WHEN a user indicates they haven't visited a restaurant THEN the system SHALL show "what are you looking forward to" field
3. WHEN a user indicates they have visited a restaurant THEN the system SHALL show current optional fields (meal service, ratings, etc.)
4. WHEN a user creates a new entry THEN the system SHALL NOT display the "closed" field
5. WHEN a user edits an existing entry THEN the system SHALL display the "closed" field
6. WHEN displaying boolean values in the UI THEN the system SHALL show "yes/no" instead of "true/false"
7. WHEN a user hovers over form fields THEN the system SHALL display tooltips explaining field usage, especially for comma-separated lists

### Requirement 2: Entry and Restaurant View Reorganization

**User Story:** As a user viewing entries and restaurants, I want better organized information display with logical field grouping, so that I can quickly find relevant information.

#### Acceptance Criteria

1. WHEN viewing an entry THEN the system SHALL organize fields as: cuisine, location, ordered items, order next time
2. WHEN viewing a restaurant page THEN the system SHALL organize static info as: cuisine, location, etc.
3. WHEN displaying restaurant entries THEN the system SHALL show most recent entry first
4. WHEN displaying entry details THEN the system SHALL show ordered, order next, don't order information prominently
5. WHEN displaying a submitter name THEN the system SHALL make it a hyperlink to their submitted entries
6. WHEN a user accesses the update entry page THEN the system SHALL pre-populate all fields with existing data

### Requirement 3: Enhanced Restaurant Pages

**User Story:** As a user browsing restaurants, I want comprehensive restaurant pages that aggregate all information about a specific restaurant, so that I can make informed dining decisions.

#### Acceptance Criteria

1. WHEN accessing /restaurants/$restaurantName THEN the system SHALL display a dedicated restaurant page
2. WHEN viewing a restaurant page THEN the system SHALL show static restaurant information in the top section
3. WHEN viewing a restaurant page THEN the system SHALL display popular ordered dishes and averaged ratings in the middle section
4. WHEN viewing a restaurant page THEN the system SHALL list all entries for that restaurant in the bottom section
5. WHEN accessing /restaurants THEN the system SHALL list all restaurants sorted by number of entries (ascending)
6. WHEN calculating popular dishes THEN the system SHALL aggregate data from the "ordered" field across all entries

### Requirement 4: Search-First Interface

**User Story:** As a user navigating the application, I want a prominent search interface that allows complex queries, so that I can quickly find specific restaurants or entries.

#### Acceptance Criteria

1. WHEN accessing any page THEN the system SHALL display a search box in a frozen top 1/4 of the screen
2. WHEN the top section is displayed THEN the system SHALL include login/logout and home buttons alongside search
3. WHEN a user performs a search THEN the system SHALL replace the bottom 3/4 content with search results
4. WHEN searching THEN the system SHALL support complex queries like "wishlist" or "want to go to"
5. WHEN searching THEN the system SHALL support filtering by author (username)
6. WHEN no search is active THEN the system SHALL display the default feed of entries in the bottom section

### Requirement 5: Rating System Enhancement

**User Story:** As a user rating restaurants, I want a clear rating system with descriptions, so that I can provide consistent and meaningful ratings.

#### Acceptance Criteria

1. WHEN viewing rating options THEN the system SHALL follow a Michelin-inspired star system
2. WHEN displaying rating fields THEN the system SHALL provide descriptions for each rating number
3. WHEN a user selects a rating THEN the system SHALL clearly indicate what that rating means
4. WHEN displaying ratings THEN the system SHALL maintain consistency across food, ambience, and value ratings

### Requirement 6: Authentication Middleware Refactoring

**User Story:** As a system administrator, I want secure and modern authentication middleware, so that user sessions are properly managed with current security standards.

#### Acceptance Criteria

1. WHEN implementing authentication THEN the system SHALL use golang-jwt/jwt library instead of legacy gocondor JWT
2. WHEN a user authenticates THEN the system SHALL generate secure JWT tokens
3. WHEN validating tokens THEN the system SHALL use modern JWT validation practices
4. WHEN handling authentication errors THEN the system SHALL provide appropriate error responses
5. WHEN implementing middleware THEN the system SHALL remove all gocondor dependencies

### Requirement 7: System Consistency Improvements

**User Story:** As a developer maintaining the application, I want consistent patterns across the codebase, so that the application is maintainable and reliable.

#### Acceptance Criteria

1. WHEN functions return values THEN the system SHALL use consistent return types across all controllers
2. WHEN validating input THEN the system SHALL apply consistent validation patterns
3. WHEN naming controller functions THEN the system SHALL follow consistent naming conventions
4. WHEN displaying UI elements THEN the system SHALL use consistent button styles and types
5. WHEN handling nil strings THEN the system SHALL follow Go best practices for nil string handling
6. WHEN implementing CRUD operations THEN the system SHALL maintain consistent HTTP method usage