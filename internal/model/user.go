package model

import "time"

type User struct {
	ID              int        `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"column:name" json:"name"`
	Email           string     `gorm:"column:email" json:"email"`
	EmailVerifiedAt *time.Time `gorm:"column:email_verified_at" json:"email_verified_at"`
	Password        string     `gorm:"column:password" json:"password"`
	PhoneNumber     string     `gorm:"column:phone_number" json:"phone_number"`
	ProfileImageURL string     `gorm:"column:profile_image_url" json:"profile_image_url"`
	RememberToken   string     `gorm:"column:remember_token" json:"remember_token"`
	Common
}

func (User) TableName() string {
	return "users"
}
