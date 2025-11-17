package dto

import (
	"mime/multipart"
	"time"
)

type (
	PayloadUser struct {
		ID              int                   `form:"id"`
		Name            string                `form:"name" binding:"required"`
		Email           string                `form:"email" binding:"required"`
		EmailVerifiedAt *time.Time            `form:"email_verified_at"`
		Password        string                `form:"password" binding:"required"`
		PhoneNumber     string                `form:"phone_number"`
		RememberToken   string                `form:"remember_token"`
		File            *multipart.FileHeader `form:"file"`
		URL             string                `form:"url"`
	}

	PayloadUpdateUser struct {
		ID           int                   `form:"id"`
		Name         string                `form:"name"`
		Email        string                `form:"email"`
		File         *multipart.FileHeader `form:"file"`
		URL          string                `form:"url"`
		NewPassword  string                `form:"new_password"`
		LastPassword string                `form:"last_password"`
		PhoneNumber  string                `form:"phone_number"`
	}

	User struct {
		ID              int     `json:"id"`
		Name            string  `json:"name" binding:"required"`
		Email           string  `json:"email" binding:"required"`
		EmailVerifiedAt *string `json:"email_verified_at"`
		ProfileImageURL string  `json:"profile_image_url"`
		PhoneNumber     string  `json:"phone_number"`
		CreatedAt       string  `json:"created_at"`
		UpdatedAt       string  `json:"updated_at"`
	}

	ResponseUser struct {
		Data []User `json:"data"`
		ResponseTotalRow
	}
)
