package models

import (
	"gorm.io/gorm"
)

type Entry struct {
	// gorm.Model contains useful automatically created fields like auto-incrementing ID, CreatedAt date, etc.
	// see https://gorm.io/docs/models.html
	gorm.Model
	// Required fields below
	Submitter string `json:"submitter" gorm:"notNull"`
	Name      string `json:"name" gorm:"notNull"`
	Location  string `json:"location" gorm:"notNull"`
	Cuisine   string `json:"cuisine" gorm:"notNull"`
	Visited   bool   `json:"visited" gorm:"notNull"`
	// Optional fields below
	Closed         bool   `json:"closed"`
	MealService    string `json:"mealservice"`
	Ordered        string `json:"ordered"`
	FoodRating     int    `json:"foodrating"`
	AmbienceRating int    `json:"ambiencerating"`
	ValueRating    int    `json:"valuerating"`
	Comments       string `json:"comments"`
}
