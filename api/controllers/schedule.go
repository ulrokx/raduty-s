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
)

type GenerateRequest struct {
	BeginDate string   `binding:"required" json:"begin"`
	EndDate   string   `binding:"required" json:"end"`
	PerShift  int      `binding:"required" json:"perShift"`
	Groups    []string `binding:"required" json:"groups"`
	Name      string   `binding:"required" json:"name"`
}

type IDToDays struct {
	RA   models.Assistant
	Days []time.Time
}

type ByDays []IDToDays

func (a ByDays) Len() int           { return len(a) }
func (a ByDays) Less(i, j int) bool { return len(a[i].Days) < len(a[j].Days) }
func (a ByDays) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func (s *Server) GenerateSchedule(c *gin.Context) { //TODO need to extract this
	var req GenerateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		fmt.Println("error")
		fmt.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "bad request",
		})
		return
	}
	var startDate, endDate time.Time
	test, err := time.Parse(time.RFC3339, req.BeginDate)
	if err != nil {
		startDate, err = util.ParseDate(req.BeginDate)
		endDate, err = util.ParseDate(req.EndDate)
	} else {
		endDate, err = time.Parse(time.RFC3339, req.EndDate)
		startDate = test
	}
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
	qres := s.DB.Preload("Unavailable").Where("group_id IN ?", req.Groups).Find(&allRAs)
	if qres.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   qres.Error,
			"message": "query error",
		})
		return
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
	//generate schedule
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

	//check if valid

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

func (s *Server) GetSchedules(c *gin.Context) {
	res := []models.Schedule{}
	dberr := s.DB.Find(&res)
	if dberr.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   dberr.Error.Error(),
			"message": "failed to get schedules",
		})
		return
	}
	c.JSON(http.StatusOK, res)
}
