package http

import (
	"clean-arch/internal/app/user"
	"clean-arch/internal/factory"
	"clean-arch/internal/middleware"
	"clean-arch/pkg/config"
	"clean-arch/pkg/helper"
	"clean-arch/pkg/tracer"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Here we define route function for user Handlers that accepts gin.Engine and factory parameters
func NewHttp(g *gin.Engine, f *factory.Factory) {
	logger, err := tracer.InitLogger(strings.ToLower(config.AppEnv()))
	if err != nil {
		panic(err)
	}
	if logger == nil {
		panic(err)
	}

	defer logger.Sync()

	helper.Index(g)

	// Here we use cors middleware
	g.Use(middleware.CORSMiddleware())

	// Here we use logger middleware before the actual API to catch any api call from clients
	g.Use(tracer.LoggingMiddleware(logger))

	// Here we use the recovery middleware to catch a panic, if panic occurs recover the application witohut shutting it off
	g.Use(tracer.RecoverMiddleware(logger))

	g.Use(gin.Logger())
	g.Use(gin.Recovery())

	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Here we define a router group
	v1 := g.Group("/api/v1")
	// Here we register the route from user handler
	user.NewHandler(f).Router(v1.Group("/user"))
}
