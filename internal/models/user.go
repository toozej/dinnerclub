package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"size:255;not null;unique" form:"username" json:"username" binding:"required,alphanum,lowercase"`
	Password     string `gorm:"size:255;not null;" form:"password" json:"-" binding:"required,min=10"`
	Firstname    string `gorm:"size:255" form:"firstname" json:"firstname"`
	Lastname     string `gorm:"size:255" form:"lastname" json:"lastname"`
	Email        string `gorm:"size:255;unique" form:"email" json:"email" binding:"required,email"`
	ReferralCode string `gorm:"size:255" form:"referralcode" json:"-"`
}
