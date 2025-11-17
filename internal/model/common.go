package model

import (
	"time"

	"gorm.io/gorm"
)

// Define common table collumns
type Common struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

type CreatedOnly struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

// Create default data for created_at and updated_at
func (c *Common) BeforeCreate(tx *gorm.DB) (err error) {
	now := time.Now()
	c.CreatedAt = now
	c.UpdatedAt = now
	return
}

// Create default data for updated_at if update happens to the row
func (c *Common) BeforeUpdate(tx *gorm.DB) (err error) {
	c.UpdatedAt = time.Now()
	return
}
