package models

import (
	"time"

	"gorm.io/gorm"
)

type Shift struct {
	gorm.Model
	AssistantID uint
	ScheduleID  uint
	Date        time.Time
}
