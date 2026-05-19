package main

import (
	"context"
	"os"

	"github.com/ZaViBiS/isitdead/integration/telegram/api"
	"github.com/ZaViBiS/isitdead/integration/telegram/bot"
	"github.com/rs/zerolog/log"
)

func main() {
	token := os.Getenv("TOKEN")
	client, err := bot.New(token)
	if err != nil {
		log.Err(err).Msg("bot init error")
		return
	}
	log.Info().Msg("telegram bot initialized")

	go client.Start(context.Background())

	server := api.New(client)
	if err := server.Start(":8080"); err != nil {
		log.Err(err).Msg("telegram api error")
	}
}
