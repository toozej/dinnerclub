# Implementation Plan

- [x] 1. Set up enhanced entry model and database migrations
  - Add `LookingForwardTo` field to Entry model with proper GORM tags
  - Create database migration to add new column to entries table
  - Add database indexes for search performance optimization
  - _Requirements: 1.2, 1.3_

- [x] 2. Implement rating system with descriptions
  - Create RatingDescription struct and rating constants
  - Implement rating validation functions with 1-5 scale
  - Create template helper functions for rating display with descriptions
  - Write unit tests for rating validation and display functions
  - _Requirements: 5.1, 5.2, 5.3_

- [x] 3. Create conditional entry form logic
  - Implement JavaScript for conditional field visibility based on visited status
  - Update new entry template with conditional field rendering
  - Add "looking forward to" field for unvisited restaurants
  - Remove "closed" field from new entry form template
  - _Requirements: 1.1, 1.2, 1.4_

- [x] 4. Implement tooltip system for form fields
  - Create tooltip helper functions with field descriptions
  - Add JavaScript tooltip functionality to form templates
  - Implement tooltips for comma-separated fields (ordered, order next, etc.)
  - Write CSS styles for consistent tooltip appearance
  - _Requirements: 1.7_

- [x] 5. Enhance update entry form with pre-population
  - Modify update entry template to pre-populate all form fields
  - Ensure "closed" field remains available in update form
  - Implement proper form value binding for all field types
  - Add form validation to prevent data loss during updates
  - _Requirements: 1.6, 2.6_

- [x] 6. Implement boolean display formatting
  - Create template helper function to convert true/false to yes/no
  - Update all templates to use yes/no display for boolean values
  - Apply consistent boolean formatting across entry and restaurant views
  - Write unit tests for boolean formatting functions
  - _Requirements: 1.5_

- [x] 7. Create search service and query parsing
  - Implement SearchService struct with database integration
  - Create SearchQuery struct for complex query handling
  - Implement query parsing for "wishlist", author, and other filters
  - Add search methods for entries and restaurants with proper indexing
  - _Requirements: 4.4, 4.5_

- [x] 8. Build search-first UI layout
  - Create fixed header template with search box and navigation
  - Implement responsive layout with 25% header and 75% content areas
  - Add JavaScript for search state management and result display
  - Create search results templates for entries and restaurants
  - _Requirements: 4.1, 4.2, 4.3, 4.6_

- [x] 9. Implement search endpoints and controllers
  - Add search endpoints to handle GET requests with query parameters
  - Create search controller methods for entries and restaurants
  - Implement search result pagination and sorting
  - Add proper error handling for invalid search queries
  - _Requirements: 4.4, 4.5_

- [x] 10. Reorganize entry view display
  - Update entry template to reorganize fields: cuisine, location, ordered items, order next
  - Implement proper field grouping and visual hierarchy
  - Make submitter name a clickable link to user's entries
  - Add proper styling for improved readability
  - _Requirements: 2.1, 2.4, 2.5_

- [x] 11. Create restaurant aggregation service
  - Implement RestaurantAggregateData struct for comprehensive restaurant data
  - Create service methods for calculating popular items with frequency
  - Implement average ratings calculation across all restaurant entries
  - Add methods for retrieving recent entries sorted by date
  - _Requirements: 3.3, 3.4, 3.6_

- [x] 12. Implement restaurant-by-name endpoint
  - Create new controller method for `/restaurants/name/:name` route
  - Implement restaurant lookup by name with proper error handling
  - Add route registration in router configuration
  - Create comprehensive restaurant page template with three sections
  - _Requirements: 3.1, 3.2_

- [x] 13. Enhance restaurant page layout
  - Update restaurant template with three-section layout (static, popular, entries)
  - Display static restaurant information in top section
  - Show popular items and averaged ratings in middle section
  - List recent entries with links to full entry pages in bottom section
  - _Requirements: 3.2, 3.3, 3.4_

- [x] 14. Update restaurant listing with entry count sorting
  - Modify FindRestaurants controller to sort by entry count instead of ID
  - Implement entry count calculation for each restaurant
  - Update restaurant index template to display entry counts
  - Add proper pagination for restaurant listings
  - _Requirements: 3.5_

- [x] 15. Replace gocondor JWT with golang-jwt/jwt
  - Remove all gocondor JWT imports and dependencies
  - Implement new AuthService using golang-jwt/jwt library
  - Create Claims struct with proper JWT claims structure
  - Implement secure token generation with appropriate expiration
  - _Requirements: 6.1, 6.2_

- [x] 16. Implement modern JWT validation middleware
  - Create new authentication middleware using golang-jwt/jwt
  - Implement secure token validation with proper error handling
  - Add token refresh functionality for session management
  - Update all protected routes to use new middleware
  - _Requirements: 6.3, 6.4_

