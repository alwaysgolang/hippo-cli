package http

func (s *Server) registerRoutes() {
	api := s.Engine.Group("/api")
	{
		api.GET("/ping", s.PingController.Ping)
	}
}
