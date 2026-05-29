package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"

	"github.com/ZaViBiS/isitdead/integration/telegram/api"
	"github.com/ZaViBiS/isitdead/integration/telegram/bot"
)

func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		log.Error().Msg("TOKEN is required")
		return
	}

	client, err := bot.New(token)
	if err != nil {
		log.Err(err).Msg("bot init error")
		return
	}
	log.Info().Msg("telegram bot initialized")

	go client.Start(context.Background())

	server := api.New(client)
	port := os.Getenv("PORT")
	if port == "" {
		port = "18081"
	}
	if err := server.Start(fmt.Sprintf(":%s", port)); err != nil {
		log.Err(err).Msg("telegram api error")
	}
}
