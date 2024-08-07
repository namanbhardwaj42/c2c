// version.go

package utils

import (
	handlers "c2c/internal/api/handlers"
	"c2c/internal/tools"
	"strings"
)

func InitProxyVersion(e *handlers.Env) {

	api_version, err := VersionProxy(e)
	if err != nil {
		tools.Logger.Error("Error at proxy version %s", err)
	} else if strings.HasPrefix(api_version, "1") {
		e.ProxyVersion = 1
	} else if strings.HasPrefix(api_version, "3") {
		e.ProxyVersion = 3
	} else {
		tools.Logger.Error("Not a known proxy version")
	}
}
