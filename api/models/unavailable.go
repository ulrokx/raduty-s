package models

import (
	"time"

	"gorm.io/gorm"
)

type Unavailable struct {
	gorm.Model
	Day         time.Time
	AssistantID int
}
