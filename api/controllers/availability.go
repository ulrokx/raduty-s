package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ulrokx/raduty-s/api/models"
	"github.com/ulrokx/raduty-s/api/util"
	"gorm.io/gorm"
)

//TODO add so if the form is alread submitted then over write it and return a message saying that
type CreateAvailabilityRequest struct {
	First string
	Last  string
	CWID  string
	Email string
	Days  []string
	Dates []string
}

type CWIDRequest struct {
	CWID string `binding:"required"`
}

func (s *Server) CreateAvailability(c *gin.Context) {
	var req CreateAvailabilityRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err,
			"message": "JSON did not bind correctly",
		})
		fmt.Println("bad json")
	}
	daymask := util.ArrToMask(req.Days)

	parsedDates, perr := util.ParseDateArr(req.Dates)
	if perr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   perr.Error(),
			"message": "invalid date formatting",
		})
	}
	assistant := models.Assistant{
		Model:           gorm.Model{},
		First:           req.First,
		Last:            req.Last,
		CWID:            req.CWID,
		Email:           req.Email,
		DayAvailability: daymask,
		Unavailables:    parsedDates,
	}
	//see if assistant already registered could be extracted
	var raEntry models.Assistant
	qRes := s.DB.First(&raEntry, "cw_id = ?", req.CWID)
	if errors.Is(qRes.Error, gorm.ErrRecordNotFound) {
		s.DB.Create(&assistant)
		c.JSON(http.StatusCreated, gin.H{
			"message": "registered RA",
		})

	} else {
		if raEntry.CanResubmit == true {
			s.DB.Where("assistant_id = ?", raEntry.ID).Delete(&models.Unavailable{})
			raEntry.DayAvailability = daymask
			raEntry.Unavailables = parsedDates
			raEntry.CanResubmit = false
			s.DB.Save(&raEntry)

			c.JSON(http.StatusCreated, gin.H{
				"message": "re-registered RA",
			})
		} else {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "not allowed to modify availability",
			})
		}
	}
}

func (s *Server) AlreadyRegistered(c *gin.Context) {
	var req CWIDRequest
	berr := c.ShouldBindJSON(&req)
	if berr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   berr.Error(),
			"message": "not valid request",
		})
	}
	var ra models.Assistant
	dberr := s.DB.First(&ra, "cw_id = ?", req.CWID)
	if errors.Is(dberr.Error, gorm.ErrRecordNotFound) {
		c.Status(http.StatusOK)
	} else {
		if ra.CanResubmit == false {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   http.StatusForbidden,
				"message": "not allowed",
			})
		} else {
			c.Status(http.StatusOK)
		}
	}

}

func (s *Server) AllowResubmit(c *gin.Context) {
	var req CWIDRequest
	berr := c.ShouldBindJSON(&req)

	if berr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   berr.Error(),
			"message": "bad request",
		})
	} else {
		var ra models.Assistant
		dberr := s.DB.First(&ra, "cw_id = ?", req.CWID)
		if errors.Is(dberr.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   dberr.Error,
				"message": "ra not found",
			})
		} else {
			ra.CanResubmit = true
			s.DB.Save(&ra)
		}
	}
}
