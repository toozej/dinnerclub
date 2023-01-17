package helpers

import (
	// "testing"

	// "github.com/gin-gonic/gin"
	"github.com/toozej/dinnerclub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	// "net/http"
	// "net/http/httptest"
)

type testable interface {
	Helper()
	Fatalf(format string, args ...any)
}

func FreshDb(t testable, path ...string) *gorm.DB {
	t.Helper()

	var dbUri string

	// Note: path can be specified in an individual test for debugging
	// purposes -- so the db file can be inspected after the test runs.
	// Normally it should be left off so that a truly fresh memory db is
	// used every time.
	if len(path) == 0 {
		dbUri = ":memory:"
	} else {
		dbUri = path[0]
	}

	db, err := gorm.Open(sqlite.Open(dbUri), &gorm.Config{})
	if err != nil {
		t.Fatalf("Error opening memory db: %s", err)
	}
	models.MigrateSchema()
	return db
}

// func getHasStatus(t *testing.T, db *gorm.DB, path string, status int) *httptest.ResponseRecorder {
// 	t.Helper()
//
// 	w := httptest.NewRecorder()
// 	ctx, router := gin.CreateTestContext(w)
// 	os.Setenv("AKLATAN_SESSION_KEY", "dummy")
// 	setupRouter(router, db)
//
// 	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
// 	if err != nil {
// 		t.Errorf("got error: %s", err)
// 	}
// 	router.ServeHTTP(w, req)
// 	if status != w.Code {
// 		t.Errorf("expected response code %d, got %d", status, w.Code)
// 	}
// 	return w
// }
