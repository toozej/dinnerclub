package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewResponseHelper(t *testing.T) {
	helper := NewResponseHelper()

	if helper == nil {
		t.Error("Expected response helper to be initialized")
	}
}

func TestResponseHelper_IsHTMLRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	helper := NewResponseHelper()

	tests := []struct {
		name       string
		setupReq   func(*gin.Context)
		expectHTML bool
	}{
		{
			name: "JSON Accept header",
			setupReq: func(c *gin.Context) {
				c.Request.Header.Set("Accept", "application/json")
			},
			expectHTML: false,
		},
		{
			name: "HTML Accept header",
			setupReq: func(c *gin.Context) {
				c.Request.Header.Set("Accept", "text/html")
			},
			expectHTML: true,
		},
		{
			name: "JSON query parameter",
			setupReq: func(c *gin.Context) {
				c.Request.URL.RawQuery = "format=json"
			},
			expectHTML: false,
		},
		{
			name: "Default browser request",
			setupReq: func(c *gin.Context) {
				// No specific headers or parameters
			},
			expectHTML: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			tt.setupReq(c)

			result := helper.IsHTMLRequest(c)
			if result != tt.expectHTML {
				t.Errorf("Expected IsHTMLRequest to return %v, got %v", tt.expectHTML, result)
			}
		})
	}
}

func TestResponseHelper_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	helper := NewResponseHelper()

	tests := []struct {
		name         string
		setupReq     func(*gin.Context)
		data         interface{}
		message      string
		expectStatus int
	}{
		{
			name: "JSON success response",
			setupReq: func(c *gin.Context) {
				c.Request.Header.Set("Accept", "application/json")
			},
			data:         map[string]string{"key": "value"},
			message:      "Success",
			expectStatus: http.StatusOK,
		},
		{
			name: "HTML success response",
			setupReq: func(c *gin.Context) {
				c.Request.Header.Set("Accept", "text/html")
			},
			data:         nil,
			message:      "Success",
			expectStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			tt.setupReq(c)

			helper.Success(c, tt.data, tt.message)

			if w.Code != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, w.Code)
			}
		})
	}
}

func TestResponseHelper_Error(t *testing.T) {
	gin.SetMode(gin.TestMode)
	helper := NewResponseHelper()

	tests := []struct {
		name         string
		setupReq     func(*gin.Context)
		statusCode   int
		errorCode    string
		message      string
		expectStatus int
	}{
		{
			name: "JSON error response",
			setupReq: func(c *gin.Context) {
				c.Request.Header.Set("Accept", "application/json")
			},
			statusCode:   http.StatusBadRequest,
			errorCode:    "BAD_REQUEST",
			message:      "Invalid input",
			expectStatus: http.StatusBadRequest,
		},
		{
			name: "HTML error response",
			setupReq: func(c *gin.Context) {
				c.Request.Header.Set("Accept", "text/html")
			},
			statusCode:   http.StatusNotFound,
			errorCode:    "NOT_FOUND",
			message:      "Resource not found",
			expectStatus: http.StatusFound, // HTML requests redirect (302)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			tt.setupReq(c)

			helper.Error(c, tt.statusCode, tt.errorCode, tt.message, nil)

			if w.Code != tt.expectStatus {
				t.Errorf("Expected status %d, got %d", tt.expectStatus, w.Code)
			}
		})
	}
}

func TestResponseHelper_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	helper := NewResponseHelper()

	validationErrors := []ValidationErrorResponse{
		{Field: "name", Message: "Name is required"},
		{Field: "email", Message: "Invalid email format"},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", nil)
	c.Request.Header.Set("Accept", "application/json")

	helper.ValidationError(c, "Validation failed", validationErrors)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
