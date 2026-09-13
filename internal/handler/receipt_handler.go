package handler

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"splitobillo/internal/dto"
	"splitobillo/internal/middleware"
	"splitobillo/internal/service"
	"splitobillo/pkg/apperr"
	"splitobillo/pkg/response"
)

const maxMemoryForm = 8 << 20

type ReceiptHandler struct {
	service *service.ReceiptService
}

func NewReceiptHandler(service *service.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{service: service}
}

func (h *ReceiptHandler) Upload(c *gin.Context) {
	if !strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		response.Error(c, apperr.BadRequest("content-type must be multipart/form-data"))
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, apperr.BadRequest("multipart field 'file' is required"))
		return
	}

	receipt, err := h.service.Upload(c.GetString(middleware.CtxSessionID), fileHeader)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, dto.UploadReceiptResponse{ID: receipt.ID, Status: receipt.Status})
}

func (h *ReceiptHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c)
	if !ok {
		return
	}
	receipt, err := h.service.Get(c.GetString(middleware.CtxSessionID), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewReceiptResponse(receipt))
}

func (h *ReceiptHandler) List(c *gin.Context) {
	receipts, err := h.service.List(c.GetString(middleware.CtxSessionID))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewReceiptListResponse(receipts))
}

func (h *ReceiptHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c)
	if !ok {
		return
	}

	var req dto.UpdateReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperr.BadRequest("invalid request body: "+err.Error()))
		return
	}

	input := service.UpdateReceiptInput{
		ReceiptNumber:   req.ReceiptNumber,
		MerchantName:    req.MerchantName,
		TransactionDate: req.TransactionDate,
		Subtotal:        req.Subtotal,
		Tax:             req.Tax,
		ServiceCharge:   req.ServiceCharge,
		Discount:        req.Discount,
		Total:           req.Total,
	}
	if req.Items != nil {
		input.Items = make([]service.UpdateItemInput, 0, len(req.Items))
		for _, it := range req.Items {
			input.Items = append(input.Items, service.UpdateItemInput{
				ID:         it.ID,
				Name:       it.Name,
				Quantity:   it.Quantity,
				UnitPrice:  it.UnitPrice,
				TotalPrice: it.TotalPrice,
				Confidence: it.Confidence,
			})
		}
	}

	receipt, err := h.service.Update(c.GetString(middleware.CtxSessionID), id, &input)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewReceiptResponse(receipt))
}

func (h *ReceiptHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.GetString(middleware.CtxSessionID), id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

func (h *ReceiptHandler) RetryOCR(c *gin.Context) {
	id, ok := parseUintParam(c)
	if !ok {
		return
	}
	receipt, err := h.service.Retry(c.GetString(middleware.CtxSessionID), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.NewReceiptResponse(receipt))
}

func parseUintParam(c *gin.Context) (uint, bool) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id64 == 0 {
		response.Error(c, apperr.BadRequest("invalid receipt id"))
		return 0, false
	}
	return uint(id64), true
}
