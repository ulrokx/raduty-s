package util

import (
	"time"

	"github.com/ulrokx/raduty-s/api/models"
)

func dayInArray(arr []models.Unavailable, day time.Time) bool {
	for _, d := range arr {
		if d.Day.UTC() == day {
			return true
		}
	}
	return false
}

func CanWorkDay(ra models.Assistant, day time.Time) bool {
	if IsDayInMask(ra.DayAvailability, day.Weekday()) && !dayInArray(ra.Unavailable, day) {
		return true
	}
	return false
}

func NumShiftInSchedule(schedule []models.Shift, day time.Time, id uint) (count int, already bool) {
	for _, d := range schedule {
		if d.Date.UTC() == day {
			count++
			if d.AssistantID == id {
				already = true
			}
		}
	}
	return
}
