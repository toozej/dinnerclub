package database

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(connString string) {
	var db *gorm.DB
	var err error

	dsn := fmt.Sprintf("%s sslmode=disable TimeZone=America/Los_Angeles", connString)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	for err != nil {
		log.Fatalf("error connecting to database %v", err)
		time.Sleep(5 * time.Second)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

// func MigrateSchema() error {
// 	var schemaModels = []interface{}{
// 		models.entry{},
// 		models.user{},
// 	}
//
// 	for _, m := range models {
// 		if err := db.AutoMigrate(&m); err != nil {
// 			// handle err
// 		}
// 	}
//
// }
