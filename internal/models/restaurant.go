package models

import (
	"gorm.io/gorm"
)

type Restaurant struct {
	// gorm.Model contains useful automatically created fields like auto-incrementing ID, CreatedAt date, etc.
	// see https://gorm.io/docs/models.html
	gorm.Model
	// Required fields below
	Name     string `form:"name" json:"name" gorm:"notNull" binding:"required"`
	Location string `form:"location" json:"location" gorm:"notNull" binding:"required"`
	Cuisine  string `form:"cuisine" json:"cuisine" gorm:"notNull" binding:"required"`
	Closed   bool   `form:"closed" json:"closed" gorm:"default:false"`
	// Optional fields below
	Address         string `form:"address" json:"address"`
	WebsiteURL      string `form:"websiteurl" json:"websiteurl"`
	ReservationsURL string `form:"reservationsurl" json:"reservationsurl"`
	MenuURL         string `form:"menuurl" json:"menuurl"`
	PhoneNumber     string `form:"phonenumber" json:"phonenumber"`
}
