package api

import (
	handlers "c2c/internal/api/handlers"
	utils "c2c/internal/api/utils"
	"c2c/internal/config"
	"c2c/internal/tools"

	"fmt"
)

type App struct {
	Env    handlers.Env
	Router Router

	DefaultTTL string
	ErrorTTL   string
}

func (a *App) Initialize(config *config.Config) error {

	a.Env.Config = config

	a.Env.Version = Version

	a.Router.Initialize(&a.Env)

	a.DefaultTTL = fmt.Sprintf("max-age=%d", a.Env.Config.Caching.TTL)
	a.ErrorTTL = fmt.Sprintf("max-age=%d", a.Env.Config.Caching.ErrorTTL)

	utils.InitProxyVersion(&a.Env)
	utils.InitRatingMap(&a.Env)

	tools.Logger.Debugf("Proxy major version is v%d", a.Env.ProxyVersion)

	return nil
}

// Run http server.
func (a *App) Run() {
	tools.Logger.Error(a.Router.Run())
}
