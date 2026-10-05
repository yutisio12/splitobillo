package model

import "time"

type ItemAssignment struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	ReceiptItemID uint      `gorm:"index" json:"receipt_item_id"`
	ParticipantID uint      `gorm:"index" json:"participant_id"`
	Quantity      int       `gorm:"default:1" json:"quantity"`
	Amount        int64     `gorm:"default:0" json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
