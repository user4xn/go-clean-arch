package auth

import (
	"clean-arch/internal/dto"
	"clean-arch/internal/factory"
	"clean-arch/pkg/config"
	"clean-arch/pkg/consts"
	"clean-arch/pkg/util"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type handler struct {
	service Service
}

func NewHandler(f *factory.Factory) *handler {
	return &handler{
		service: NewService(f),
	}
}

func (h *handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || refreshToken == "" {
		response := util.APIResponse("refresh token not found", http.StatusUnauthorized, "failed", nil)
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	fmt.Printf("cookie found %s", refreshToken)

	clientIP := c.ClientIP()

	newTokens, newRefreshToken, err := h.service.Refresh(c, refreshToken, clientIP)
	if err != nil {
		response := util.APIResponse(fmt.Sprintf("refresh failed: %s", err.Error()), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	util.SetRefreshTokenCookie(c, *newRefreshToken, config.GetRefreshDuration())

	response := util.APIResponse("refresh success", http.StatusOK, "success", newTokens)
	c.JSON(http.StatusOK, response)
}

func (h *handler) Logout(c *gin.Context) {
	header := c.Request.Header["Authorization"]
	rep := regexp.MustCompile(`(Bearer)\s?`)
	bearerStr := rep.ReplaceAllString(header[0], "")

	err := h.service.Logout(c, bearerStr)
	if err != nil {
		response := util.APIResponse(fmt.Sprintf("logout failed otp %s", err.Error()), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := util.APIResponse("logout successfull", http.StatusOK, "success", nil)
	c.JSON(http.StatusOK, response)
}

func (h *handler) VerifyOTP(c *gin.Context) {
	var body dto.PayloadVerifyOtp

	err := c.ShouldBind(&body)
	if err != nil {
		response := util.APIResponse("verify otp failed", http.StatusUnprocessableEntity, "failed", err.Error())
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	err = validation.ValidateStruct(&body,
		validation.Field(&body.Email,
			validation.Required,
		),
		validation.Field(&body.OTP,
			validation.Required,
		),
	)
	if err != nil {
		response := util.APIResponse("verify otp failed", http.StatusUnprocessableEntity, "failed", err.Error())
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	bodyUpdate := dto.PayloadVerifyOtpTraced{
		Email:     body.Email,
		OTP:       body.OTP,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	res, refreshToken, err := h.service.VerifyOTP(c, bodyUpdate)
	if err != nil {
		response := util.APIResponse(fmt.Sprintf("verify otp failed otp %s", err.Error()), http.StatusBadRequest, "failed", res)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	util.SetRefreshTokenCookie(c, *refreshToken, config.GetRefreshDuration())

	response := util.APIResponse("verify otp successfull", http.StatusOK, "success", res)
	c.JSON(http.StatusOK, response)
}

func (h *handler) RequestOTP(c *gin.Context) {
	var body dto.PayloadOtp

	err := c.ShouldBind(&body)
	if err != nil {
		response := util.APIResponse("request otp failed", http.StatusUnprocessableEntity, "failed", err.Error())
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	err = validation.ValidateStruct(&body,
		validation.Field(&body.Email,
			validation.Required,
		),
	)

	if err != nil {
		response := util.APIResponse("request otp failed", http.StatusUnprocessableEntity, "failed", err.Error())
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	res, err := h.service.RequestOTP(c, body)
	if err != nil {
		response := util.APIResponse(fmt.Sprintf("request otp failed otp %s", err.Error()), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := util.APIResponse("request otp successfull", http.StatusOK, "success", res)
	c.JSON(http.StatusOK, response)
}

func (h *handler) VerifyEmail(c *gin.Context) {
	base64String := c.Param("hash")

	if base64String == "" {
		response := util.APIResponse("data not valid", http.StatusUnprocessableEntity, "failed", nil)
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	err := h.service.VerifyEmail(c, base64String)
	if err != nil {
		response := util.APIResponse(fmt.Sprintf("failed to verify email %s", err.Error()), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := util.APIResponse("successfully verify email", http.StatusOK, "success", nil)
	c.JSON(http.StatusOK, response)
}

// Login godoc
// @Summary Login user
// @Description Login using email & password, record IP & User-Agent, and generate JWT session.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.PayloadLogin true "Login payload"
// @Success 200 {object} util.Response "Login success"
// @Failure 400 {object} util.Response "Invalid password or failed login"
// @Failure 422 {object} util.Response "Validation error"
// @Router /auth/login [post]
func (h *handler) Login(c *gin.Context) {
	var body dto.PayloadLogin
	if err := c.ShouldBind(&body); err != nil {
		errorMessage := gin.H{"errors": "please fill data"}
		if err != io.EOF {
			errors := util.FormatValidationError(err)
			errorMessage = gin.H{"errors": errors}
		}
		response := util.APIResponse("Failed Login", http.StatusUnprocessableEntity, "failed", errorMessage)
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	err := validation.ValidateStruct(&body,
		validation.Field(&body.Email,
			validation.Required,
		),
		validation.Field(&body.Password,
			validation.Required,
		),
	)

	if err != nil {
		response := util.APIResponse("login failed", http.StatusUnprocessableEntity, "failed", err.Error())
		c.JSON(http.StatusUnprocessableEntity, response)
		return
	}

	bodyUpdate := dto.PayloadLoginTraced{
		Email:     body.Email,
		Password:  body.Password,
		IP:        c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
	}

	data, refreshToken, err := h.service.LoginAttempt(c, bodyUpdate)
	if err == consts.UserNotFound {
		response := util.APIResponse(fmt.Sprintf("%s", consts.UserNotFound), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err == consts.Required2FA {
		response := util.APIResponse(fmt.Sprintf("%s", consts.Required2FA), http.StatusOK, "success", nil)
		c.JSON(http.StatusOK, response)
		return
	}

	if err == consts.InvalidPassword {
		response := util.APIResponse(fmt.Sprintf("%s", consts.InvalidPassword), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err == consts.ErrorLoadLocationTime {
		response := util.APIResponse(fmt.Sprintf("%s", consts.ErrorLoadLocationTime), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err == consts.ErrorGenerateJwt {
		response := util.APIResponse(fmt.Sprintf("%s", consts.ErrorGenerateJwt), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err == consts.UserNotVerifyEmail {
		response := util.APIResponse(fmt.Sprintf("%s", consts.UserNotVerifyEmail), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if err == consts.EmptyGenerateJwt {
		response := util.APIResponse(fmt.Sprintf("%s", consts.EmptyGenerateJwt), http.StatusBadRequest, "failed", nil)
		c.JSON(http.StatusBadRequest, response)
		return
	}

	util.SetRefreshTokenCookie(c, *refreshToken, config.GetRefreshDuration())

	response := util.APIResponse("Success Login", http.StatusOK, "success", data)
	c.JSON(http.StatusOK, response)
}
