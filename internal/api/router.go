package api

import (
	handlers "c2c/internal/api/handlers"
	"c2c/internal/tools"
	"fmt"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	ginlogrus "github.com/toorop/gin-logrus"
)

type Router struct {
	RouterGin *gin.Engine

	Routes Routes
	Env    *handlers.Env
}

func (r *Router) Initialize(env *handlers.Env) {

	r.Env = env

	r.Routes.Initialize(r.Env)

	r.GinInitialize()
}

func (r *Router) Run() error {
	err := r.GinRun()

	return err
}

func (r *Router) GinInitialize() {

	if tools.Logger.GetLevel() < logrus.WarnLevel {
		gin.SetMode(gin.ReleaseMode)
	}

	r.RouterGin = gin.New()

	if r.Env.Config.VersionPrefixedRoutes {
		r.RouterGin.Use(ginlogrus.Logger(tools.Logger, fmt.Sprintf("/v%s/healthz", MajorVersion), fmt.Sprintf("/v%s/health", MajorVersion)), gin.Recovery())
	} else {
		r.RouterGin.Use(ginlogrus.Logger(tools.Logger, "/healthz", "/health"), gin.Recovery())
	}

	if r.Env.Config.EnableDebugProfiling {
		pprof.Register(r.RouterGin)
	}

	r.GinRoutesBinding()
}

func (r *Router) GinRun() error {

	r.RouterGin.Use(cors.New(cors.Config{
		AllowOrigins: strings.Split(r.Env.Config.Cors.AllowedOrigins, ","),
		AllowMethods: strings.Split(r.Env.Config.Cors.AllowedMethods, ","),
		AllowHeaders: strings.Split(r.Env.Config.Cors.AllowedHeaders, ","),
	}))

	r.RouterGin.RedirectFixedPath = true
	return r.RouterGin.Run(r.Env.Config.Port)
}

func (r *Router) GinRoutesBinding() {

	group := r.RouterGin.Group(fmt.Sprintf("/v%s", MajorVersion))

	for _, route := range r.Routes.routes {
		if nil != route.GinHandlerFunc {
			if r.Env.Config.VersionPrefixedRoutes {
				group.Handle(route.Method, route.Pattern, route.GinHandlerFunc)
			} else {
				r.RouterGin.Handle(route.Method, route.Pattern, route.GinHandlerFunc)
			}
		}
	}

	// r.RouterGin.Static("/css", fmt.Sprintf("%s/%s",
	// 	r.Env.Config.Installpath,
	// 	"view/css"))
	// r.RouterGin.LoadHTMLFiles(fmt.Sprintf("%s/%s",
	// 	r.Env.Config.Installpath,
	// 	"view/html/404.html"))
	// r.RouterGin.NoRoute(func(c *gin.Context) {
	// 	c.HTML(404, "404.html", gin.H{})
	// })
}
