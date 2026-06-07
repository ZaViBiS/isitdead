package main

import (
	"io/fs"
	"net/http"
	"os"

	"github.com/ZaViBiS/isitdead/internal/api"
	"github.com/ZaViBiS/isitdead/internal/database"
	"github.com/ZaViBiS/isitdead/web"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	dist, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load frontend files")
	}

	databaseURL := os.Getenv("DATABASE_URL")

	storage, err := database.New(databaseURL)
	if err != nil {
		log.Err(err).Msg("error while db init")
		panic(err)
	}
	storage.Migrate()

	r := api.NewRouter(dist)
	log.Info().Msg("starting server on 8080")
	log.Fatal().Err(http.ListenAndServe(":8080", r)).Msg("server stopped")
}
