package api

import (
	"io/fs"
	"net/http"
	"strings"
)

func staticHandler(staticFiles fs.FS) http.HandlerFunc {
	fileServer := http.FileServer(http.FS(staticFiles))

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if file, err := staticFiles.Open(path); err == nil {
			_ = file.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		r.URL.Path = "/index.html"
		fileServer.ServeHTTP(w, r)
	}
}
