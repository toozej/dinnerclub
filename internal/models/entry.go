package models

import (
	"gorm.io/gorm"
)

type Entry struct {
	// gorm.Model contains useful automatically created fields like auto-incrementing ID, CreatedAt date, etc.
	// see https://gorm.io/docs/models.html
	gorm.Model
	// Required fields below
	Submitter string `form:"submitter" json:"submitter" gorm:"notNull"`
	Name      string `form:"name" json:"name" gorm:"notNull" binding:"required"`
	Location  string `form:"location" json:"location" gorm:"notNull" binding:"required"`
	Cuisine   string `form:"cuisine" json:"cuisine" gorm:"notNull" binding:"required"`
	Visited   bool   `form:"visited" json:"visited"`
	// Optional fields below
	Closed         bool   `form:"closed" json:"closed" gorm:"default:false"`
	MealService    string `form:"mealservice" json:"mealservice"`
	Ordered        string `form:"ordered" json:"ordered"`
	FoodRating     int    `form:"foodrating" json:"foodrating"`
	AmbienceRating int    `form:"ambiencerating" json:"ambiencerating"`
	ValueRating    int    `form:"valuerating" json:"valuerating"`
	Comments       string `form:"comments" json:"comments"`
}
