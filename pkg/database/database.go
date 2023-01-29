package database

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

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

	dsn := fmt.Sprintf("%s sslmode=disable TimeZone=America/Los_Angeles", connString)
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

	DB = db
}
