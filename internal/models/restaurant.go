// Package models defines the data models for the dinnerclub application
// including database schemas and validation rules.
package models

import (
	"gorm.io/gorm"
)

// Restaurant represents a restaurant with its basic information and optional details
// such as contact information and web links.
type Restaurant struct {
	// gorm.Model contains useful automatically created fields like auto-incrementing ID, CreatedAt date, etc.
	// see https://gorm.io/docs/models.html
	gorm.Model
	// Required fields below
	Name     string `form:"name" json:"name" gorm:"notNull;unique;index" binding:"required"`
	Location string `form:"location" json:"location" gorm:"notNull;index" binding:"required"`
	Cuisine  string `form:"cuisine" json:"cuisine" gorm:"notNull;index" binding:"required"`
	Closed   bool   `form:"closed" json:"closed" gorm:"default:false"`
	// Optional fields below
	Address         string `form:"address" json:"address"`
	WebsiteURL      string `form:"websiteurl" json:"websiteurl"`
	ReservationsURL string `form:"reservationsurl" json:"reservationsurl"`
	MenuURL         string `form:"menuurl" json:"menuurl"`
	PhoneNumber     string `form:"phonenumber" json:"phonenumber"`
}
