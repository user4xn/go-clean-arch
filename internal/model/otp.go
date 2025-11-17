package model

import (
	"time"
)

type OTP struct {
	ID            int       `gorm:"primaryKey" json:"id"`
	UserID        int       `gorm:"column:user_id" json:"user_id"`
	OTP           string    `gorm:"column:otp" json:"otp"`
	Attempt       int       `gorm:"column:attempt" json:"attempt"`
	NextRequestAt time.Time `gorm:"column:next_request_at" json:"next_request_at"`
	ExpiredAt     time.Time `gorm:"column:expired_at" json:"expired_at"`
	Common
}

func (OTP) TableName() string {
	return "otps"
}
