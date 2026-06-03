package api

import (
	"io/fs"
	"net/http"

	"github.com/go-chi/chi"
)

func routes(staticFiles fs.FS) http.Handler {
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", healthHandler)
	})

	r.Get("/*", staticHandler(staticFiles))

	return r
}
