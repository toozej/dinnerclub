package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"size:255;not null;unique" json:"username"`
	Password     string `gorm:"size:255;not null;" json:"-"`
	Firstname    string `gorm:"size:255" json:"firstname"`
	Lastname     string `gorm:"size:255" json:"lastname"`
	Email        string `gorm:"size:255;unique" json:"email" binding:"email"`
	ReferralCode string `gorm:"size:255;unique" json:"-"`
}
