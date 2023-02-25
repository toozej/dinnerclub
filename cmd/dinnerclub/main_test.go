package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gocondor/core/jwt"
	"github.com/gocondor/core/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/toozej/dinnerclub/internal/routers"
	"github.com/toozej/dinnerclub/pkg/authentication"
	"github.com/toozej/dinnerclub/pkg/session"
)

func sessionSetup() gin.HandlerFunc {
	return session.InitSession("testing123")
}

func authSetup() {
	jwt.New()
	authentication.New(sessions.Resolve(), jwt.Resolve())
}

func routerSetup() *gin.Engine {
	r := routers.NewRouter()
	r.Use(sessionSetup())
	routers.SetupTemplates()
	routers.SetupRouterDefaults("TST", "test_referral_code")
	routers.SetupStaticAssets()
	routers.SetupPublicRoutes()
	routers.SetupPrivateRoutes()
	authSetup()
	return r
}

func TestFavicon(t *testing.T) {
	contentType := "image/vnd.microsoft.icon"

	fileContents, err := os.ReadFile("../../assets/favicon.ico")
	assert.NoError(t, err, "Expected to read favicon file from assets/favicon.ico")

	r := routerSetup()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
	r.ServeHTTP(w, req)

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

	r := routerSetup()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)

	body, err := io.ReadAll(w.Body)
	assert.NoError(t, err, "Expected to read http body from response")

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.NotEqual(t, contentType, w.Header().Get("Content-Type"))
	assert.NotEqual(t, fileContents, body)
}

func TestHealthRoute(t *testing.T) {
	r := routerSetup()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestNotRoute(t *testing.T) {
	r := routerSetup()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/notfound", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "404")
}