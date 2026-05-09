package app

import (
	"embed"

	"github.com/ZaViBiS/isitdead/internal/api"
	"github.com/ZaViBiS/isitdead/internal/database"
)

type App struct {
	server *api.Server
}

func New(staticFiles embed.FS) (*App, error) {
	// БД
	db, err := database.Init("/tmp/isitdead.db")
	if err != nil {
		return nil, err
	}

	// Backend + Frontend (embed)
	server, err := api.New(db, staticFiles)
	if err != nil {
		return nil, err
	}

	return &App{
		server: server,
	}, nil
}

func (a *App) Run() error {
	return a.server.Listen(":8080")
}
