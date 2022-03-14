package scheduling

import (
	"time"

	"github.com/ulrokx/raduty-s/api/models"
)

// Generates a map of assistant id to the number of shifts they are scheduled for
func AssistantMap(schedule []models.Shift) map[uint]int {
	res := make(map[uint]int)
	for _, a := range schedule {
		res[a.AssistantID] += 1
	}
	return res
}

// Generates a map of times to the number of shifts that fall on that time
func DayMap(schedule []models.Shift) map[time.Time]int {
	res := make(map[time.Time]int)
	for _, s := range schedule {
		res[s.Date] += 1
	}
	return res
}
