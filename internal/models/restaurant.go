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
	// TODO additional fields for address, website URL, reservations URL, menu URL, phone number, etc.
}
