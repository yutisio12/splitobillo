package model

import "time"

type Participant struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ReceiptID uint      `gorm:"index" json:"receipt_id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
