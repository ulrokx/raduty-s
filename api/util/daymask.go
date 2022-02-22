package util

import "github.com/ulrokx/raduty-s/api/models"

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
