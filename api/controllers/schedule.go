package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulrokx/raduty-s/api/models"
	"github.com/ulrokx/raduty-s/api/util"
	"google.golang.org/api/calendar/v3"
)

type GenerateRequest struct {
	BeginDate string `binding:"required"`
	EndDate   string `binding:"required"`
	PerShift  int    `binding:"required"`
	Group     int    `binding:"required"`
	Name      string `binding:"required"`
}

type IDToDays struct {
	RA   models.Assistant
	Days []time.Time
}

type ByDays []IDToDays

func (a ByDays) Len() int           { return len(a) }
func (a ByDays) Less(i, j int) bool { return len(a[i].Days) < len(a[j].Days) }
func (a ByDays) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (s *Server) GenerateSchedule(c *gin.Context) {
	var req GenerateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "bad request",
		})
		return
	}
	startDate, err := util.ParseDate(req.BeginDate)
	endDate, err := util.ParseDate(req.EndDate)
	// calculate the number of days each ra must work

	var totalWeekdays, totalWeekends, totalDays int
	for d := util.RangeDate(startDate, endDate); ; {
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
	// find how many days each ra can work
	// get all ras
	var allRAs []models.Assistant
	qres := s.DB.Preload("Unavailable").Find(&allRAs, "group_id = ?", req.Group)
	if qres.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   qres.Error,
			"message": "query error",
		})
	}
	//make map of ra to the days they can work
	var IDMap []IDToDays
	for idx, r := range allRAs {
		IDMap = append(IDMap, IDToDays{
			RA:   r,
			Days: []time.Time{},
		})
		for d := util.RangeDate(startDate, endDate); ; {
			date := d()
			if date.IsZero() {
				break
			}
			if util.CanWorkDay(r, date) {
				IDMap[idx].Days = append(IDMap[idx].Days, date)
			}
		}
	}

	// sort the ras by number of days available to work
	sort.Sort(ByDays(IDMap))
	// until a schedule is found, randomly try to create schedule
	numOfRAs := len(allRAs)
	if numOfRAs == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "divide by zero",
			"message": "empty ra list",
		})
		return
	}
	weekdaysPer, weekdayRem := (totalWeekdays*req.PerShift)/numOfRAs, (totalWeekdays*req.PerShift)%numOfRAs
	weekendsPer, weekendRem := (totalWeekends*req.PerShift)/numOfRAs, (totalWeekends*req.PerShift)%numOfRAs

	var schedule []models.Shift

	randy := rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Printf("----\nRAs: %d | weekendsPer: %d | weekdaysPer: %d\n----\n", numOfRAs, weekendsPer, weekdaysPer)
generator:

	for idx, ra := range IDMap {
		var weekendsLeft, weekdaysLeft = weekendsPer, weekdaysPer
		if weekendRem > 0 {
			weekendsLeft++
			weekendRem-- //if extra weekend days, give to person
		}
		if (weekdayRem+weekendRem > numOfRAs && idx >= numOfRAs-weekdayRem) || (weekendRem == 0 && weekdayRem > 0) {
			weekdaysLeft++ //if more extra days than people, stick onto end, or if there are no more weekends, stick after
			weekdayRem--
		}
		fmt.Printf("ra id: %d | weekdaysLeft: %d | weekendsLeft: %d\n", ra.RA.ID, weekdaysLeft, weekendsLeft)
		var tried []int
	stab:
		for weekdaysLeft > 0 || weekendsLeft > 0 {
			if len(tried) == totalDays {
				continue generator
			}

			offset := randy.Intn(totalDays)
			for _, v := range tried {
				if offset == v {
					continue stab
				}
			}
			tried = append(tried, offset)

			fmt.Printf("raid: %d weekdaysleft: %d weekendsleft: %d tried: %v\n", ra.RA.ID, weekdaysLeft, weekendsLeft, tried)
			toTest := startDate.AddDate(0, 0, offset)

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
	cerr := s.DB.Create(&models.Schedule{
		Shifts: schedule,
		Name:   req.Name,
	})
	if cerr.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   cerr.Error.Error(),
			"message": "failed to create schedule",
		})
	}
	c.JSON(200, schedule)

}

type CreateCalendarRequest struct {
	Schedule uint `binding:"required"`
}

func (s *Server) CreateCalendar(c *gin.Context) {
	//bind JSON to request
	var req CreateCalendarRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "invalid request",
		})
		return
	}

	//retrieve a calendar service object
	srv, err := util.GetCalendar()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "failed to get calendar",
		})
		return
	}

	// load schedule from database with id
	var schedule models.Schedule
	lres := s.DB.Preload("Shifts").Preload("Shifts.Assistant").Find(&schedule)
	if lres.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   lres.Error.Error(),
			"message": "could not find schedule to create calendar on",
		})
		return
	}

	//create a new calendar on google api
	cal, err := srv.Calendars.Insert(&calendar.Calendar{
		Summary: schedule.Name,
	}).Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "could not create calendar",
		})
		return
	}
	fmt.Printf("cal id: %v\n", cal.Id)
	//save google calendar id in table
	schedule.Calendar = cal.Id
	serr := s.DB.Save(&schedule)
	if serr.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   serr.Error.Error(),
			"message": "failed to save google calendar id",
		})
	}

	beginDuration, _ := time.ParseDuration("25h")
	endDuration, _ := time.ParseDuration("33h")
	for _, shift := range schedule.Shifts {
		event := &calendar.Event{
			Attendees: []*calendar.EventAttendee{
				{
					Email:       shift.Assistant.Email,
					DisplayName: fmt.Sprintf("%s %s", shift.Assistant.First, shift.Assistant.Last),
				},
			},
			Start: &calendar.EventDateTime{
				DateTime: shift.Date.Add(beginDuration).Format(time.RFC3339),
				TimeZone: "America/New_York",
			},
			End: &calendar.EventDateTime{
				DateTime: shift.Date.Add(endDuration).Format(time.RFC3339),
				TimeZone: "America/New_York",
			},
			Summary: fmt.Sprintf("%s %s", shift.Assistant.First, shift.Assistant.Last),
		}
		event, err = srv.Events.Insert(cal.Id, event).Do()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   err.Error(),
				"message": "failed to create event",
			})
			return
		}
		shift.CalendarID = event.Id
		s.DB.Save(&shift)

	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "all good",
	})

}
