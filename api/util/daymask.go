package util

import (
	"time"

	"github.com/ulrokx/raduty-s/api/models"
)

func ArrToMask(arr []string) (daymask models.DayEnum) {
	for _, d := range arr {
		switch d {
		case "sunday":
			daymask |= models.Sunday
		case "monday":
			daymask |= models.Monday
		case "tuesday":
			daymask |= models.Tuesday
		case "wednesday":
			daymask |= models.Wednesday
		case "thursday":
			daymask |= models.Thursday
		case "friday":
			daymask |= models.Friday
		case "saturday":
			daymask |= models.Saturday
		}
	}
	return
}

func IsDayInMask(mask models.DayEnum, day time.Weekday) bool {
	switch day {
	case time.Sunday:
		return mask&models.Sunday == models.Sunday
	case time.Monday:
		return mask&models.Monday == models.Monday
	case time.Tuesday:
		return mask&models.Tuesday == models.Tuesday
	case time.Wednesday:
		return mask&models.Wednesday == models.Wednesday
	case time.Thursday:
		return mask&models.Thursday == models.Thursday
	case time.Friday:
		return mask&models.Friday == models.Friday
	case time.Saturday:
		return mask&models.Saturday == models.Saturday
	}
	return false
}
