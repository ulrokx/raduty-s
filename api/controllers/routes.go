package controllers

func (s *Server) initializeRoutes() {
	v1 := s.Router.Group("/api/v1")
	{
		v1.POST("/availability", s.CreateAvailability)
		v1.GET("/availability/already", s.AlreadyRegistered)
		v1.POST("/availability/resubmit", s.AllowResubmit)

		v1.GET("/assistants/all", s.AllAssistants)

		v1.POST("/schedule/generate", s.GenerateSchedule)
		v1.GET("/schedule/get", s.GetSchedules)

		v1.POST("/calendar/create", s.CreateCalendar)
		v1.GET("/calendar/get", s.GetSelectedCalendars)
		v1.POST("/calendar/set", s.SetSelectedCalendar)

		v1.POST("/groups/create", s.CreateGroup)
		v1.GET("/groups/get", s.GetGroups)
		v1.POST("/groups/delete", s.DeleteGroup)
	}
}
