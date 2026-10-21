package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"splitobillo/internal/dto"
	"splitobillo/internal/middleware"
	"splitobillo/internal/service"
	"splitobillo/pkg/apperr"
	"splitobillo/pkg/response"
)

type ParticipantHandler struct {
	service *service.ParticipantService
}

func NewParticipantHandler(service *service.ParticipantService) *ParticipantHandler {
	return &ParticipantHandler{service: service}
}

func (h *ParticipantHandler) Add(c *gin.Context) {
	receiptID, ok := parseUintNamed(c, "id", "receipt")
	if !ok {
		return
	}
	var req dto.ParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest("invalid request body"))
		return
	}
	p, err := h.service.Add(c.GetString(middleware.CtxSessionID), receiptID, req.Name)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, dto.NewParticipantResponse(p))
}

func (h *ParticipantHandler) List(c *gin.Context) {
	receiptID, ok := parseUintNamed(c, "id", "receipt")
	if !ok {
		return
	}
	participants, err := h.service.List(c.GetString(middleware.CtxSessionID), receiptID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewParticipantListResponse(participants))
}

func (h *ParticipantHandler) Update(c *gin.Context) {
	receiptID, ok := parseUintNamed(c, "id", "receipt")
	if !ok {
		return
	}
	participantID, ok := parseUintNamed(c, "participantId", "participant")
	if !ok {
		return
	}
	var req dto.ParticipantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest("invalid request body"))
		return
	}
	p, err := h.service.Update(c.GetString(middleware.CtxSessionID), receiptID, participantID, req.Name)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewParticipantResponse(p))
}

func (h *ParticipantHandler) Delete(c *gin.Context) {
	receiptID, ok := parseUintNamed(c, "id", "receipt")
	if !ok {
		return
	}
	participantID, ok := parseUintNamed(c, "participantId", "participant")
	if !ok {
		return
	}
	if err := h.service.Delete(c.GetString(middleware.CtxSessionID), receiptID, participantID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *ParticipantHandler) Assign(c *gin.Context) {
	receiptID, ok := parseUintNamed(c, "id", "receipt")
	if !ok {
		return
	}
	var req dto.CreateAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest("invalid request body"))
		return
	}
	if req.ReceiptItemID == 0 || req.ParticipantID == 0 {
		response.Error(c, apperr.InvalidAssignment("receipt_item_id and participant_id are required"))
		return
	}
	assignment, err := h.service.Assign(
		c.GetString(middleware.CtxSessionID),
		receiptID, req.ReceiptItemID, req.ParticipantID, req.Quantity,
	)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, dto.NewAssignmentResponse(assignment))
}

func (h *ParticipantHandler) ListAssignments(c *gin.Context) {
	receiptID, ok := parseUintNamed(c, "id", "receipt")
	if !ok {
		return
	}
	assignments, err := h.service.ListAssignments(c.GetString(middleware.CtxSessionID), receiptID)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewAssignmentListResponse(assignments))
}

func (h *ParticipantHandler) RemoveAssignment(c *gin.Context) {
	receiptID, ok := parseUintNamed(c, "id", "receipt")
	if !ok {
		return
	}
	assignmentID, ok := parseUintNamed(c, "assignmentId", "assignment")
	if !ok {
		return
	}
	if err := h.service.RemoveAssignment(c.GetString(middleware.CtxSessionID), receiptID, assignmentID); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func parseUintNamed(c *gin.Context, name, label string) (uint, bool) {
	v, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || v == 0 {
		response.Error(c, apperr.BadRequest("invalid "+label+" id"))
		return 0, false
	}
	return uint(v), true
}
