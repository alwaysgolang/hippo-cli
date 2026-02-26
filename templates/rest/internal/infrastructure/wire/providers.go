package wire

import (
	"github.com/google/wire"

	pingController "gotemplate/internal/adapter/http/controllers/ping"
	"gotemplate/internal/config"
	httpServer "gotemplate/internal/infrastructure/http"
)

var ConfigSet = wire.NewSet(
	ProvideAppConfig,
	ProvideHTTPConfig,
)

func ProvideAppConfig(cfg *config.Config) *config.AppConfig   { return &cfg.Application }
func ProvideHTTPConfig(cfg *config.Config) *config.HTTPConfig { return &cfg.HTTP }

var ControllerSet = wire.NewSet(
	pingController.NewController,
)

var ServerSet = wire.NewSet(
	httpServer.NewServer,
)

var AllProviders = wire.NewSet(
	ConfigSet,
	ControllerSet,
	ServerSet,
)
