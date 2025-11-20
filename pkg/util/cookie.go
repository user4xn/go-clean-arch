package util

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type CookieOptions struct {
	Name     string
	Value    string
	MaxAge   int
	Path     string
	Domain   string
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
}

func SetCookie(c *gin.Context, opts CookieOptions) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     opts.Name,
		Value:    opts.Value,
		Path:     opts.Path,
		Domain:   opts.Domain,
		Expires:  time.Now().Add(time.Duration(opts.MaxAge) * time.Second),
		MaxAge:   opts.MaxAge,
		Secure:   opts.Secure,
		HttpOnly: opts.HttpOnly,
		SameSite: opts.SameSite,
	})
}

func SetRefreshTokenCookie(c *gin.Context, token string, maxDays int) {
	maxAge := 60 * 60 * 24 * maxDays
	secure := false
	JWTMode := GetEnv("JWT_MODE", "fallback")
	if JWTMode == "release" {
		secure = true
	}

	SetCookie(c, CookieOptions{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   maxAge,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
}
