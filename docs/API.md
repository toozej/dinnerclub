# API Documentation

This document describes the dinnerclub REST API endpoints, authentication, and usage examples.

## Base URL

- Development: `http://localhost:8080`
- Production: `https://your-domain.com`

## Authentication

The dinnerclub API uses JWT (JSON Web Tokens) for authentication. Some endpoints are public, while others require authentication.

### Authentication Flow

#### 1. Register a New User

```http
POST /auth/register
Content-Type: application/x-www-form-urlencoded

username=johndoe&password=securepassword&referralcode=your_referral_code
```

**Response:**
```json
{
  "message": "User registered successfully",
  "user_id": 123
}
```

#### 2. Login

```http
POST /auth/login
Content-Type: application/x-www-form-urlencoded

username=johndoe&password=securepassword
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-01T12:00:00Z",
  "user": {
    "id": 123,
    "username": "johndoe"
  }
}
```

#### 3. Using Authentication Tokens

Include the JWT token in the Authorization header for authenticated requests:

```http
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### 4. Refresh Token

```http
POST /auth/refresh
Content-Type: application/json
Authorization: Bearer <refresh_token>
```

**Response:**
```json
{
  "token": "new_access_token",
  "expires_at": "2024-01-01T13:00:00Z"
}
```

#### 5. Logout

```http
POST /auth/logout
Authorization: Bearer <access_token>
```

## API Endpoints

### Public Endpoints (No Authentication Required)

#### Get All Entries

```http
GET /entries?page=1&limit=20
```

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20, max: 100)

**Response:**
```json
{
  "entries": [
    {
      "id": 1,
      "name": "Tasty Restaurant",
      "location": "NE",
      "cuisine": "Italian",
      "submitter": "johndoe",
      "visited": true,
      "food_rating": 4,
      "ambience_rating": 3,
      "value_rating": 4,
      "ordered": "Pizza Margherita, Tiramisu",
      "order_next": "Pasta Carbonara",
      "comments": "Great pizza!",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### Get Single Entry

```http
GET /entries/{id}
```

**Response:**
```json
{
  "id": 1,
  "name": "Tasty Restaurant",
  "location": "NE",
  "cuisine": "Italian",
  "submitter": "johndoe",
  "visited": true,
  "food_rating": 4,
  "ambience_rating": 3,
  "value_rating": 4,
  "ordered": "Pizza Margherita, Tiramisu",
  "order_next": "Pasta Carbonara",
  "do_not_order": "",
  "meal_service": "Dinner",
  "closed": false,
  "looking_forward_to": "",
  "comments": "Great pizza!",
  "created_at": "2024-01-01T12:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z"
}
```

#### Get All Restaurants

```http
GET /restaurants?page=1&limit=20
```

**Response:**
```json
{
  "restaurants": [
    {
      "id": 1,
      "name": "Tasty Restaurant",
      "location": "NE",
      "cuisine": "Italian",
      "entry_count": 5,
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

#### Get Restaurant by Name

```http
GET /restaurants/name/{restaurant_name}
```

**Response:**
```json
{
  "restaurant": {
    "id": 1,
    "name": "Tasty Restaurant",
    "location": "NE",
    "cuisine": "Italian",
    "entry_count": 5
  },
  "popular_items": [
    {
      "name": "Pizza Margherita",
      "frequency": 8
    },
    {
      "name": "Tiramisu",
      "frequency": 5
    }
  ],
  "average_ratings": {
    "food": 4.2,
    "ambience": 3.8,
    "value": 4.0
  },
  "recent_entries": [
    {
      "id": 15,
      "submitter": "johndoe",
      "visited": true,
      "food_rating": 4,
      "comments": "Excellent as always",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

#### Search

```http
GET /search?q=pizza&page=1&limit=20
GET /search/entries?q=italian&author=johndoe&visited=true
GET /search/restaurants?q=northeast&cuisine=italian
```

**Query Parameters:**
- `q`: Search query
- `page`: Page number
- `limit`: Items per page
- `author`: Filter by username (entries only)
- `visited`: Filter by visited status (entries only)
- `cuisine`: Filter by cuisine type
- `location`: Filter by location

**Response:**
```json
{
  "results": {
    "entries": [...],
    "restaurants": [...]
  },
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 25,
    "total_pages": 2
  }
}
```

#### Get Search Suggestions

```http
GET /search/suggestions?q=piz
```

**Response:**
```json
{
  "suggestions": [
    "pizza",
    "pizza margherita",
    "pizzeria"
  ]
}
```

#### Auto-complete Endpoints

```http
GET /api/suggestions/cuisine?q=ital
GET /api/suggestions/restaurant-names?q=tasty
GET /api/suggestions/meal-service?q=din
GET /api/suggestions/dishes?q=pizza
GET /api/suggestions/locations?q=ne
```

**Response:**
```json
{
  "suggestions": ["Italian", "Italian-American"]
}
```

#### Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Authenticated Endpoints (Require JWT Token)

#### Create New Entry

```http
POST /entries/new
Authorization: Bearer <token>
Content-Type: application/x-www-form-urlencoded

name=New Restaurant&location=SE&cuisine=Mexican&visited=true&food_rating=4&ambience_rating=3&value_rating=4&ordered=Tacos&comments=Delicious!
```

**Request Body Parameters:**
- `name` (required): Restaurant name
- `location` (required): Location code (N, NE, NW, SW, SE, S)
- `cuisine` (required): Cuisine type
- `visited` (required): Boolean (true/false)
- `looking_forward_to` (required if visited=false): What you're looking forward to
- `food_rating` (required if visited=true): Rating 1-5
- `ambience_rating` (optional): Rating 1-5
- `value_rating` (optional): Rating 1-5
- `meal_service` (optional): Breakfast, Lunch, Dinner, etc.
- `ordered` (optional): Comma-separated list of dishes
- `order_next` (optional): Comma-separated list of dishes to try next
- `do_not_order` (optional): Comma-separated list of dishes to avoid
- `comments` (optional): Additional comments

**Response:**
```json
{
  "message": "Entry created successfully",
  "entry_id": 123
}
```

#### Update Entry

```http
PUT /entries/{id}
Authorization: Bearer <token>
Content-Type: application/x-www-form-urlencoded

name=Updated Restaurant&cuisine=Italian-American&comments=Updated review
```

**Response:**
```json
{
  "message": "Entry updated successfully"
}
```

#### Delete Entry

```http
POST /entries/{id}/delete
Authorization: Bearer <token>
```

**Response:**
```json
{
  "message": "Entry deleted successfully"
}
```

#### Get User's Entries

```http
GET /entries/submittedby/{username}?page=1&limit=20
Authorization: Bearer <token>
```

#### Get User Profile

```http
GET /profile
Authorization: Bearer <token>
```

**Response:**
```json
{
  "user": {
    "id": 123,
    "username": "johndoe",
    "created_at": "2024-01-01T12:00:00Z"
  },
  "stats": {
    "total_entries": 25,
    "restaurants_visited": 20,
    "restaurants_wishlist": 5
  }
}
```

#### Update Profile

```http
POST /profile/update
Authorization: Bearer <token>
Content-Type: application/x-www-form-urlencoded

current_password=oldpass&new_password=newpass&confirm_password=newpass
```

#### Delete Profile

```http
POST /profile/delete
Authorization: Bearer <token>
Content-Type: application/x-www-form-urlencoded

password=currentpassword&confirm_delete=true
```

## Response Formats

### Success Responses

All successful responses return HTTP 200 (or 201 for creation) with JSON data.

### Error Responses

Error responses include appropriate HTTP status codes and error details:

```json
{
  "error": "Validation failed",
  "code": 400,
  "details": "Restaurant name is required"
}
```

### Common HTTP Status Codes

- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required or invalid token
- `403 Forbidden`: Access denied
- `404 Not Found`: Resource not found
- `422 Unprocessable Entity`: Validation errors
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **General endpoints**: 100 requests per minute per IP
- **Authentication endpoints**: 10 requests per minute per IP

When rate limits are exceeded, the API returns HTTP 429 with:

```json
{
  "error": "Rate limit exceeded",
  "code": 429,
  "retry_after": 60
}
```

## CSRF Protection

For web form submissions, CSRF tokens are required. The token is available in the HTML forms or can be retrieved from the CSRF cookie.

## Content Types

### Request Content Types

- **Form submissions**: `application/x-www-form-urlencoded`
- **JSON requests**: `application/json` (for some endpoints)

### Response Content Types

- **API responses**: `application/json`
- **Web pages**: `text/html`

## Example Usage

### JavaScript/Fetch Example

```javascript
// Login
const loginResponse = await fetch('/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded',
  },
  body: 'username=johndoe&password=securepassword'
});

const loginData = await loginResponse.json();
const token = loginData.token;

// Create entry
const entryResponse = await fetch('/entries/new', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/x-www-form-urlencoded',
  },
  body: 'name=Great Restaurant&location=NE&cuisine=Italian&visited=true&food_rating=5'
});

// Search
const searchResponse = await fetch('/search?q=italian&format=json');
const searchData = await searchResponse.json();
```

### cURL Examples

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -d "username=johndoe&password=securepass&referralcode=ref123"

# Login
curl -X POST http://localhost:8080/auth/login \
  -d "username=johndoe&password=securepass"

# Get entries (with token)
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/entries

# Create entry
curl -X POST http://localhost:8080/entries/new \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d "name=New Place&location=SE&cuisine=Thai&visited=true&food_rating=4"

# Search
curl "http://localhost:8080/search?q=thai&format=json"
```

## Pagination

Most list endpoints support pagination:

- `page`: Page number (starts at 1)
- `limit`: Items per page (default: 20, max: 100)

Pagination information is included in responses:

```json
{
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8,
    "has_next": true,
    "has_prev": false
  }
}
```

## Filtering and Sorting

### Search Filters

- `author`: Filter by username
- `visited`: Filter by visited status (true/false)
- `cuisine`: Filter by cuisine type
- `location`: Filter by location code
- `rating`: Filter by minimum rating

### Special Search Queries

- `wishlist` or `want to go`: Show unvisited restaurants
- `author:username`: Filter by specific user
- `cuisine:italian`: Filter by cuisine
- `location:ne`: Filter by location
- `rating:4+`: Filter by minimum rating

## WebSocket Support

Currently, the API does not support WebSocket connections. All communication is via HTTP REST endpoints.

## API Versioning

The current API is version 1 and does not use explicit versioning in URLs. Future versions may introduce versioning if breaking changes are needed.

## SDK and Client Libraries

Currently, no official SDK or client libraries are provided. The API follows standard REST conventions and can be used with any HTTP client library.