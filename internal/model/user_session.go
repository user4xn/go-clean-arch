package model

import (
	"clean-arch/pkg/consts"
	"time"
)

type UserSession struct {
	ID               int                  `gorm:"primaryKey" json:"id"`
	UserID           int                  `gorm:"column:user_id" json:"user_id"`
	IPAddress        string               `gorm:"column:ip_address" json:"ip_address"`
	RefreshTokenHash string               `gorm:"column:refresh_token_hash" json:"refresh_token_hash"`
	Revoked          consts.SessionStatus `gorm:"column:revoked" json:"revoked"`
	ExpiresAt        time.Time            `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt        time.Time            `gorm:"column:created_at" json:"created_at"`
}

func (UserSession) TableName() string {
	return "user_sessions"
}
