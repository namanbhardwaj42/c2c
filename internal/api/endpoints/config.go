//config.go

package endpoints

import (
	handlers "c2c/internal/api/handlers"
	utils "c2c/internal/api/utils"
	"c2c/internal/config"
	"c2c/internal/tools"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type ConfigResponse struct {
	Version   string           `json:"version" example:"3.1.0"`
	AssConfig config.Assistant `json:"assistant" example:"0"`
}

type ConfigSaveResponse struct {
	Version string `json:"version" example:"3.1.0"`
	Status  string `json:"status" example:"Success"`
}

// Search godoc
// @Tags Config
// @ID Config
// @Summary Get current assistant Config
// @Description If alive, answer a pong
// @Produce json
// @Success 200 {string} string "pong"
// @Router /ping [get]
func HandleAssConfigCall(c *gin.Context, e *handlers.Env) {
	switch c.Request.Method {
	case "GET":
		var resp ConfigResponse
		resp.Version = e.Version
		resp.AssConfig = e.Config.Assistant
		c.JSON(http.StatusOK, resp)
	case "PUT":
	case "POST":
		var resp ConfigSaveResponse
		resp.Version = e.Version
		ass := ConfigResponse{}
		if err := c.ShouldBindBodyWith(&ass, binding.JSON); err != nil {
			tools.Logger.Error("Incorrect payload")
			resp.Status = "Incorrect payload"
			c.JSON(http.StatusBadRequest, resp)
			return
		}

		e.Config.Assistant = ass.AssConfig

		if err := utils.WriteConfiguration(e.ConfigFileName, e.Config); err != nil {
			resp.Status = fmt.Sprintf("!!!! Can't save the configuration into the file !!!! => %s", err)
			c.JSON(http.StatusInternalServerError, resp)
			return
		}

		resp.Status = "Successfully saved the configuration file"
		c.JSON(http.StatusOK, resp)

	default:
		tools.Logger.Errorf("Unsupported method received: %s", c.Request.Method)

		c.JSON(http.StatusMethodNotAllowed, fmt.Sprintf("Unsupported method received: %s", c.Request.Method))
	}
}
