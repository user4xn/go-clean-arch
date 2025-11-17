package helper

import (
	"clean-arch/pkg/util"

	"github.com/gin-gonic/gin"
)

func Index(g *gin.Engine) {
	appName := util.GetEnv("APP_NAME", "Clean Arch")
	g.GET("/", func(context *gin.Context) {
		context.JSON(200, struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}{
			Name:    appName,
			Version: util.GetEnv("APP_VERSION", "1.0"),
		})
	})
}
