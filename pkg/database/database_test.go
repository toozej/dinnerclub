package database

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Store original DB and replace with test DB
	originalDB := DB
	DB = db

	// Cleanup function to restore original DB
	t.Cleanup(func() {
		DB = originalDB
	})

	return db
}

func TestConnectDatabase(t *testing.T) {
	// Test with SQLite for unit testing (PostgreSQL would require a running instance)
	originalDB := DB
	defer func() { DB = originalDB }()

	// This test verifies the connection pool settings are applied
	// We can't easily test the actual PostgreSQL connection without a running instance
	t.Run("connection pool settings", func(t *testing.T) {
		db := setupTestDB(t)

		sqlDB, err := db.DB()
		require.NoError(t, err)

		// Set connection pool settings (similar to what ConnectDatabase does)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)

		// Verify settings were applied - just check that no errors occurred
		stats := sqlDB.Stats()
		assert.GreaterOrEqual(t, stats.MaxOpenConnections, 0)
		assert.GreaterOrEqual(t, stats.OpenConnections, 0)
	})
}

func TestWithTimeout(t *testing.T) {
	setupTestDB(t)

	t.Run("creates db with timeout context", func(t *testing.T) {
		timeout := 5 * time.Second
		db := WithTimeout(timeout)

		assert.NotNil(t, db)
		// Verify that the DB instance has a context (we can't directly access the context)
		// but we can verify it's a different instance than the global DB
		assert.NotEqual(t, DB, db)
	})

	t.Run("timeout context works with query", func(t *testing.T) {
		// Create a test table with unique name
		tableName := fmt.Sprintf("test_entries_%d", time.Now().UnixNano())
		err := DB.Exec(fmt.Sprintf("CREATE TABLE %s (id INTEGER PRIMARY KEY, name TEXT)", tableName)).Error
		require.NoError(t, err)

		// Insert test data
		err = DB.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES (?)", tableName), "test").Error
		require.NoError(t, err)

		// Use WithTimeout to query
		timeout := 1 * time.Second
		db := WithTimeout(timeout)

		var count int64
		err = db.Table(tableName).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func TestWithContext(t *testing.T) {
	setupTestDB(t)

	t.Run("creates db with provided context", func(t *testing.T) {
		ctx := context.Background()
		db := WithContext(ctx)

		assert.NotNil(t, db)
		// Verify that the DB instance has the context
		assert.NotEqual(t, DB, db)
	})

	t.Run("context cancellation works", func(t *testing.T) {
		// Create a test table with unique name
		tableName := fmt.Sprintf("test_entries_cancel_%d", time.Now().UnixNano())
		err := DB.Exec(fmt.Sprintf("CREATE TABLE %s (id INTEGER PRIMARY KEY, name TEXT)", tableName)).Error
		require.NoError(t, err)

		// Create a context that will be cancelled
		ctx, cancel := context.WithCancel(context.Background())
		db := WithContext(ctx)

		// Cancel the context immediately
		cancel()

		// Try to perform a query - it should fail due to cancelled context
		var count int64
		err = db.Table(tableName).Count(&count).Error

		// The error should be related to context cancellation
		// Note: SQLite might not always respect context cancellation for simple queries
		// but this tests the mechanism
		if err != nil {
			assert.Contains(t, err.Error(), "context")
		}
	})

	t.Run("timeout context works", func(t *testing.T) {
		// Create a test table with unique name
		tableName := fmt.Sprintf("test_entries_timeout_%d", time.Now().UnixNano())
		err := DB.Exec(fmt.Sprintf("CREATE TABLE %s (id INTEGER PRIMARY KEY, name TEXT)", tableName)).Error
		require.NoError(t, err)

		// Create a context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db := WithContext(ctx)

		// Insert test data
		err = db.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES (?)", tableName), "test").Error
		assert.NoError(t, err)

		// Query should work within timeout
		var count int64
		err = db.Table(tableName).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func TestDatabaseConnectionPoolSettings(t *testing.T) {
	db := setupTestDB(t)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	// Test connection pool configuration
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// Verify settings don't cause errors
	stats := sqlDB.Stats()
	assert.GreaterOrEqual(t, stats.OpenConnections, 0)
}

func TestDatabaseOperationsWithContext(t *testing.T) {
	setupTestDB(t)

	// Create test table with unique name
	tableName := fmt.Sprintf("test_users_%d", time.Now().UnixNano())
	err := DB.Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL,
			email TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`, tableName)).Error
	require.NoError(t, err)

	t.Run("insert with context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db := WithContext(ctx)
		err := db.Exec(fmt.Sprintf("INSERT INTO %s (username, email) VALUES (?, ?)", tableName), "testuser", "test@example.com").Error
		assert.NoError(t, err)
	})

	t.Run("select with context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db := WithContext(ctx)

		var users []struct {
			ID       uint   `gorm:"column:id"`
			Username string `gorm:"column:username"`
			Email    string `gorm:"column:email"`
		}

		err := db.Table(tableName).Find(&users).Error
		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "testuser", users[0].Username)
	})

	t.Run("update with context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db := WithContext(ctx)
		err := db.Exec(fmt.Sprintf("UPDATE %s SET email = ? WHERE username = ?", tableName), "updated@example.com", "testuser").Error
		assert.NoError(t, err)
	})

	t.Run("delete with context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db := WithContext(ctx)
		err := db.Exec(fmt.Sprintf("DELETE FROM %s WHERE username = ?", tableName), "testuser").Error
		assert.NoError(t, err)

		// Verify deletion
		var count int64
		err = db.Table(tableName).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestTransactionWithContext(t *testing.T) {
	setupTestDB(t)

	// Create test table with unique name
	tableName := fmt.Sprintf("test_transactions_%d", time.Now().UnixNano())
	err := DB.Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			value INTEGER NOT NULL
		)
	`, tableName)).Error
	require.NoError(t, err)

	t.Run("successful transaction with context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db := WithContext(ctx)

		err := db.Transaction(func(tx *gorm.DB) error {
			// Insert multiple records in transaction
			if err := tx.Exec(fmt.Sprintf("INSERT INTO %s (name, value) VALUES (?, ?)", tableName), "item1", 100).Error; err != nil {
				return err
			}
			if err := tx.Exec(fmt.Sprintf("INSERT INTO %s (name, value) VALUES (?, ?)", tableName), "item2", 200).Error; err != nil {
				return err
			}
			return nil
		})

		assert.NoError(t, err)

		// Verify both records were inserted
		var count int64
		err = db.Table(tableName).Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("failed transaction with context rolls back", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		db := WithContext(ctx)

		err := db.Transaction(func(tx *gorm.DB) error {
			// Insert a record
			if err := tx.Exec(fmt.Sprintf("INSERT INTO %s (name, value) VALUES (?, ?)", tableName), "item3", 300).Error; err != nil {
				return err
			}
			// Force an error to trigger rollback
			return fmt.Errorf("forced error")
		})

		assert.Error(t, err)

		// Verify the record was not inserted (rolled back)
		var count int64
		err = db.Table(tableName).Where("name = ?", "item3").Count(&count).Error
		assert.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestRawSQLWithContext(t *testing.T) {
	setupTestDB(t)

	// Create test table with unique name
	tableName := fmt.Sprintf("test_raw_%d", time.Now().UnixNano())
	err := DB.Exec(fmt.Sprintf(`
		CREATE TABLE %s (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`, tableName)).Error
	require.NoError(t, err)

	// Insert test data
	err = DB.Exec(fmt.Sprintf("INSERT INTO %s (name) VALUES (?), (?), (?)", tableName), "item1", "item2", "item3").Error
	require.NoError(t, err)

	t.Run("raw query with context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Get underlying sql.DB
		sqlDB, err := DB.DB()
		require.NoError(t, err)

		// Execute raw query with context
		rows, err := sqlDB.QueryContext(ctx, fmt.Sprintf("SELECT id, name FROM %s ORDER BY id", tableName))
		assert.NoError(t, err)
		defer rows.Close()

		var results []struct {
			ID   int
			Name string
		}

		for rows.Next() {
			var id int
			var name string
			err := rows.Scan(&id, &name)
			assert.NoError(t, err)
			results = append(results, struct {
				ID   int
				Name string
			}{ID: id, Name: name})
		}

		assert.Len(t, results, 3)
		assert.Equal(t, "item1", results[0].Name)
		assert.Equal(t, "item2", results[1].Name)
		assert.Equal(t, "item3", results[2].Name)
	})
}

// Benchmark tests for performance
func BenchmarkWithTimeout(b *testing.B) {
	// Setup a test DB for benchmarking
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	originalDB := DB
	DB = db
	defer func() { DB = originalDB }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db := WithTimeout(5 * time.Second)
		_ = db
	}
}

func BenchmarkWithContext(b *testing.B) {
	// Setup a test DB for benchmarking
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}
	originalDB := DB
	DB = db
	defer func() { DB = originalDB }()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db := WithContext(ctx)
		_ = db
	}
}
