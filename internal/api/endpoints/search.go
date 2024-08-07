//search.go

package endpoints

import (
	handlers "c2c/internal/api/handlers"
	utils "c2c/internal/api/utils"
	"c2c/internal/models"
	"c2c/internal/tools"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Search godoc
// @Tags Search
// @ID Search
// @Summary Do a Search
// @Description If alive, answer a pong
// @Produce json
// @Success 200 {string} string "pong"
// @Router /ping [get]
func HandleSearchCall(c *gin.Context, e *handlers.Env) {
	var resp models.GasResponse
	var proxy_items []interface{}
	var st models.SearchType
	var status_code int
	st.CanFallback = false
	st.ShouldFallback = true

	car, err := utils.ParsingParams(c)
	if err != nil {
		tools.Logger.Error("Parsing params failed with: ", err)
		resp.SetError(err)
		c.JSON(http.StatusOK, resp)
		return
	}

	for st.ShouldFallback {
		params, err := utils.ConvertToProxyParams(car, &e.Config.Assistant, &st)
		if err != nil {
			tools.Logger.Error("Converting to proxy params failed with: ", err)
			resp.SetError(err)
			c.JSON(http.StatusOK, resp)
			return
		}

		proxy_resp, err := utils.SearchProxy(e, params)
		if err != nil {
			tools.Logger.Error("Proxy search failed with ", err)
			resp.SetError(err)
			c.JSON(http.StatusOK, resp)
			return
		}

		proxy_items, status_code, err = utils.ParseRespProxy(e, proxy_resp)
		if err != nil {
			tools.Logger.Error("Parsing proxy response failed with: ", err)
			resp.SetError(err)
			c.JSON(http.StatusOK, resp)
			return
		}

		st.ShouldFallback = len(proxy_items) == 0 && st.CanFallback
	}

	err = utils.ConvertProxyItemsToC(&resp, e, car, &st, proxy_items, status_code)
	if err != nil {

		tools.Logger.Error("Converting to GAS response failed with: ", err)
		resp.SetError(err)
		c.JSON(http.StatusOK, resp)
		return
	}
	if json_data, erra := json.Marshal(resp); erra == nil {
		tools.Logger.Debugf("Response is %s", json_data)
	}
	c.JSON(http.StatusOK, resp)
}
