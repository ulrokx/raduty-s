package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ulrokx/raduty-s/api/models"
)

func (s *Server) AllAssistants(c *gin.Context) {
	var assistants []models.Assistant
	s.DB.Find(&assistants)
	c.JSON(http.StatusOK, assistants)
}
