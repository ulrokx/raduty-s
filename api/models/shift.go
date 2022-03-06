package models

import (
	"time"

	"gorm.io/gorm"
)

type Shift struct {
	gorm.Model
	AssistantID uint
	Assistant   Assistant
	ScheduleID  uint
	Date        time.Time
	CalendarID  string
}
