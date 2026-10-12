package server

import (
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

func root(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("ip", r.RemoteAddr).Str("lang", r.Header.Get("Accept-Language")).Msg("request")

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	fmt.Fprint(w, page)
}

func Run(port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", root)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	log.Info().Int("port", port).Msg("starting server")
	if err := srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
