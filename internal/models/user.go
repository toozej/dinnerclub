package models

type User struct {
	ID        uint   `json:"id" gorm:"primary_key"`
	Username  string `json:"username"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}
