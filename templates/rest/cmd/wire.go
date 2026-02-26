//go:build wireinject
// +build wireinject

package main

import (
	wirePkg "github.com/google/wire"

	"gotemplate/internal/config"
	httpServer "gotemplate/internal/infrastructure/http"
	"gotemplate/internal/infrastructure/wire"
)

type App struct {
	Server *httpServer.Server
}

func NewApp(server *httpServer.Server) *App {
	return &App{
		Server: server,
	}
}

func InitializeApp(cfg *config.Config) (*App, func(), error) {
	wirePkg.Build(
		wire.AllProviders,
		NewApp,
	)
	return &App{}, nil, nil
}
