package repository

import (
	"errors"

	"gorm.io/gorm"
	"splitobillo/internal/model"
)

type ParticipantRepository struct {
	db *gorm.DB
}

func NewParticipantRepository(db *gorm.DB) *ParticipantRepository {
	return &ParticipantRepository{db: db}
}

func (r *ParticipantRepository) Create(p *model.Participant) error {
	return r.db.Create(p).Error
}

func (r *ParticipantRepository) ListByReceipt(receiptID uint) ([]model.Participant, error) {
	var participants []model.Participant
	err := r.db.
		Where("receipt_id = ?", receiptID).
		Order("id ASC").
		Find(&participants).Error
	return participants, err
}

func (r *ParticipantRepository) FindByID(receiptID, id uint) (*model.Participant, error) {
	var p model.Participant
	err := r.db.Where("id = ? AND receipt_id = ?", id, receiptID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ParticipantRepository) UpdateName(receiptID, id uint, name string) error {
	res := r.db.Model(&model.Participant{}).
		Where("id = ? AND receipt_id = ?", id, receiptID).
		Update("name", name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ParticipantRepository) Delete(receiptID, id uint) error {
	res := r.db.Where("id = ? AND receipt_id = ?", id, receiptID).Delete(&model.Participant{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ParticipantRepository) FindAssignmentByItemAndParticipant(itemID, participantID uint) (*model.ItemAssignment, error) {
	var a model.ItemAssignment
	err := r.db.Where("receipt_item_id = ? AND participant_id = ?", itemID, participantID).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ParticipantRepository) CreateAssignment(a *model.ItemAssignment) error {
	return r.db.Create(a).Error
}

func (r *ParticipantRepository) UpdateAssignment(a *model.ItemAssignment) error {
	return r.db.Save(a).Error
}

// FindAssignmentScope memastikan assignment berada di receipt yang benar dengan
// men-join ke receipt_items.
func (r *ParticipantRepository) FindAssignmentScope(receiptID, assignmentID uint) (*model.ItemAssignment, error) {
	var a model.ItemAssignment
	err := r.db.
		Joins("JOIN receipt_items ON receipt_items.id = item_assignments.receipt_item_id").
		Where("item_assignments.id = ? AND receipt_items.receipt_id = ?", assignmentID, receiptID).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ParticipantRepository) DeleteAssignment(receiptID, assignmentID uint) error {
	sub := r.db.Model(&model.ReceiptItem{}).Select("id").Where("receipt_id = ?", receiptID)
	res := r.db.Where("id = ? AND receipt_item_id IN (?)", assignmentID, sub).Delete(&model.ItemAssignment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ParticipantRepository) ListAssignmentsByItem(itemID uint) ([]model.ItemAssignment, error) {
	var assignments []model.ItemAssignment
	err := r.db.
		Where("receipt_item_id = ?", itemID).
		Order("id ASC").
		Find(&assignments).Error
	return assignments, err
}

func (r *ParticipantRepository) ListAssignmentsByReceipt(receiptID uint) ([]model.ItemAssignment, error) {
	var assignments []model.ItemAssignment
	err := r.db.
		Joins("JOIN receipt_items ON receipt_items.id = item_assignments.receipt_item_id").
		Where("receipt_items.receipt_id = ?", receiptID).
		Order("item_assignments.id ASC").
		Find(&assignments).Error
	return assignments, err
}

// SaveAssignmentAmounts menyimpan ulang amount pada tiap assignment dalam transaksi.
func (r *ParticipantRepository) SaveAssignmentAmounts(assignments []model.ItemAssignment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for i := range assignments {
			if err := tx.Model(&model.ItemAssignment{}).
				Where("id = ?", assignments[i].ID).
				Update("amount", assignments[i].Amount).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
