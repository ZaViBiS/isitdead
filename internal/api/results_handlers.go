package api

import (
	"strconv"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/gofiber/fiber/v3"
)

type resultsRequest struct {
	ServerID      uint
	IncidentsOnly bool
	Limit         int
	Hours         int
}

func (s *Server) handleGetServerResults(c fiber.Ctx) error {
	userID := c.Locals("user_id").(uint)
	req, err := parseResultsQuery(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request params"})
	}

	if s.userOwnsServer(userID, req.ServerID) {
		return s.respondWithServerResults(c, req)
	}
	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Access denied"})
}

func (s *Server) respondWithServerResults(c fiber.Ctx, req resultsRequest) error {
	var (
		results []model.CheckResult
		err     error
	)

	if req.IncidentsOnly {
		results, err = s.DB.GetIncidents(req.ServerID, req.Limit)
	} else if req.Hours > 0 {
		since := time.Now().UTC().Add(-time.Duration(req.Hours) * time.Hour)
		results, err = s.DB.GetHistorySince(req.ServerID, since)
	} else {
		limit := req.Limit
		if limit <= 0 {
			limit = 100
		}
		results, err = s.DB.GetRecentHistory(req.ServerID, limit)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not fetch results"})
	}

	return c.JSON(results)
}

func parseResultsQuery(c fiber.Ctx) (resultsRequest, error) {
	var res resultsRequest
	var err error

	res.ServerID, err = parseServerID(c)
	if err != nil {
		panic(err)
	}
	res.IncidentsOnly = c.Query("incidents") == "true"
	limitStr := c.Query("limit")
	if limitStr != "" {
		res.Limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return resultsRequest{}, err
		}
	}

	hoursStr := c.Query("hours")
	if hoursStr != "" {
		res.Hours, err = strconv.Atoi(hoursStr)
		if err != nil {
			return resultsRequest{}, err
		}
	}

	return res, nil
}
