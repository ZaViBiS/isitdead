package api

func (s *Server) setupRoutes() {
	api := s.App.Group("/api")

	api.Get("/ping", s.handlePing)
	api.Post("/register", s.handleRegister)
	api.Post("/login", s.handleLogin)

	// Protected routes
	api.Use(s.authMiddleware)
	api.Get("/servers", s.handleGetServers)
	api.Post("/servers", s.handleAddServer)
	api.Delete("/servers/:id", s.handleDeleteServer)
	api.Get("/servers/:id/results", s.handleGetServerResults)
}
