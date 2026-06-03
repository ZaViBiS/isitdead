package main

import (
	"io/fs"
	"net/http"

	"github.com/ZaViBiS/isitdead/internal/api"
	"github.com/ZaViBiS/isitdead/web"
	"github.com/rs/zerolog/log"
)

func main() {
	dist, err := fs.Sub(web.Dist, "dist")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load frontend files")
	}

	r := api.NewRouter(dist)
	log.Info().Msg("starting server on 8080")
	log.Fatal().Err(http.ListenAndServe(":8080", r)).Msg("server stopped")
}
