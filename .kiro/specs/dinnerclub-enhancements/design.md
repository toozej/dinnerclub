# Design Document

## Overview

This design document outlines the comprehensive enhancements to the Dinnerclub Golang web application. The enhancements focus on improving user experience through conditional forms, advanced search capabilities, enhanced restaurant pages, modern authentication, and system-wide consistency improvements.

The design maintains the existing Gin web framework architecture while introducing new components for search functionality, enhanced UI patterns, and modernized authentication middleware.

## Architecture

### Current Architecture
- **Web Framework**: Gin (Go HTTP web framework)
- **Database**: GORM with PostgreSQL
- **Templates**: Go HTML templates
- **Authentication**: Legacy gocondor JWT implementation
- **Static Assets**: Embedded assets package

### Enhanced Architecture Components
- **Search Service**: New search engine with query parsing and filtering
- **Rating System**: Michelin-inspired rating descriptions and validation
- **UI Components**: Conditional form rendering and tooltip system
- **Modern Authentication**: golang-jwt/jwt implementation
- **Template Helpers**: Utility functions for consistent UI rendering

## Components and Interfaces

### 1. Enhanced Entry Management

#### Entry Model Extensions
```go
type Entry struct {
    gorm.Model
    // Existing fields...
    Submitter string
    Name      string
    Location  string
    Cuisine   string
    Visited   bool
    
    // Enhanced fields
    LookingForwardTo string `form:"lookingforwardto" json:"lookingforwardto"` // New field for unvisited restaurants
    
    // Existing optional fields remain unchanged
    Closed         bool
    MealService    string
    Ordered        string
    OrderNext      string
    DoNotOrder     string
    FoodRating     int
    AmbienceRating int
    ValueRating    int
    Comments       string
}
```

#### Conditional Form Rendering
- **New Entry Form**: Conditionally shows fields based on `visited` status
- **Update Entry Form**: Pre-populates all fields with existing data
- **Field Visibility Logic**: JavaScript-driven conditional field display

#### Rating System Enhancement
```go
type RatingDescription struct {
    Value       int    `json:"value"`
    Label       string `json:"label"`
    Description string `json:"description"`
}

var RatingDescriptions = []RatingDescription{
    {1, "Poor", "Disappointing experience, would not recommend"},
    {2, "Fair", "Below average, has significant issues"},
    {3, "Good", "Solid choice, meets expectations"},
    {4, "Very Good", "Excellent quality, highly recommended"},
    {5, "Outstanding", "Exceptional experience, destination worthy"},
}
```

### 2. Search-First Interface

#### Search Service
```go
type SearchService struct {
    db *gorm.DB
}

type SearchQuery struct {
    Text     string
    Author   string
    Wishlist bool
    Visited  *bool
    Location string
    Cuisine  string
}

func (s *SearchService) ParseQuery(query string) SearchQuery
func (s *SearchService) SearchEntries(query SearchQuery) ([]models.Entry, error)
func (s *SearchService) SearchRestaurants(query SearchQuery) ([]models.Restaurant, error)
```

#### UI Layout Changes
- **Fixed Header**: Top 25% of screen with search, navigation, and auth controls
- **Dynamic Content Area**: Bottom 75% showing search results or default feed
- **Search State Management**: JavaScript handling of search/default view transitions

### 3. Enhanced Restaurant Pages

#### Restaurant Aggregation Service
```go
type RestaurantAggregateData struct {
    Restaurant    models.Restaurant
    EntryCount    int
    PopularItems  []PopularItem
    AverageRatings AverageRatings
    RecentEntries []models.Entry
}

type PopularItem struct {
    Name      string
    Frequency int
}

type AverageRatings struct {
    Food     float64
    Ambience float64
    Value    float64
}
```

#### Restaurant Controller Enhancements
- **FindRestaurantByName**: New endpoint for `/restaurants/name/:name`
- **Enhanced Popular Items**: Improved aggregation with frequency counting
- **Entry Sorting**: Most recent entries first
- **Restaurant Listing**: Sort by entry count instead of ID

### 4. Modern Authentication System

#### JWT Implementation Migration
```go
// Replace gocondor JWT with golang-jwt/jwt
import "github.com/golang-jwt/jwt/v5"

type AuthService struct {
    secretKey []byte
    issuer    string
}

type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}

func (a *AuthService) GenerateToken(userID uint, username string) (string, error)
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error)
func (a *AuthService) RefreshToken(tokenString string) (string, error)
```

#### Middleware Updates
- **Remove gocondor dependencies**: Complete removal of legacy authentication
- **Modern JWT validation**: Implement secure token validation
- **Session management**: Improved session handling with proper expiration

### 5. UI Consistency Improvements

#### Template Helper Functions
```go
type TemplateHelpers struct{}

func (h *TemplateHelpers) FormatBoolean(value bool) string
func (h *TemplateHelpers) FormatRating(rating int) string
func (h *TemplateHelpers) UserProfileLink(username string) template.HTML
func (h *TemplateHelpers) TooltipText(fieldName string) string
```