- [x] 17. Update authentication controllers
  - Modify LoginPost to use new JWT implementation
  - Update authentication error handling with consistent responses
  - Remove all gocondor authentication dependencies
  - Implement proper session cleanup on logout
  - _Requirements: 6.4, 6.5_

- [x] 18. Standardize controller return types
  - Review all controller methods for consistent return type patterns
  - Implement consistent error response structures
  - Update all controllers to use standardized response formats
  - Add proper HTTP status code handling across all endpoints
  - _Requirements: 7.1_

- [x] 19. Implement consistent input validation
  - Create validation helper functions for common patterns
  - Update all controllers to use consistent validation approaches
  - Implement proper validation error handling and user feedback
  - Add validation for conditional fields based on entry state
  - _Requirements: 7.2_

- [x] 20. Standardize controller function naming
  - Review and update controller method names for consistency
  - Ensure consistent HTTP method usage across CRUD operations
  - Update route definitions to match standardized naming
  - Document naming conventions for future development
  - _Requirements: 7.3, 7.6_

- [x] 21. Create consistent UI button styles
  - Implement unified CSS classes for all button types
  - Update all templates to use consistent button styling
  - Ensure menu buttons and action buttons have consistent appearance
  - Add hover and active states for better user experience
  - _Requirements: 7.4_

- [x] 22. Implement template helper functions
  - Create comprehensive template helper package
  - Implement helpers for boolean formatting, rating display, and user links
  - Add tooltip text generation functions
  - Register all helper functions with template engine
  - _Requirements: 1.5, 1.7, 2.5_

- [x] 23. Add comprehensive error handling
  - Implement consistent error response structures
  - Add user-friendly error messages for all validation failures
  - Create proper error logging for debugging and monitoring
  - Add graceful error handling for database and authentication errors
  - _Requirements: 7.1, 7.2_

- [x] 24. Write comprehensive unit tests
  - Create unit tests for all new service layer functions
  - Test search functionality with various query combinations
  - Add tests for rating system validation and display
  - Test authentication JWT generation and validation
  - _Requirements: All requirements - testing coverage_

- [x] 25. Write integration tests for new endpoints
  - Test search endpoints with complex queries
  - Test restaurant-by-name endpoint functionality
  - Test conditional form submission and validation
  - Test authentication flow with new JWT implementation
  - _Requirements: All requirements - integration testing_

- [x] 26. Update CSS and JavaScript for enhanced UI
  - Implement responsive design improvements for mobile devices
  - Add JavaScript for conditional form field management
  - Create smooth transitions for search interface state changes
  - Optimize CSS for consistent styling across all components
  - _Requirements: 1.1, 1.7, 4.1, 4.2, 4.3_

- [x] 27. Implement CSRF protection using Gin and filippo.io/csrf/gorilla
  - Implement CSRF protection using filippo.io/csrf/gorilla package
  - Implement any TODO comments related to CSRF protection, particularly in HTML templates and Go code
  - Ensure dinnerclub application can be built with `make local-build` after changes
  - Ensure all tests pass using `go test ./...` and resolve any issues found, particularly those related to addition of CSRF protection 
  - Ensure all pre-commit checks pass using `make pre-commit-run` and resolve any issues found, particularly those related to addition of CSRF protection 
  - Update documentation, and the config package + .env.sample configuration file with any CSRF changes
  - _Requirements: 6_

- [x] 28. Implement rate limiting using Gin and Gin-compatible package
  - Implement rate limiting, using github.com/JGLTechnologies/gin-rate-limit package or another Gin-compatible package
  - Set default rate limits and overrides using env variables, with related config package updates
  - Ensure dinnerclub application can be built with `make local-build` after changes
  - Ensure all tests pass using `go test ./...` and resolve any issues found after changes
  - Ensure all pre-commit checks pass using `make pre-commit-run` and resolve any issues found after changes
  - Update documentation with rate limiting changes

- [x] 29. Add go doc standard comments
  - Follow the standards as described in https://go.dev/doc/comment such that they appear properly in https://pkg.go.dev/
  - Add go doc comments for each exported function 
  - Add go doc comments for each type 
  - Add go doc comments for each package with their purpose

- [x] 30. Documentation
  - Write documentation on how to run dinnerclub in development mode using Make in DEVELOPMENT.md
  - Write API documentation including how to authenticate with and interact with the API in API.md
  - Write documentation on what the various configuration options for dinnerclub are, and detailed example configurations in CONFIGURATION.md
  - Write documentation on to run dinnerclub in production mode using Make, Docker, and Docker Compose in DEPLOYMENT.md
  - create `docs/` directory with DEVELOPMENT.md, CONFIGURATION.md, DEPLOYMENT.md, API.md documentation files within `docs/` directory