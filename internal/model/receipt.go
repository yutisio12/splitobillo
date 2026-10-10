package model

import "time"

type ReceiptStatus string

const (
	StatusUploaded     ReceiptStatus = "UPLOADED"
	StatusProcessing   ReceiptStatus = "PROCESSING"
	StatusOCRCompleted ReceiptStatus = "OCR_COMPLETED"
	StatusReview       ReceiptStatus = "REVIEW"
	StatusSplitting    ReceiptStatus = "SPLITTING"
	StatusCompleted    ReceiptStatus = "COMPLETED"
	StatusOCRFailed    ReceiptStatus = "OCR_FAILED"
)

type Receipt struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	SessionID       string        `gorm:"size:64;index" json:"-"`
	ReceiptNumber   string        `gorm:"size:128;default:''" json:"receipt_number,omitempty"`
	MerchantName    string        `gorm:"size:255;default:''" json:"merchant_name,omitempty"`
	TransactionDate *time.Time    `json:"transaction_date,omitempty"`
	Subtotal        int64         `gorm:"default:0" json:"subtotal"`
	Tax             int64         `gorm:"default:0" json:"tax"`
	ServiceCharge   int64         `gorm:"default:0" json:"service_charge"`
	Discount        int64         `gorm:"default:0" json:"discount"`
	Total           int64         `gorm:"default:0" json:"total"`
	ImageURL        string        `gorm:"size:512;default:''" json:"image_url,omitempty"`
	OCRRawText      string        `gorm:"type:text" json:"-"`
	Status          ReceiptStatus `gorm:"size:32;default:'UPLOADED';index" json:"status"`
	Items           []ReceiptItem `gorm:"foreignKey:ReceiptID" json:"items,omitempty"`
	Participants    []Participant `gorm:"foreignKey:ReceiptID" json:"participants,omitempty"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}
