package scheduling

import "time"

func RangeDate(start, end time.Time) func() time.Time {
	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, time.UTC)

	return func() time.Time {
		date := start
		start = start.AddDate(0, 0, 1)
		if start.After(end) {
			return time.Time{}
		}
		return date
	}

}
