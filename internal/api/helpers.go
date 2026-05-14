package api

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog/log"
)

func parseServerID(c fiber.Ctx) (uint, error) {
	serverID, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(serverID), nil
}

func (s *Server) userOwnsServer(userID, serverID uint) bool {
	server, err := s.DB.GetServerByID(serverID)
	if err != nil {
		log.Err(err).Msg("db error")
		return false
	}
	return server.UserID == userID
}
