package repository

import (
	"errors"

	"gorm.io/gorm"
	"splitobillo/internal/model"
)

var ErrNotFound = errors.New("record not found")

type ReceiptRepository struct {
	db *gorm.DB
}

func NewReceiptRepository(db *gorm.DB) *ReceiptRepository {
	return &ReceiptRepository{db: db}
}

func (r *ReceiptRepository) Create(receipt *model.Receipt) error {
	return r.db.Create(receipt).Error
}

func (r *ReceiptRepository) FindByID(sessionID string, id uint) (*model.Receipt, error) {
	var receipt model.Receipt
	err := r.db.
		Preload("Items").
		Where("id = ? AND session_id = ?", id, sessionID).
		First(&receipt).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &receipt, nil
}

func (r *ReceiptRepository) List(sessionID string) ([]model.Receipt, error) {
	var receipts []model.Receipt
	err := r.db.
		Preload("Items").
		Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Find(&receipts).Error
	return receipts, err
}

func (r *ReceiptRepository) Update(receipt *model.Receipt) error {
	return r.db.Save(receipt).Error
}

func (r *ReceiptRepository) Delete(sessionID string, id uint) error {
	res := r.db.Where("id = ? AND session_id = ?", id, sessionID).Delete(&model.Receipt{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ReceiptRepository) SetStatus(id uint, status model.ReceiptStatus) error {
	return r.db.Model(&model.Receipt{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ReceiptRepository) SetStatusWithRawError(id uint, status model.ReceiptStatus, rawText string) error {
	updates := map[string]any{"status": status}
	if rawText != "" {
		updates["ocr_raw_text"] = rawText
	}
	return r.db.Model(&model.Receipt{}).Where("id = ?", id).Updates(updates).Error
}

// ReplaceItemsAndUpdateReceipt menyimpan ulang items hasil OCR dan update receipt
// dalam satu transaksi.
func (r *ReceiptRepository) ReplaceItemsAndUpdateReceipt(receipt *model.Receipt) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("receipt_id = ?", receipt.ID).Delete(&model.ReceiptItem{}).Error; err != nil {
			return err
		}
		for i := range receipt.Items {
			receipt.Items[i].ID = 0
			if err := tx.Create(&receipt.Items[i]).Error; err != nil {
				return err
			}
		}
		return tx.Omit("Items").Save(receipt).Error
	})
}
