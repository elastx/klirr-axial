package models

import "time"

type Base struct {
	ID string `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;not null"`
}

func (b* Base) BeforeCreate(ID string) {
	b.ID = ID
	if b.CreatedAt.IsZero() {
		b.CreatedAt = time.Now()
	}
}
