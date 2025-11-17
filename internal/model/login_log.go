package model

type LoginLog struct {
	ID        int    `gorm:"primaryKey" json:"id"`
	UserID    int    `gorm:"column:user_id" json:"user_id"`
	IPAddress string `gorm:"column:ip_address" json:"ip_address"`
	UserAgent string `gorm:"column:user_agent" json:"user_agent"`
	Common
}

func (LoginLog) TableName() string {
	return "login_logs"
}
