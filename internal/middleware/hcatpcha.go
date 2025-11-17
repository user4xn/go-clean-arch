package middleware

import (
	"bytes"
	"clean-arch/internal/dto"
	"clean-arch/pkg/util"
	"encoding/json"
	"io"

	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type hCaptchaResponse struct {
	Success bool `json:"success"`
}

func HCaptcha() gin.HandlerFunc {
	secret := util.GetEnv("HCAPTCHA_SECRET", "0x0000000000000000000000000000000000000000")

	return func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to read request body"})
			c.Abort()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var body dto.HCaptcha
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			response := util.APIResponse("hCaptcha internal error occured. Invalid JSON Payload", http.StatusBadRequest, "error", nil)
			c.JSON(http.StatusBadRequest, response)
			c.Abort()
			return
		}

		token := body.Response
		if token == "" {
			response := util.APIResponse("hCaptcha token not provided, are you robot ?", http.StatusBadRequest, "error", nil)
			c.JSON(http.StatusBadRequest, response)
			c.Abort()
			return
		}

		if !verifyHCaptcha(secret, token) {
			response := util.APIResponse("hCaptcha verification failed, are you robot ?", http.StatusForbidden, "error", nil)
			c.JSON(http.StatusForbidden, response)
			c.Abort()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		c.Next()
	}
}

func verifyHCaptcha(secret, token string) bool {
	data := url.Values{}
	data.Set("secret", secret)
	data.Set("response", token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post("https://hcaptcha.com/siteverify", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var hCaptchaResp hCaptchaResponse
	if err := json.NewDecoder(resp.Body).Decode(&hCaptchaResp); err != nil {
		return false
	}

	return hCaptchaResp.Success
}
