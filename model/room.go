package model

import "gorm.io/gorm"

type Room struct {
	gorm.Model
	MasterName string `json:"master_name" gorm:"index"`
	GuestName  string `json:"guest_name" gorm:"index"`
}
