package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ulrokx/raduty-s/api/models"
	"gorm.io/gorm"
)

type CreateGroupRequest struct {
	gorm.Model
	Name string `binding:"required"`
}

func (s *Server) CreateGroup(c *gin.Context) {
	var req CreateGroupRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "bad request",
		})
		return
	}

	dberr := s.DB.Create(&models.Group{
		Name: req.Name,
	})
	if dberr.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "failed to create group",
		})
	}
}

func (s *Server) GetGroups(c *gin.Context) {
	var res []models.Group
	dberr := s.DB.Find(&res)
	if dberr.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   dberr.Error.Error(),
			"message": "failed to get all groups",
		})
		return
	}
	c.JSON(http.StatusOK, res)
}

type DeleteGroupRequest struct {
	Group uint `binding:"required"`
}

func (s *Server) DeleteGroup(c *gin.Context) {
	req := DeleteGroupRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "bad request",
		})
	}
	dberr := s.DB.Delete(&models.Group{}, req.Group)
	if dberr.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   dberr.Error.Error(),
			"message": "failed to delete group",
		})
	}
	c.Status(http.StatusOK)
}
