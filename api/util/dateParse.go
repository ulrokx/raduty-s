package util

import (
	"fmt"
	"time"

	"github.com/ulrokx/raduty-s/api/models"
)

func ParseDate(d string) (parsed time.Time, err error) {
	timeFormat := "01/02/2006"
	if d[1] == '/' {
		d = fmt.Sprintf("0%s", d)
	}
	if d[4] == '/' {
		d = fmt.Sprintf("%s0%s", d[:3], d[3:])
	}
	parsed, err = time.Parse(timeFormat, d)
	if err != nil {
		return
	}
	return
}

func ParseDateArr(arr []string) (parsed []models.Unavailable, err error) {

	for _, d := range arr {
		var parsedDate time.Time
		parsedDate, err = ParseDate(d)
		if err != nil {
			return
		}
		parsed = append(parsed, models.Unavailable{
			Day: parsedDate,
		})
	}
	return
}
