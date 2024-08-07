// env.go

package env

import (
	"c2c/internal/config"
	"c2c/internal/models"
)

type Env struct {
	ConfigFileName string
	Config         *config.Config
	Version        string
	ProxyVersion   int
	RatingMap      models.RatingMap
}
