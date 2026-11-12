package repository

import (
	"context"
	"strings"

	"github.com/lesmana/sekolah-karir-applicant-api/internal/model"
	"gorm.io/gorm"
)

// ApplicantRepository defines data access methods for applicants
type ApplicantRepository interface {
	Create(ctx context.Context, applicant *model.Applicant) error
	FindByID(ctx context.Context, id string) (*model.Applicant, error)
	FindByEmail(ctx context.Context, email string) (*model.Applicant, error)
	FindAll(ctx context.Context, query model.ListApplicantsQuery) ([]model.Applicant, int64, error)
	UpdateStage(ctx context.Context, id string, stage string, notes string) (*model.Applicant, error)
}

type applicantRepository struct {
	db *gorm.DB
}

// NewApplicantRepository creates a new ApplicantRepository instance
func NewApplicantRepository(db *gorm.DB) ApplicantRepository {
	return &applicantRepository{db: db}
}

func (r *applicantRepository) Create(ctx context.Context, applicant *model.Applicant) error {
	return r.db.WithContext(ctx).Create(applicant).Error
}

func (r *applicantRepository) FindByID(ctx context.Context, id string) (*model.Applicant, error) {
	var applicant model.Applicant
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&applicant).Error
	if err != nil {
		return nil, err
	}
	return &applicant, nil
}

func (r *applicantRepository) FindByEmail(ctx context.Context, email string) (*model.Applicant, error) {
	var applicant model.Applicant
	err := r.db.WithContext(ctx).Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(email))).First(&applicant).Error
	if err != nil {
		return nil, err
	}
	return &applicant, nil
}

func (r *applicantRepository) FindAll(ctx context.Context, query model.ListApplicantsQuery) ([]model.Applicant, int64, error) {
	var applicants []model.Applicant
	var totalRecords int64

	tx := r.db.WithContext(ctx).Model(&model.Applicant{})

	// Filter by stage
	if query.Stage != "" {
		tx = tx.Where("stage = ?", strings.ToUpper(strings.TrimSpace(query.Stage)))
	}

	// Filter by position
	if query.Position != "" {
		tx = tx.Where("LOWER(applied_position) LIKE ?", "%"+strings.ToLower(strings.TrimSpace(query.Position))+"%")
	}

	// Keyword search on name or email
	if query.Search != "" {
		searchPattern := "%" + strings.ToLower(strings.TrimSpace(query.Search)) + "%"
		tx = tx.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", searchPattern, searchPattern)
	}

	// Count total records matching filters
	if err := tx.Count(&totalRecords).Error; err != nil {
		return nil, 0, err
	}

	// Pagination
	offset := (query.Page - 1) * query.Limit
	err := tx.Order("created_at DESC").Offset(offset).Limit(query.Limit).Find(&applicants).Error
	if err != nil {
		return nil, 0, err
	}

	return applicants, totalRecords, nil
}

func (r *applicantRepository) UpdateStage(ctx context.Context, id string, stage string, notes string) (*model.Applicant, error) {
	var applicant model.Applicant
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&applicant).Error
	if err != nil {
		return nil, err
	}

	updates := map[string]interface{}{
		"stage": strings.ToUpper(strings.TrimSpace(stage)),
	}
	if notes != "" {
		updates["notes"] = notes
	}

	if err := r.db.WithContext(ctx).Model(&applicant).Updates(updates).Error; err != nil {
		return nil, err
	}

	// Reload updated fields
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&applicant).Error; err != nil {
		return nil, err
	}

	return &applicant, nil
}

