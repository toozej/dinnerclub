// Package models defines the data models for the dinnerclub application
// including database schemas and validation rules.
package models

import (
	"gorm.io/gorm"
)

// Entry represents a restaurant entry submitted by a user, containing both required
// and optional information about their dining experience or future dining plans.
type Entry struct {
	// gorm.Model contains useful automatically created fields like auto-incrementing ID, CreatedAt date, etc.
	// see https://gorm.io/docs/models.html
	gorm.Model
	// Required fields below
	Submitter string `form:"submitter" json:"submitter" gorm:"notNull;index"`
	Name      string `form:"name" json:"name" gorm:"notNull;index" binding:"required"`
	Location  string `form:"location" json:"location" gorm:"notNull" binding:"required"`
	Cuisine   string `form:"cuisine" json:"cuisine" gorm:"notNull;index" binding:"required"`
	Visited   bool   `form:"visited" json:"visited" gorm:"index"`
	// Enhanced fields for conditional entry forms
	LookingForwardTo string `form:"lookingforwardto" json:"lookingforwardto" gorm:"type:text"`
	// Optional fields below
	Closed         bool   `form:"closed" json:"closed" gorm:"default:false"`
	MealService    string `form:"mealservice" json:"mealservice"`
	Ordered        string `form:"ordered" json:"ordered"`
	OrderNext      string `form:"ordernext" json:"ordernext"`
	DoNotOrder     string `form:"donotorder" json:"donotorder"`
	FoodRating     int    `form:"foodrating" json:"foodrating" gorm:"check:food_rating >= 0 AND food_rating <= 5"`
	AmbienceRating int    `form:"ambiencerating" json:"ambiencerating" gorm:"check:ambience_rating >= 0 AND ambience_rating <= 5"`
	ValueRating    int    `form:"valuerating" json:"valuerating" gorm:"check:value_rating >= 0 AND value_rating <= 5"`
	Comments       string `form:"comments" json:"comments"`
}
