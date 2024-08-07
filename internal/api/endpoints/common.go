//common.go

package endpoints

import (
	handlers "c2c/internal/api/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Ping godoc
// @Tags ping
// @ID ping
// @Summary Do a ping
// @Description If alive, answer a pong
// @Produce json
// @Success 200 {string} string "pong"
// @Router /ping [get]
func HandlePing(e *handlers.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	}
}

type VersionStruct struct {
	Version string `example:"9.17.84" json:"api_version"`
}

// GetVersion godoc
// @Tags version
// @ID version
// @Summary Get current RestApi version
// @Produce json
// @Success 200 {object} VersionStruct "Current RestApi version"
// @Router /version [get]

func HandleGetVersion(e *handlers.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		var v VersionStruct
		v.Version = e.Version
		c.JSON(http.StatusOK, v)
	}
}

type HealthResponse struct {
	Version string `json:"version" example:"3.1.0"`
	Status  string `json:"status" example:"0"`
}

// GetHealth godoc
// @Tags health
// @ID health
// @Summary Get current health of the restApi
// @Produce json
// @Success 200 {object} HealthResponse "Current RestApi health"
// @Router /health [get]
func HandleGetHealth(e *handlers.Env) gin.HandlerFunc {
	return func(c *gin.Context) {
		var resp HealthResponse
		resp.Version = e.Version
		resp.Status = "healthy"
		c.JSON(http.StatusOK, resp)
	}
}
