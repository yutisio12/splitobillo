package model

import "time"

type ReceiptItem struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ReceiptID  uint      `gorm:"index" json:"receipt_id"`
	Name       string    `gorm:"size:255;default:''" json:"name"`
	Quantity   int       `gorm:"default:1" json:"quantity"`
	UnitPrice  int64     `gorm:"default:0" json:"unit_price"`
	TotalPrice int64     `gorm:"default:0" json:"total_price"`
	Confidence float64   `gorm:"default:0" json:"confidence"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
