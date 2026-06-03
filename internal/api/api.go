package api

import (
	"io/fs"
	"net/http"
)

func NewRouter(staticFiles fs.FS) http.Handler {
	return routes(staticFiles)
}
