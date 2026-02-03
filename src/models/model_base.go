package models

import (
	"time"

	"gorm.io/gorm"
)

type Base struct {
	ID string `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (b *Base) BeforeCreate(tx *gorm.DB) (err error) {
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
	return nil
}
