package dto

import (
	"time"

	"splitobillo/internal/model"
)

type ReceiptResponse struct {
	ID              uint                `json:"id"`
	ReceiptNumber   string              `json:"receipt_number,omitempty"`
	MerchantName    string              `json:"merchant_name,omitempty"`
	TransactionDate *time.Time          `json:"transaction_date,omitempty"`
	Subtotal        int64               `json:"subtotal"`
	Tax             int64               `json:"tax"`
	ServiceCharge   int64               `json:"service_charge"`
	Discount        int64               `json:"discount"`
	Total           int64               `json:"total"`
	Status          model.ReceiptStatus `json:"status"`
	ImageURL        string              `json:"image_url,omitempty"`
	Items           []ItemResponse      `json:"items,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type ItemResponse struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	UnitPrice  int64   `json:"unit_price"`
	TotalPrice int64   `json:"total_price"`
	Confidence float64 `json:"confidence"`
}

type UploadReceiptResponse struct {
	ID     uint                `json:"id"`
	Status model.ReceiptStatus `json:"status"`
}

type UpdateReceiptRequest struct {
	ReceiptNumber   *string             `json:"receipt_number"`
	MerchantName    *string             `json:"merchant_name"`
	TransactionDate *time.Time          `json:"transaction_date"`
	Subtotal        *int64              `json:"subtotal"`
	Tax             *int64              `json:"tax"`
	ServiceCharge   *int64              `json:"service_charge"`
	Discount        *int64              `json:"discount"`
	Total           *int64              `json:"total"`
	Items           []UpdateItemRequest `json:"items"`
}

type UpdateItemRequest struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Quantity   int     `json:"quantity"`
	UnitPrice  int64   `json:"unit_price"`
	TotalPrice int64   `json:"total_price"`
	Confidence float64 `json:"confidence"`
}

func NewReceiptResponse(r *model.Receipt) *ReceiptResponse {
	resp := &ReceiptResponse{
		ID:              r.ID,
		ReceiptNumber:   r.ReceiptNumber,
		MerchantName:    r.MerchantName,
		TransactionDate: r.TransactionDate,
		Subtotal:        r.Subtotal,
		Tax:             r.Tax,
		ServiceCharge:   r.ServiceCharge,
		Discount:        r.Discount,
		Total:           r.Total,
		Status:          r.Status,
		ImageURL:        r.ImageURL,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
	if len(r.Items) > 0 {
		resp.Items = make([]ItemResponse, 0, len(r.Items))
		for i := range r.Items {
			item := &r.Items[i]
			resp.Items = append(resp.Items, ItemResponse{
				ID:         item.ID,
				Name:       item.Name,
				Quantity:   item.Quantity,
				UnitPrice:  item.UnitPrice,
				TotalPrice: item.TotalPrice,
				Confidence: item.Confidence,
			})
		}
	}
	return resp
}

func NewReceiptListResponse(receipts []model.Receipt) []ReceiptResponse {
	out := make([]ReceiptResponse, 0, len(receipts))
	for i := range receipts {
		out = append(out, *NewReceiptResponse(&receipts[i]))
	}
	return out
}
