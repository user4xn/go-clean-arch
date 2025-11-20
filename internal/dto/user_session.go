package dto

import (
	"clean-arch/pkg/consts"
	"time"
)

type (
	UserSession struct {
		UserID           int                  `json:"user_id"`
		IPAddress        string               `json:"ip_address"`
		RefreshTokenHash string               `json:"refresh_token_hash"`
		ExpiresAt        string               `json:"expires_at"`
		Revoked          consts.SessionStatus `json:"revoked"`
	}

	JwtSession struct {
		ID              int        `json:"id"`
		Name            string     `json:"name"`
		Email           string     `json:"email"`
		EmailVerifiedAt *time.Time `json:"email_verified_at"`
		PhoneNumber     string     `json:"phone_number"`
		CreatedAt       time.Time  `json:"created_at"`
	}
)
