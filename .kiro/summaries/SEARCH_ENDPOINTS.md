# Search Endpoints Documentation

This document describes the search endpoints implemented for the dinnerclub application.

## Endpoints

### 1. General Search: `GET /search`

Performs a comprehensive search across both entries and restaurants.

**Query Parameters:**
- `q` (required): Search query string
- `page` (optional): Page number for pagination (default: 1)
- `limit` (optional): Number of results per page (default: 20, max: 100)
- `format` (optional): Response format ("json" for JSON response)

**Example:**
```
GET /search?q=Italian&page=1&limit=10
GET /search?q=author:testuser&format=json
GET /search?q=wishlist&page=2
```

**Response:** HTML template or JSON with entries, restaurants, pagination info

### 2. Entry Search: `GET /search/entries`

Searches specifically for entries with enhanced filtering and sorting.

**Query Parameters:**
- `q` (required): Search query string
- `page` (optional): Page number for pagination (default: 1)
- `limit` (optional): Number of results per page (default: 20, max: 100)
- `sort` (optional): Sort field (created_at, name, submitter, location, cuisine) (default: created_at)
- `order` (optional): Sort order (asc, desc) (default: desc)
- `format` (optional): Response format ("json" for JSON response)

**Example:**
```
GET /search/entries?q=Italian&sort=name&order=asc
GET /search/entries?q=author:john&sort=created_at&order=desc
GET /search/entries?q=rating:4&format=json
```

**Response:** HTML template or JSON with entries and pagination info

### 3. Restaurant Search: `GET /search/restaurants`

Searches specifically for restaurants with enhanced filtering and sorting.

**Query Parameters:**
- `q` (required): Search query string
- `page` (optional): Page number for pagination (default: 1)
- `limit` (optional): Number of results per page (default: 20, max: 100)
- `sort` (optional): Sort field (name, location, cuisine, entry_count) (default: name)
- `order` (optional): Sort order (asc, desc) (default: asc)
- `format` (optional): Response format ("json" for JSON response)

**Example:**
```
GET /search/restaurants?q=Italian&sort=entry_count&order=desc
GET /search/restaurants?q=location:N&sort=name&order=asc
GET /search/restaurants?q=cuisine:Chinese&format=json
```

**Response:** HTML template or JSON with restaurants and pagination info

### 4. Search Suggestions: `GET /search/suggestions`

Returns search suggestions for autocomplete functionality.

**Query Parameters:**
- `q` (required): Partial search query (minimum 2 characters)

**Example:**
```
GET /search/suggestions?q=Ital
GET /search/suggestions?q=Test
```

**Response:** JSON with array of suggestions

## Search Query Syntax

The search system supports advanced query syntax:

### Author/Submitter Filters
- `author:username` - Find entries by specific user
- `submitter:username` - Same as author
- `by:username` - Same as author
- `from:username` - Same as author

### Wishlist/Unvisited Filters
- `wishlist` - Find unvisited restaurants (entries with visited=false)
- `want to go to` - Same as wishlist
- `looking forward` - Same as wishlist
- `to try` - Same as wishlist

### Location Filters
- `location:N` - Filter by location (N, NE, NW, S, SE, SW)
- `area:N` - Same as location
- `in:N` - Same as location

### Cuisine Filters
- `cuisine:Italian` - Filter by cuisine type
- `type:Italian` - Same as cuisine
- `food:Italian` - Same as cuisine

### Rating Filters
- `rating:4` - Find entries with any rating >= 4
- `rated:4` - Same as rating
- `stars:4` - Same as rating

### Text Search
Any text not matching the above patterns will be searched across:
- Restaurant/entry names
- Cuisine types
- Comments
- Ordered items

## Error Handling

All endpoints return appropriate HTTP status codes:

- `200 OK` - Successful search
- `400 Bad Request` - Missing or invalid parameters
- `500 Internal Server Error` - Database or server errors

Error responses include:
```json
{
  "error": "Error message",
  "details": "Additional error details (optional)"
}
```

## Pagination

All search endpoints support pagination with the following structure:

```json
{
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_items": 47,
    "items_per_page": 10,
    "has_next": true,
    "has_prev": false
  }
}
```

## Implementation Notes

1. **Case Insensitive**: All text searches are case-insensitive
2. **Partial Matching**: Text searches use LIKE queries with wildcards
3. **Performance**: Database indexes are recommended for name, submitter, location, and cuisine fields
4. **Security**: All inputs are properly sanitized to prevent SQL injection
5. **Validation**: Query parameters are validated for type and range
6. **Sorting**: Default sorting varies by endpoint (entries: newest first, restaurants: alphabetical)