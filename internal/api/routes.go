package api

func (s *Server) setupRoutes() {
	api := s.App.Group("/api")

	api.Get("/ping", s.handlePing)
	api.Post("/register", s.handleRegister)
	api.Post("/login", s.handleLogin)

	// Google OAuth
	api.Get("/auth/google", s.handleGoogleLogin)
	api.Get("/auth/google/callback", s.handleGoogleCallback)

	// Protected routes
	api.Use(s.authMiddleware)
	api.Get("/servers", s.handleGetServers)
	api.Post("/servers", s.handleAddServer)
	api.Delete("/servers/:id", s.handleDeleteServer)
	api.Get("/servers/:id/results", s.handleGetServerResults)
}
