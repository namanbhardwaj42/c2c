package endpoints

import (
	handlers "c2c/internal/api/handlers"
	utils "c2c/internal/api/utils"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProxyHealthResponse struct {
	Version     string `json:"version" example:"3.1.0"`
	C2CStatus   string `json:"c2c_status" example:"0"`
	ProxyStatus struct {
		Url     string `json:"url" example:"http://proxy.url.net/"`
		Version string `json:"version" example:"3.1.0"`
		Health  string `json:"health" example:"OK"`
		Message string `json:"message" example:"none"`
	} `json:"proxy_status" example:"0"`
}

// Search godoc
// @Tags Search
// @ID Search
// @Summary Do a Search
// @Description If alive, answer a pong
// @Produce json
// @Success 200 {object} ProxyHealthResponse "pong"
// @Router /proxyhealth [get]
func HandleGetProxyHealth(c *gin.Context, e *handlers.Env) {
	var resp ProxyHealthResponse

	resp.Version = e.Version
	resp.C2CStatus = "healthy"

	resp.ProxyStatus.Url = e.Config.ProxyUrl
	proxy_version, err := utils.VersionProxy(e)
	if err != nil {
		resp.ProxyStatus.Version = "No Version"
		resp.ProxyStatus.Health = "KO"
		resp.ProxyStatus.Message = fmt.Sprint(err)
	} else {
		resp.ProxyStatus.Version = proxy_version
		resp.ProxyStatus.Health = "OK"
		resp.ProxyStatus.Message = ""
	}

	c.JSON(http.StatusOK, resp)
}
