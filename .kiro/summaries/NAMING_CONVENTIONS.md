# Controller Function Naming Conventions

## Overview
This document defines the standardized naming conventions for controller functions in the dinnerclub application.

## Naming Convention Rules

### 1. HTTP Method Prefixes
All controller functions should be prefixed with the HTTP method they handle:
- `Get` - for GET requests
- `Post` - for POST requests  
- `Put` - for PUT requests
- `Patch` - for PATCH requests
- `Delete` - for DELETE requests

### 2. Resource-Based Naming
Function names should follow the pattern: `{HTTPMethod}{Resource}{Action?}`

Examples:
- `GetEntries` - GET /entries (list all entries)
- `GetEntry` - GET /entries/:id (get single entry)
- `PostEntry` - POST /entries (create new entry)
- `PutEntry` - PUT /entries/:id (update entry)
- `DeleteEntry` - DELETE /entries/:id (delete entry)

### 3. Specific Action Naming
For specific actions beyond basic CRUD:
- `GetEntryForm` - GET /entries/new (show create form)
- `GetEntryUpdateForm` - GET /entries/:id/update (show update form)
- `GetEntriesByUsername` - GET /entries/submittedby/:username
- `GetRestaurantByName` - GET /restaurants/name/:name

### 4. Authentication and Special Cases
- `GetLogin` - GET /auth/login
- `PostLogin` - POST /auth/login
- `PostLogout` - POST /auth/logout
- `GetRegister` - GET /auth/register
- `PostRegister` - POST /auth/register
- `PostRefreshToken` - POST /auth/refresh

### 5. Search Operations
- `GetSearch` - GET /search
- `GetSearchEntries` - GET /search/entries
- `GetSearchRestaurants` - GET /search/restaurants
- `GetSearchSuggestions` - GET /search/suggestions

### 6. Profile Operations
- `GetProfile` - GET /profile
- `PostProfileUpdate` - POST /profile/update
- `PostProfileDelete` - POST /profile/delete

### 7. Status and Health
- `GetHealth` - GET /health
- `GetStatus` - GET /status

## Route Mapping Standards

### RESTful Route Patterns
| HTTP Method | Route Pattern | Controller Function | Description |
|-------------|---------------|-------------------|-------------|
| GET | `/resources` | `GetResources` | List all resources |
| GET | `/resources/:id` | `GetResource` | Get single resource |
| GET | `/resources/new` | `GetResourceForm` | Show create form |
| POST | `/resources` | `PostResource` | Create new resource |
| GET | `/resources/:id/edit` | `GetResourceUpdateForm` | Show update form |
| PUT/PATCH | `/resources/:id` | `PutResource` | Update resource |
| DELETE | `/resources/:id` | `DeleteResource` | Delete resource |

### Non-RESTful Patterns
For actions that don't fit standard CRUD:
- Use descriptive action names: `GetEntriesByUsername`
- Include the action in the function name: `PostRefreshToken`
- Maintain consistency with HTTP method prefix

## Implementation Guidelines

1. **Consistency**: All functions must follow the naming convention
2. **Clarity**: Function names should clearly indicate their purpose
3. **HTTP Method Alignment**: Function names must match the HTTP method they handle
4. **Resource Focus**: Names should be resource-centric, not action-centric
5. **Avoid Abbreviations**: Use full words for clarity

## Migration Strategy

1. Update function names to follow conventions
2. Update route registrations to match new function names
3. Ensure all imports and references are updated
4. Test all endpoints to verify functionality
5. Update documentation and comments

## Examples of Changes

### Before (Inconsistent)
```go
func FindEntries(c *gin.Context)        // GET /entries
func CreateEntryGet(c *gin.Context)     // GET /entries/new  
func CreateEntryPost(c *gin.Context)    // POST /entries
func UpdateEntryGet(c *gin.Context)     // GET /entries/:id/update
func UpdateEntryPost(c *gin.Context)    // POST /entries/:id/update
func LoginGet(c *gin.Context)           // GET /auth/login
func LoginPost(c *gin.Context)          // POST /auth/login
```

### After (Standardized)
```go
func GetEntries(c *gin.Context)         // GET /entries
func GetEntryForm(c *gin.Context)       // GET /entries/new
func PostEntry(c *gin.Context)          // POST /entries  
func GetEntryUpdateForm(c *gin.Context) // GET /entries/:id/update
func PutEntry(c *gin.Context)           // PUT /entries/:id
func GetLogin(c *gin.Context)           // GET /auth/login
func PostLogin(c *gin.Context)          // POST /auth/login
```

## Benefits

1. **Predictability**: Developers can predict function names from routes
2. **Consistency**: All controllers follow the same patterns
3. **Maintainability**: Easier to understand and modify code
4. **Documentation**: Self-documenting function names
5. **Tooling**: Better IDE support and code navigation