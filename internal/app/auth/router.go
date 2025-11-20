package auth

import (
	"clean-arch/internal/middleware"
	"clean-arch/pkg/util"

	"github.com/gin-gonic/gin"
)

func (h *handler) Secured(g *gin.RouterGroup) {
	enableCaptcha := util.GetEnv("ENABLE_HCAPTCHA", "false")

	if enableCaptcha == "true" {
		g.Use(middleware.HCaptcha())
	}
	g.POST("login", h.Login)
}

func (h *handler) Router(g *gin.RouterGroup) {
	g.POST("verify-email/:hash", h.VerifyEmail)
	g.POST("request-otp", h.RequestOTP)
	g.POST("verify-otp", h.VerifyOTP)
	g.POST("logout", h.Logout)
	g.POST("refresh", h.Refresh)
}
