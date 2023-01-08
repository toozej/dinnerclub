package models

type Entry struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	Name   string `json:"name"`
	Author string `json:"author"`
}
