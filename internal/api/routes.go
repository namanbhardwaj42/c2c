//routes.go

package api

import (
	// "fmt"
	"time"

	"net/http"

	"github.com/gin-gonic/gin"

	endpoints "c2c/internal/api/endpoints"
	handlers "c2c/internal/api/handlers"
	utils "c2c/internal/api/utils"
	"c2c/internal/models"
	"c2c/internal/tools"

	"github.com/gin-contrib/timeout"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
)

type Route struct {
	Name           string
	Method         string
	Pattern        string
	GinHandlerFunc gin.HandlerFunc
}

type Routes struct {
	routes []Route
	Env    *handlers.Env
}

var lAllRoutes []string = []string{"GET", "POST", "PUT", "DELETE"}
var lGetRoute []string = []string{"GET"}
var lPostRoute []string = []string{"POST", "PUT"}
var lGetPostRoutes []string = []string{"GET", "POST", "PUT"}

func (r *Routes) addRoute(routetypes []string, name string, pattern string, ginhandler gin.HandlerFunc) {
	for _, rtype := range routetypes {
		if utils.Contains(lAllRoutes, rtype) {
			r.routes = append(r.routes, Route{name, rtype, pattern, ginhandler})
		}
	}
}
func timeoutResponse(c *gin.Context) {
	tools.Logger.Error("Search fall into the rabbit hole, timeout")
	var resp models.GasResponse
	err := models.C2CError{ErrorCode: "RESPONSE_CODE_UNSPECIFIED"}
	resp.SetError(&err)
	c.JSON(http.StatusOK, resp)
}

func (r *Routes) makeRouteList() {

	docs.SwaggerInfo.BasePath = r.Env.Config.Basepath
	r.addRoute(lGetRoute, "Swagger", "/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Default routes
	r.addRoute(lGetRoute, "PingCall", "ping", endpoints.HandlePing(r.Env))
	r.addRoute(lGetRoute, "Version", "version", endpoints.HandleGetVersion(r.Env))

	// Check health
	r.addRoute(lGetRoute, "HealthCall", "health", endpoints.HandleGetHealth(r.Env))
	r.addRoute(lGetRoute, "HealthCall", "healthz", endpoints.HandleGetHealth(r.Env))

	// Check proxy
	r.addRoute(lGetPostRoutes,
		"ProxyHealthCall",
		"proxyhealth",
		timeout.New(
			timeout.WithTimeout(time.Duration(r.Env.Config.Assistant.SearchConfig.Timeout)*time.Millisecond),
			timeout.WithHandler(utils.IsAuthorized(endpoints.HandleGetProxyHealth, r.Env, lGetRoute)),
			timeout.WithResponse(timeoutResponse)))

	// Search
	// r.addRoute(lGetPostRoutes, "SearchCall", "search", utils.IsAuthorized(endpoints.HandleSearchCall, r.Env))
	r.addRoute(lGetPostRoutes,
		"SearchCall",
		"search",
		timeout.New(
			timeout.WithTimeout(time.Duration(r.Env.Config.Assistant.SearchConfig.Timeout)*time.Millisecond),
			timeout.WithHandler(utils.IsAuthorized(endpoints.HandleSearchCall, r.Env, lGetPostRoutes)),
			timeout.WithResponse(timeoutResponse)))

	// Config
	r.addRoute(lGetPostRoutes, "AssConfigCall", "assconfig", utils.IsAuthorized(endpoints.HandleAssConfigCall, r.Env, lPostRoute))
}

func (r *Routes) Initialize(env *handlers.Env) {

	r.Env = env

	r.makeRouteList()
}
