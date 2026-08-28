package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"splitobillo/internal/config"
	"splitobillo/internal/model"
	"splitobillo/internal/ocrparser"
	"splitobillo/internal/repository"
	"splitobillo/pkg/apperr"
)

var allowedImageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

var receiptNumberPattern = regexp.MustCompile(`^[A-Za-z0-9\-_/]{0,128}$`)

type ReceiptService struct {
	repo      *repository.ReceiptRepository
	cfg       *config.Config
	extractor OCRService
}

func NewReceiptService(repo *repository.ReceiptRepository, cfg *config.Config, extractor OCRService) *ReceiptService {
	return &ReceiptService{repo: repo, cfg: cfg, extractor: extractor}
}

func (s *ReceiptService) Upload(sessionID string, fileHeader *multipart.FileHeader) (*model.Receipt, error) {
	if fileHeader == nil {
		return nil, apperr.BadRequest("file is required")
	}
	if fileHeader.Size > s.cfg.MaxUploadBytes {
		return nil, apperr.BadRequest(fmt.Sprintf("file too large, maximum %d bytes", s.cfg.MaxUploadBytes))
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedImageExt[ext] {
		return nil, apperr.BadRequest("allowed file types: jpg, jpeg, png, webp")
	}

	src, err := fileHeader.Open()
	if err != nil {
		return nil, apperr.BadRequest("unable to open uploaded file")
	}
	defer src.Close()

	if err := os.MkdirAll(s.cfg.UploadDir, 0o755); err != nil {
		return nil, apperr.Internal("failed to prepare upload directory")
	}

	fileName := fmt.Sprintf("%s%s", uuid.NewString(), ext)
	filePath := filepath.Join(s.cfg.UploadDir, fileName)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, apperr.Internal("failed to store uploaded file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		os.Remove(filePath)
		return nil, apperr.Internal("failed to write uploaded file")
	}

	receipt := &model.Receipt{
		SessionID: sessionID,
		ImageURL:  filePath,
		Status:    model.StatusUploaded,
	}
	if err := s.repo.Create(receipt); err != nil {
		os.Remove(filePath)
		return nil, apperr.Internal("failed to create receipt")
	}

	if err := s.RunOCR(receipt); err != nil {
		// Upload tetap sukses; status receipt menjadi OCR_FAILED dan dapat di-retry.
		return receipt, nil
	}
	return receipt, nil
}

// RunOCR memproses gambar receipt: PROCESSING -> OCR (parser) -> OCR_COMPLETED / OCR_FAILED.
func (s *ReceiptService) RunOCR(receipt *model.Receipt) error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := s.repo.SetStatus(receipt.ID, model.StatusProcessing); err != nil {
		return apperr.Internal("failed to update receipt status")
	}

	data, err := os.ReadFile(receipt.ImageURL)
	if err != nil {
		s.failOCR(receipt.ID, "image file no longer available")
		return apperr.Internal("image file no longer available")
	}

	rawText, err := s.extractor.ExtractReceipt(ctx, data)
	if err != nil {
		s.failOCR(receipt.ID, err.Error())
		return apperr.New("OCR_FAILED", err.Error(), http.StatusUnprocessableEntity)
	}

	parsed := ocrparser.Parse(rawText)

	items := make([]model.ReceiptItem, 0, len(parsed.Items))
	for _, p := range parsed.Items {
		items = append(items, model.ReceiptItem{
			ReceiptID:  receipt.ID,
			Name:       p.Name,
			Quantity:   p.Quantity,
			UnitPrice:  p.UnitPrice,
			TotalPrice: p.TotalPrice,
			Confidence: p.Confidence,
		})
	}

	receipt.OCRRawText = rawText
	receipt.Status = model.StatusOCRCompleted
	receipt.MerchantName = parsed.MerchantName
	receipt.Subtotal = parsed.Subtotal
	receipt.Tax = parsed.Tax
	receipt.ServiceCharge = parsed.ServiceCharge
	receipt.Discount = parsed.Discount

	var sumItems int64
	for i := range items {
		sumItems += items[i].TotalPrice * int64(items[i].Quantity)
	}
	receipt.Items = items
	if receipt.Total == 0 {
		if parsed.Total > 0 {
			receipt.Total = parsed.Total
		} else if sumItems > 0 {
			receipt.Total = sumItems + receipt.Tax + receipt.ServiceCharge - receipt.Discount
		}
	}
	if receipt.Subtotal == 0 && sumItems > 0 {
		receipt.Subtotal = sumItems
	}

	if err := s.replaceItemsAndSave(receipt); err != nil {
		s.failOCR(receipt.ID, err.Error())
		return apperr.Internal("failed to save OCR result")
	}
	return nil
}

func (s *ReceiptService) replaceItemsAndSave(receipt *model.Receipt) error {
	return s.repo.ReplaceItemsAndUpdateReceipt(receipt)
}

func (s *ReceiptService) failOCR(id uint, reason string) {
	_ = s.repo.SetStatusWithRawError(id, model.StatusOCRFailed, reason)
}

// Retry menjalankan ulang OCR untuk receipt yang sebelumnya gagal.
func (s *ReceiptService) Retry(sessionID string, id uint) (*model.Receipt, error) {
	receipt, err := s.repo.FindByID(sessionID, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, apperr.NotFound("receipt not found")
		}
		return nil, apperr.Internal("failed to get receipt")
	}

	if err := s.RunOCR(receipt); err != nil {
		return nil, err
	}
	return s.repo.FindByID(sessionID, id)
}

