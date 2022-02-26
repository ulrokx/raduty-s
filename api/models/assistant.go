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
	First           string
	Last            string
	CWID            string `gorm:"unique"`
	Email           string
	DayAvailability DayEnum
	Unavailable     []Unavailable `gorm:"constraint:OnUpdate:CASCADE"`
	CanResubmit     bool          `gorm:"default:FALSE"`
}
