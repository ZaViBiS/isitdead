// Package api відповідає з взіїмодію з isitdead
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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

type sendMessageRequest struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
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

func (s *Server) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if err := verifySharedSecret(r); err != nil {
		_ = writeJSON(w, http.StatusUnauthorized, response{Message: err.Error()})
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&req); err != nil {
		_ = writeJSON(w, http.StatusBadRequest, response{Message: "invalid request body"})
		return
	}
	if req.ChatID == 0 || req.Text == "" {
		_ = writeJSON(w, http.StatusBadRequest, response{Message: "chat_id and text are required"})
		return
	}

	if err := s.sender.SendMessage(r.Context(), req.ChatID, req.Text); err != nil {
		_ = writeJSON(w, http.StatusBadGateway, response{Message: "failed to send telegram message"})
		return
	}

	_ = writeJSON(w, http.StatusAccepted, response{Message: "accepted"})
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
		r.Post("/messages", s.handleSendMessage)
	})
}

func verifySharedSecret(r *http.Request) error {
	secret := os.Getenv("TELEGRAM_API_SECRET")
	if secret == "" {
		return nil
	}
	if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", secret) {
		return errors.New("invalid token")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}
