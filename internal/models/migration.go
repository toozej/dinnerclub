package models

import (
	"log"

	"github.com/toozej/dinnerclub/pkg/database"
)

func MigrateSchema() {
	var schemaModels = []interface{}{
		Entry{},
		User{},
	}

	for m := range schemaModels {
		if err := database.DB.AutoMigrate(&schemaModels[m]); err != nil {
			log.Fatalf("Failed to migrate database schema: %s", err)
		}
	}
	log.Printf("Successfully migrated all database schemas")
}