func (s *ReceiptService) Get(sessionID string, id uint) (*model.Receipt, error) {
	receipt, err := s.repo.FindByID(sessionID, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, apperr.NotFound("receipt not found")
		}
		return nil, apperr.Internal("failed to get receipt")
	}
	return receipt, nil
}

func (s *ReceiptService) List(sessionID string) ([]model.Receipt, error) {
	receipts, err := s.repo.List(sessionID)
	if err != nil {
		return nil, apperr.Internal("failed to list receipts")
	}
	return receipts, nil
}

func (s *ReceiptService) Update(sessionID string, id uint, req *UpdateReceiptInput) (*model.Receipt, error) {
	if req == nil {
		return nil, apperr.BadRequest("request body is required")
	}
	if req.ReceiptNumber != nil && !receiptNumberPattern.MatchString(*req.ReceiptNumber) {
		return nil, apperr.BadRequest("invalid receipt number")
	}

	receipt, err := s.repo.FindByID(sessionID, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, apperr.NotFound("receipt not found")
		}
		return nil, apperr.Internal("failed to get receipt")
	}

	applyUpdate(receipt, req)

	if receipt.Total < 0 || receipt.Subtotal < 0 || receipt.Tax < 0 ||
		receipt.ServiceCharge < 0 || receipt.Discount < 0 {
		return nil, apperr.Unprocessable("amount fields cannot be negative")
	}

	if len(req.Items) > 0 {
		if err := validateItems(req.Items); err != nil {
			return nil, err
		}
		items := make([]model.ReceiptItem, 0, len(req.Items))
		var newSubtotal int64
		for _, it := range req.Items {
			unit := it.UnitPrice
			total := it.TotalPrice
			if unit == 0 && total > 0 {
				unit = total / int64(maxInt(it.Quantity, 1))
			}
			if total == 0 && unit > 0 {
				total = unit * int64(maxInt(it.Quantity, 1))
			}
			newSubtotal += total
			items = append(items, model.ReceiptItem{
				ID:         it.ID,
				ReceiptID:  receipt.ID,
				Name:       strings.TrimSpace(it.Name),
				Quantity:   it.Quantity,
				UnitPrice:  unit,
				TotalPrice: total,
				Confidence: it.Confidence,
			})
		}
		receipt.Items = items
		receipt.Subtotal = newSubtotal
		receipt.Total = newSubtotal + receipt.Tax + receipt.ServiceCharge - receipt.Discount
		receipt.Status = model.StatusReview

		if err := s.repo.ReplaceItemsAndUpdateReceipt(receipt); err != nil {
			return nil, apperr.Internal("failed to save receipt items")
		}
		return s.repo.FindByID(sessionID, id)
	}

	if err := s.repo.Update(receipt); err != nil {
		return nil, apperr.Internal("failed to update receipt")
	}
	return receipt, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func validateItems(reqs []UpdateItemInput) *apperr.AppError {
	for _, it := range reqs {
		if strings.TrimSpace(it.Name) == "" {
			return apperr.Unprocessable("item name cannot be empty")
		}
		if it.Quantity < 1 {
			return apperr.Unprocessable("item quantity must be at least 1")
		}
		if it.UnitPrice < 0 || it.TotalPrice < 0 {
			return apperr.Unprocessable("item price cannot be negative")
		}
		if it.UnitPrice == 0 && it.TotalPrice == 0 {
			return apperr.Unprocessable("item must have a price")
		}
	}
	return nil
}

func applyUpdate(r *model.Receipt, req *UpdateReceiptInput) {
	if req.ReceiptNumber != nil {
		r.ReceiptNumber = *req.ReceiptNumber
	}
	if req.MerchantName != nil {
		r.MerchantName = *req.MerchantName
	}
	if req.TransactionDate != nil {
		t := *req.TransactionDate
		r.TransactionDate = &t
	}
	if req.Subtotal != nil {
		r.Subtotal = *req.Subtotal
	}
	if req.Tax != nil {
		r.Tax = *req.Tax
	}
	if req.ServiceCharge != nil {
		r.ServiceCharge = *req.ServiceCharge
	}
	if req.Discount != nil {
		r.Discount = *req.Discount
	}
	if req.Total != nil {
		r.Total = *req.Total
	}
}

type UpdateReceiptInput struct {
	ReceiptNumber   *string
	MerchantName    *string
	TransactionDate *time.Time
	Subtotal        *int64
	Tax             *int64
	ServiceCharge   *int64
	Discount        *int64
	Total           *int64
	Items           []UpdateItemInput
}

type UpdateItemInput struct {
	ID         uint
	Name       string
	Quantity   int
	UnitPrice  int64
	TotalPrice int64
	Confidence float64
}

func (s *ReceiptService) Delete(sessionID string, id uint) error {
	receipt, err := s.repo.FindByID(sessionID, id)
	if err != nil {
		if err == repository.ErrNotFound {
			return apperr.NotFound("receipt not found")
		}
		return apperr.Internal("failed to get receipt")
	}

	if err := s.repo.Delete(sessionID, id); err != nil {
		return apperr.Internal("failed to delete receipt")
	}

	if receipt.ImageURL != "" {
		os.Remove(receipt.ImageURL)
	}
	return nil
}
