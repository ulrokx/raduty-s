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

type PairD struct {
	Key   time.Time
	Value int
}

type PairA struct {
	Key   uint
	Value int
}

type PairListD []PairD
type PairListA []PairA

func (p PairListD) Len() int           { return len(p) }
func (p PairListD) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairListD) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func (p PairListA) Len() int           { return len(p) }
func (p PairListA) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairListA) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

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

	totalDays, totalWeekdays, totalWeekends := scheduling.DaysInRange(startDate, endDate)
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

	// fmt.Printf("IDMap: %v", IDMap)
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
	var best float64
	var toKeep []models.Shift
	for i := 0; i < 1000; i++ {
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
		var avg float64
		dayFreq := scheduling.DayMap(schedule)
		for _, s := range dayFreq {
			avg += float64(s) / float64(totalDays)
		}
		if avg > float64(best) {
			best = avg
			toKeep = schedule
			fmt.Println(best)
		}
	}
	// dayFreq := scheduling.DayMap(toKeep)
	dayFreq := scheduling.DayMap(toKeep)
	assFreq := scheduling.AssistantMap(toKeep)
	var dfs PairListD
	var afs PairListA
	for k, v := range dayFreq {
		dfs = append(dfs, PairD{Key: k, Value: v})
	}
	for k, v := range assFreq {
		afs = append(afs, PairA{Key: k, Value: v})
	}
	sort.Sort(dfs)
	sort.Sort(afs)
	fmt.Println(afs)
	var missingShifts int
	cerr := s.DB.Create(&models.Schedule{
		Shifts: toKeep,
		Name:   req.Name,
	})
	if cerr.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   cerr.Error.Error(),
			"message": "failed to create schedule",
		})
		return
	}
	if missingShifts >= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "not all days were able to be filled",
			"message": "failed to fill all days",
		})
		return
	}
	fmt.Printf("rows created: %d", cerr.RowsAffected)
	c.JSON(200, toKeep)

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