#### CSS and JavaScript Enhancements
- **Consistent Button Styles**: Unified button appearance and behavior
- **Tooltip System**: JavaScript-based tooltips for form fields
- **Responsive Design**: Mobile-friendly layout improvements
- **Loading States**: Better user feedback during operations

## Data Models

### Database Schema Changes

#### New Fields
```sql
-- Add new field to entries table
ALTER TABLE entries ADD COLUMN looking_forward_to TEXT;

-- Add indexes for search performance
CREATE INDEX idx_entries_submitter ON entries(submitter);
CREATE INDEX idx_entries_name_visited ON entries(name, visited);
CREATE INDEX idx_restaurants_name ON restaurants(name);
```

#### Search Optimization
- **Full-text search indexes**: For restaurant names and cuisine types
- **Composite indexes**: For common query patterns
- **Performance considerations**: Query optimization for search operations

### Data Validation

#### Enhanced Validation Rules
```go
type EntryValidation struct {
    Name     string `binding:"required,min=1,max=100"`
    Location string `binding:"required,oneof=N NE NW SW SE S"`
    Cuisine  string `binding:"required,min=1,max=50"`
    
    // Conditional validation based on visited status
    LookingForwardTo string `binding:"required_if=Visited false"`
    FoodRating       int    `binding:"required_if=Visited true,min=1,max=5"`
}
```

## Error Handling

### Consistent Error Responses
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Code    int    `json:"code"`
    Details string `json:"details,omitempty"`
}

func HandleValidationError(c *gin.Context, err error)
func HandleDatabaseError(c *gin.Context, err error)
func HandleAuthenticationError(c *gin.Context, err error)
```

### User-Friendly Error Messages
- **Form validation errors**: Clear, actionable error messages
- **Search errors**: Helpful suggestions for query refinement
- **Authentication errors**: Secure error messages that don't leak information

## Testing Strategy

### Unit Tests
- **Model validation tests**: Comprehensive validation rule testing
- **Service layer tests**: Search functionality and data aggregation
- **Authentication tests**: JWT token generation and validation
- **Template helper tests**: UI utility function testing

### Integration Tests
- **API endpoint tests**: Full request/response cycle testing
- **Database integration**: CRUD operations and search queries
- **Authentication flow**: Login/logout and session management
- **Form submission**: End-to-end form processing

### UI Tests
- **Conditional form rendering**: JavaScript behavior testing
- **Search interface**: Search functionality and result display
- **Responsive design**: Cross-device compatibility testing
- **Accessibility**: Screen reader and keyboard navigation testing

## Implementation Phases

### Phase 1: Foundation (Entry Forms & UI)
1. Add conditional field logic to entry forms
2. Implement tooltip system
3. Update templates with consistent styling
4. Add field pre-population for update forms

### Phase 2: Search Infrastructure
1. Implement search service and query parsing
2. Create search UI components
3. Add search endpoints to controllers
4. Integrate search with existing templates

### Phase 3: Restaurant Enhancements
1. Create restaurant aggregation service
2. Implement restaurant-by-name endpoints
3. Enhance popular items calculation
4. Update restaurant templates with new layout

### Phase 4: Authentication Modernization
1. Replace gocondor JWT with golang-jwt/jwt
2. Update authentication middleware
3. Implement secure session management
4. Update authentication templates

### Phase 5: Consistency & Polish
1. Standardize return types across controllers
2. Implement consistent validation patterns
3. Update all templates with helper functions
4. Add comprehensive error handling

## Security Considerations

### Authentication Security
- **JWT Secret Management**: Secure secret key storage and rotation
- **Token Expiration**: Appropriate token lifetime and refresh logic
- **Session Security**: Secure cookie settings and CSRF protection

### Input Validation
- **SQL Injection Prevention**: Parameterized queries and input sanitization
- **XSS Protection**: Template escaping and content security policy
- **Data Validation**: Server-side validation for all user inputs

### Search Security
- **Query Sanitization**: Prevent malicious search queries
- **Rate Limiting**: Prevent search abuse and DoS attacks
- **Access Control**: Ensure users can only access authorized data

## Performance Considerations

### Database Optimization
- **Query Optimization**: Efficient queries for search and aggregation
- **Index Strategy**: Appropriate indexes for common access patterns
- **Connection Pooling**: Optimal database connection management

### Caching Strategy
- **Template Caching**: Cache compiled templates for better performance
- **Search Result Caching**: Cache frequent search queries
- **Static Asset Caching**: Optimize CSS/JS delivery

### Frontend Performance
- **JavaScript Optimization**: Minimize and optimize client-side code
- **CSS Optimization**: Efficient styling and responsive design
- **Image Optimization**: Optimize any images or icons used