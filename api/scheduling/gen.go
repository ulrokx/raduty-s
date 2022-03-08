package scheduling

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ulrokx/raduty-s/api/models"
	"github.com/ulrokx/raduty-s/api/util"
)

type GenShiftsReq struct {
	WeekendsPer int
	WeekdaysPer int
	WeekendRem  int
	WeekdayRem  int
	NumOfRAs    int
	TotalDays   int
	PerShift    int
	StartDate   time.Time
	IDMap       []IDToDays
}

func GenerateShifts(req GenShiftsReq) (
	schedule []models.Shift) {

	randy := rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Printf("----\nRAs: %d | weekendsPer: %d | weekdaysPer: %d\n----\n", req.NumOfRAs, req.WeekendsPer, req.WeekdaysPer)
generator:

	for idx, ra := range req.IDMap {
		var weekendsLeft, weekdaysLeft = req.WeekendsPer, req.WeekdaysPer
		if req.WeekendRem > 0 {
			weekendsLeft++
			req.WeekendRem-- //if extra weekend days, give to person
		}
		if (req.WeekdayRem+req.WeekendRem > req.NumOfRAs && idx >= req.NumOfRAs-req.WeekdayRem) || (req.WeekendRem == 0 && req.WeekdayRem > 0) {
			weekdaysLeft++ //if more extra days than people, stick onto end, or if there are no more weekends, stick after
			req.WeekdayRem--
		}
		fmt.Printf("ra id: %d | weekdaysLeft: %d | weekendsLeft: %d\n", ra.RA.ID, weekdaysLeft, weekendsLeft)
		var tried []int
	stab:
		for weekdaysLeft > 0 || weekendsLeft > 0 {
			if len(tried) == req.TotalDays {
				continue generator
			}

			offset := randy.Intn(req.TotalDays)
			for _, v := range tried {
				if offset == v {
					continue stab
				}
			}
			tried = append(tried, offset)

			fmt.Printf("raid: %d weekdaysleft: %d weekendsleft: %d tried: %v\n", ra.RA.ID, weekdaysLeft, weekendsLeft, tried)
			toTest := req.StartDate.AddDate(0, 0, offset)

			if util.CanWorkDay(ra.RA, toTest) {

				fmt.Printf("can work day: %d | offset: %d\n", toTest.Day(), offset)

				if (toTest.Weekday() <= time.Thursday && weekdaysLeft == 0) || (toTest.Weekday() >= time.Friday && weekendsLeft == 0) {
					fmt.Printf("alreadyMetLimit\n")
					continue
				} //if the day goes over that type of shift

				count, inAlready := util.NumShiftInSchedule(schedule, toTest, ra.RA.ID)
				if inAlready || count == req.PerShift {
					fmt.Printf("inAlready: %v | count: %d\n", inAlready, count)
					continue
				} //if that person has alaready been given this slot

				schedule = append(schedule, models.Shift{
					Date:        toTest,
					AssistantID: ra.RA.ID,
				})
				if toTest.Weekday() <= time.Thursday {
					weekdaysLeft--
				} else {
					weekendsLeft--
				}
			}
		}

	}
	return
}

func DaysInRange(startDate, endDate time.Time) (totalDays, totalWeekdays, totalWeekends int) {

	for d := RangeDate(startDate, endDate); ; {
		date := d()
		if date.IsZero() {
			break
		}
		if date.Weekday() <= time.Thursday {
			totalWeekdays++
		} else {
			totalWeekends++
		}
		totalDays++
	}
	return
}

type IDToDays struct {
	RA   models.Assistant
	Days []time.Time
}

func MapIDToDays(allRAs []models.Assistant, startDate time.Time, endDate time.Time) (IDMap []IDToDays) {
	for idx, r := range allRAs {
		IDMap = append(IDMap, IDToDays{
			RA:   r,
			Days: []time.Time{},
		})
		for d := RangeDate(startDate, endDate); ; {
			date := d()
			if date.IsZero() {
				break
			}
			if util.CanWorkDay(r, date) {
				IDMap[idx].Days = append(IDMap[idx].Days, date)
			}
		}
	}
	return
}
