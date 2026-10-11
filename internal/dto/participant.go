package dto

import (
	"time"

	"splitobillo/internal/model"
)

type ParticipantRequest struct {
	Name string `json:"name"`
}

type ParticipantResponse struct {
	ID        uint      `json:"id"`
	ReceiptID uint      `json:"receipt_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAssignmentRequest struct {
	ReceiptItemID uint `json:"receipt_item_id"`
	ParticipantID uint `json:"participant_id"`
	Quantity      int  `json:"quantity"`
}

type AssignmentResponse struct {
	ID            uint      `json:"id"`
	ReceiptItemID uint      `json:"receipt_item_id"`
	ParticipantID uint      `json:"participant_id"`
	Quantity      int       `json:"quantity"`
	Amount        int64     `json:"amount"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewParticipantResponse(p *model.Participant) ParticipantResponse {
	return ParticipantResponse{
		ID:        p.ID,
		ReceiptID: p.ReceiptID,
		Name:      p.Name,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func NewParticipantListResponse(participants []model.Participant) []ParticipantResponse {
	out := make([]ParticipantResponse, 0, len(participants))
	for i := range participants {
		out = append(out, NewParticipantResponse(&participants[i]))
	}
	return out
}

func NewAssignmentResponse(a *model.ItemAssignment) AssignmentResponse {
	return AssignmentResponse{
		ID:            a.ID,
		ReceiptItemID: a.ReceiptItemID,
		ParticipantID: a.ParticipantID,
		Quantity:      a.Quantity,
		Amount:        a.Amount,
		CreatedAt:     a.CreatedAt,
	}
}

func NewAssignmentListResponse(assignments []model.ItemAssignment) []AssignmentResponse {
	out := make([]AssignmentResponse, 0, len(assignments))
	for i := range assignments {
		out = append(out, NewAssignmentResponse(&assignments[i]))
	}
	return out
}
