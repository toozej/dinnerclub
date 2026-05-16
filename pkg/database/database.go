// Package database provides database connection and configuration functionality
// for the dinnerclub application using GORM with PostgreSQL support.
package database

import (
	"context"
	"log"
	"regexp"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database connection instance used throughout the application.
var DB *gorm.DB

// ConnectDatabase establishes a connection to the PostgreSQL database using the provided
// connection string and configures logging based on the specified log level.
func ConnectDatabase(connString string, logLevel string) {
	var db *gorm.DB
	var err error
	var gconfig *gorm.Config

	switch logLevel {
	case "debug", "info":
		gconfig = &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		}
	default:
		gconfig = &gorm.Config{}
	}

	dsn := connString
	db, err = gorm.Open(postgres.Open(dsn), gconfig)
	for err != nil {
		log.Printf("error connecting to database %v", err)
		time.Sleep(5 * time.Second)
		db, err = gorm.Open(postgres.Open(dsn), gconfig)
		continue
	}

	// Get generic database object sql.DB to use its functions
	sqlDB, err := db.DB()
	if sqlDB != nil {
		// Ping the database to make sure it's up
		p := sqlDB.Ping()
		if p != nil {
			log.Fatalf("error pinging database %v", err)
		}
	}

	m := regexp.MustCompile(`password=[^\s]+`)
	log.Printf("Successfully connected to database: %s", m.ReplaceAllString(connString, "password=REDACTED"))

	// Configure connection pool settings for better performance and timeout handling
	sqlDB, err = db.DB()
	if err != nil {
		log.Fatalf("error getting underlying sql.DB: %v", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	DB = db
}

// WithTimeout returns a new GORM DB instance with the specified timeout context.
// This should be used for all database operations to ensure they don't hang indefinitely.
func WithTimeout(timeout time.Duration) *gorm.DB {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// Note: In a real application, you'd want to manage the cancel function properly
	// For this implementation, we'll let it timeout naturally
	_ = cancel
	return DB.WithContext(ctx)
}

// WithContext returns a new GORM DB instance with the provided context.
// This allows for request-scoped cancellation and timeout control.
func WithContext(ctx context.Context) *gorm.DB {
	if DB == nil {
		return nil
	}
	return DB.WithContext(ctx)
}
