package models

import "gorm.io/gorm"

type Schedule struct {
	gorm.Model
	Name     string
	Shifts   []Shift `gorm:"foreignKey:ScheduleID"`
	Calendar string
}
