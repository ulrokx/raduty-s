package models

type Selected struct {
	ScheduleID uint
	Schedule   Schedule
	GroupID    uint `gorm:"unique"`
	Group      Group
}
