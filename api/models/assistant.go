package models

import (
	"gorm.io/gorm"
)

type DayEnum uint8

const (
	Sunday DayEnum = 1 << iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

type Assistant struct {
	gorm.Model
	First           string        `json:"first"`
	Last            string        `json:"last"`
	CWID            string        `gorm:"unique" json:"cwid"`
	Email           string        `json:"email"`
	DayAvailability DayEnum       `json:"availability"`
	Unavailable     []Unavailable `gorm:"constraint:OnUpdate:CASCADE" json:"unavailable"`
	CanResubmit     bool          `gorm:"default:FALSE" json:"canResubmit"`
	GroupID         int           `gorm:"default:0" json:"groupId"`
	Shifts          []Shift       `gorm:"foreignKey:AssistantID" json:"shifts"`
}
