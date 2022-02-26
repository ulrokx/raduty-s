package controllers

func (s *Server) initializeRoutes() {
	v1 := s.Router.Group("/api/v1")
	{
		v1.POST("/availability", s.CreateAvailability)
		v1.GET("/availability/already", s.AlreadyRegistered)
		v1.POST("/availability/resubmit", s.AllowResubmit)

		v1.GET("/assistants/all", s.AllAssistants)

		v1.POST("/schedule/generate", s.GenerateSchedule)
	}
}
