package api

import (
	"errors"
	"strconv"
	"time"

	"github.com/ZaViBiS/isitdead/internal/model"
	"github.com/gofiber/fiber/v3"
)

const (
	defaultResultsLimit = 100
	maxResultsLimit     = 1000
	maxResultsHours     = 24 * 31
)

type resultsRequest struct {
	ServerID      uint
	IncidentsOnly bool
	Limit         int
	Hours         int
	Region        string
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
		results, err = s.DB.GetIncidentsForRegion(req.ServerID, req.Region, req.Limit)
	} else if req.Hours > 0 {
		since := time.Now().UTC().Add(-time.Duration(req.Hours) * time.Hour)
		results, err = s.DB.GetHistorySinceForRegion(req.ServerID, req.Region, since)
	} else {
		limit := req.Limit
		if limit <= 0 {
			limit = defaultResultsLimit
		}
		results, err = s.DB.GetRecentHistoryForRegion(req.ServerID, req.Region, limit)
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
		return resultsRequest{}, err
	}
	res.IncidentsOnly = c.Query("incidents") == "true"
	res.Region = c.Query("region")
	limitStr := c.Query("limit")
	if limitStr != "" {
		res.Limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return resultsRequest{}, err
		}
		if res.Limit < 0 || res.Limit > maxResultsLimit {
			return resultsRequest{}, errors.New("limit out of range")
		}
	}

	hoursStr := c.Query("hours")
	if hoursStr != "" {
		res.Hours, err = strconv.Atoi(hoursStr)
		if err != nil {
			return resultsRequest{}, err
		}
		if res.Hours < 0 || res.Hours > maxResultsHours {
			return resultsRequest{}, errors.New("hours out of range")
		}
	}

	return res, nil
}
