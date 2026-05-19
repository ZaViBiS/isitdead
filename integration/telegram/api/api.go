// Package api відповідає з взіїмодію з isitdead
package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type Server struct {
	router chi.Router
	sender MessageSender
}

type MessageSender interface {
	SendMessage(ctx context.Context, chatID int64, message string) error
}

type response struct {
	Message string `json:"message"`
}

func New(sender MessageSender) Server {
	r := chi.NewRouter()

	server := Server{
		router: r,
		sender: sender,
	}

	server.setupRoutes()

	return server
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	_ = writeJSON(w, http.StatusOK, response{Message: "pong"})
}

func (s *Server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Dur("latency", time.Since(start)).
			Msg("request")
	})
}

func (s *Server) setupRoutes() {
	s.router.Use(middleware.Recoverer)
	s.router.Use(s.logRequest)

	s.router.Route("/api", func(r chi.Router) {
		r.Get("/ping", s.handlePing)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}
