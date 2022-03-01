package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulrokx/raduty-s/api/models"
	"github.com/ulrokx/raduty-s/api/util"
	"google.golang.org/api/calendar/v3"
	"gorm.io/gorm"
)

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
	_, err = srv.Acl.Insert(cal.Id, &calendar.AclRule{
		Role: "reader",
		Scope: &calendar.AclRuleScope{
			Type: "default",
		},
	}).Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "failed to make calendar public",
		})
	}
	c.JSON(http.StatusCreated, gin.H{
		"message": "all good",
	})
}

func (s *Server) GetSelectedCalendars(c *gin.Context) {
	var selected models.Selected
	dberr := s.DB.Joins("Schedule").Find(&selected)
	if dberr.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   dberr.Error.Error(),
			"message": "could not load selected calendar",
		})
		return
	}
	c.JSON(http.StatusOK, selected)

}

type SetSelectedCalendarRequest struct {
	Group    uint `binding:"required"`
	Schedule uint `binding:"required"`
}

func (s *Server) SetSelectedCalendar(c *gin.Context) {
	var req SetSelectedCalendarRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "didn't bind json",
		})
		return
	}
	var query models.Selected

	dberr := s.DB.First(&query, "group_id = ?", req.Group)

	if errors.Is(dberr.Error, gorm.ErrRecordNotFound) {
		dberr = s.DB.Create(&models.Selected{
			ScheduleID: req.Schedule,
			GroupID:    req.Group,
		})
		if dberr.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   dberr.Error.Error(),
				"message": "bad first create",
			})
			return
		}
		c.Status(http.StatusCreated)
		return
	}

	s.DB.Model(&models.Selected{}).Where("")
	if dberr.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   dberr.Error.Error(),
			"message": "failed to update primary calendar",
		})
		return
	}
	c.Status(http.StatusCreated)
}
