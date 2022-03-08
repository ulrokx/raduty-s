package controllers

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulrokx/raduty-s/api/models"
	"github.com/ulrokx/raduty-s/api/scheduling"
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

type ByDays []scheduling.IDToDays

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

	totalWeekdays, totalWeekends, totalDays := scheduling.DaysInRange(startDate, endDate)
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
	IDMap := scheduling.MapIDToDays(allRAs, startDate, endDate)

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

	schedule := scheduling.GenerateShifts(scheduling.GenShiftsReq{
		WeekendsPer: weekendsPer,
		WeekdaysPer: weekdaysPer,
		WeekendRem:  weekendRem,
		WeekdayRem:  weekdayRem,
		NumOfRAs:    numOfRAs,
		TotalDays:   totalDays,
		PerShift:    req.PerShift,
		StartDate:   startDate,
		IDMap:       IDMap,
	})
	//generate schedule

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
