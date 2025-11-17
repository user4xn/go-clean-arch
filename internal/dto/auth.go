package dto

import "time"

type (
	PayloadLogin struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	PayloadLoginTraced struct {
		Email     string `json:"email" binding:"required"`
		Password  string `json:"password" binding:"required"`
		IP        string `json:"ip"`
		UserAgent string `json:"user_agent"`
	}

	ResponseJWT struct {
		TokenJwt  string         `json:"token_jwt"`
		ExpiredAt string         `json:"expired_at"`
		DataUser  *DataUserLogin `json:"data_user"`
	}

	DataUserLogin struct {
		ID              int       `json:"id"`
		Email           string    `json:"email"`
		Name            string    `json:"name"`
		EmailVerifiedAt time.Time `json:"email_verify_at"`
		ProfileImageURL string    `json:"profile_image_url"`
	}

	HCaptcha struct {
		Response string `json:"h-captcha-response"`
	}
)
