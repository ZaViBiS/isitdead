package api

func (s *Server) setupRoutes() {
	api := s.App.Group("/api")

	api.Get("/ping", s.handlePing)
	api.Post("/register", s.handleRegister)
	api.Post("/login", s.handleLogin)
	api.Get("/auth/confirm", s.handleConfirmEmail)

	// Google OAuth
	api.Get("/auth/google", s.handleGoogleLogin)
	api.Get("/auth/google/callback", s.handleGoogleCallback)

	// Public monitor pages
	s.App.Get("/status/:slug", s.handlePublicStatusPage)
	s.App.Get("/sitemap.xml", s.handleSitemap)
	api.Get("/public/monitors/:slug", s.handleGetPublicMonitor)
	api.Get("/public/monitors/:slug/results", s.handleGetPublicMonitorResults)

	// Protected routes
	api.Use(s.authMiddleware)
	api.Get("/servers", s.handleGetServers)
	api.Post("/servers", s.handleAddServer)
	api.Put("/servers/:id", s.handleUpdateServer)
	api.Get("/servers/:id/notifications", s.handleGetNotificationPreferences)
	api.Put("/servers/:id/notifications", s.handleUpdateNotificationPreferences)
	api.Delete("/servers/:id", s.handleDeleteServer)
	api.Get("/servers/:id/results", s.handleGetServerResults)

	admin := api.Group("/admin", s.adminMiddleware)
	admin.Get("/public-monitors", s.handleAdminGetPublicMonitors)
	admin.Put("/servers/:id/public", s.handleAdminUpdatePublicMonitor)
}
