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

	// Telegram bot callback.
	tg := s.App.Group("/api/telegram")
	tg.Get("/token/:chat_id/:token", s.handleTelegramNewUser)
	tg.Get("/token", s.handleTelegramNewUser)

	// Protected routes
	api.Use(s.authMiddleware)
	api.Get("/me", s.handleGetMe)
	api.Get("/servers", s.handleGetServers)
	api.Get("/dashboard/servers", s.handleGetDashboardServers)
	api.Post("/servers", s.handleAddServer)
	api.Put("/servers/:id", s.handleUpdateServer)
	api.Get("/servers/:id/notifications", s.handleGetNotificationPreferences)
	api.Put("/servers/:id/notifications", s.handleUpdateNotificationPreferences)
	api.Delete("/servers/:id", s.handleDeleteServer)
	api.Get("/servers/:id/results", s.handleGetServerResults)
	api.Get("/telegram/status", s.handleGetTelegramStatus)
	api.Post("/telegram/link-token", s.handleCreateTelegramLinkToken)
}
