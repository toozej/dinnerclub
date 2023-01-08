package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFavicon(t *testing.T) {
	contentType := "image/vnd.microsoft.icon"

	fileContents, err := os.ReadFile("../../assets/favicon.ico")
	assert.NoError(t, err, "Expected to read favicon file from assets/favicon.ico")

	router := setupRouter("../../")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err, "Expected to read http body from response")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, contentType, w.Header().Get("Content-Type"))
	assert.Equal(t, fileContents, body)
}

func TestNotFavicon(t *testing.T) {
	contentType := "application/octet-stream"

	fileContents, err := os.ReadFile("../../assets/favicon.ico")
	assert.NoError(t, err, "Expected to read favicon file from assets/favicon.ico")

	router := setupRouter("../../")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err, "Expected to read http body from response")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, contentType, w.Header().Get("Content-Type"))
	assert.NotEqual(t, fileContents, body)
}

func TestPingRoute(t *testing.T) {
	router := setupRouter("../../")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestNotRoute(t *testing.T) {
	router := setupRouter("../../")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/notfound", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "404")
}
