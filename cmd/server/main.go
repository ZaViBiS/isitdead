package main

import (
	"os"
	"time"

	"ZaViBiS/isitdead/internal/config"
	"ZaViBiS/isitdead/internal/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).With().Caller().Logger()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("error importing env")
	}

	server.Run(cfg.ServerPort)
}
