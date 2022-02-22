package util

import (
	"fmt"
	"time"

	"github.com/ulrokx/raduty-s/api/models"
)

func ParseDateArr(arr []string) (parsed []models.Unavailable, err error) {

	timeFormat := "01/02/2006"

	for _, d := range arr {
		if d[1] == '/' {
			d = fmt.Sprintf("0%s", d)
		}
		if d[4] == '/' {
			d = fmt.Sprintf("%s0%s", d[:3], d[3:])
		}
		parsedTime, perr := time.Parse(timeFormat, d)
		if err != nil {
			err = perr
			return
		}
		parsed = append(parsed, models.Unavailable{
			Day: parsedTime,
		})
	}
	return
}
