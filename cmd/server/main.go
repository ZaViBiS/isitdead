// Package main є точкою входу сервера.
package main

import (
	"os"

	"github.com/ZaViBiS/isitdead"
	"github.com/ZaViBiS/isitdead/internal/app"
	"github.com/ZaViBiS/isitdead/internal/logger"

	"github.com/rs/zerolog/log"
)

func main() {
	// Ініціалізуємо логер
	debug := os.Getenv("DEBUG") == "true"
	logger.Setup(debug)

	a, err := app.New(isitdead.StaticFiles)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize application")
	}

	if err := a.Run(); err != nil {
		log.Fatal().Err(err).Msg("application error")
	}
}
