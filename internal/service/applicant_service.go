package service

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/model"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/repository"
	"github.com/lesmana/sekolah-karir-applicant-api/pkg/response"
	"gorm.io/gorm"
)

var (
	ErrApplicantNotFound  = errors.New("applicant not found")
	ErrEmailAlreadyExists = errors.New("applicant with this email already exists")
	ErrInvalidStage       = errors.New("invalid recruitment stage")
)

// ApplicantService defines business logic methods for applicants
type ApplicantService interface {
	CreateApplicant(ctx context.Context, req model.CreateApplicantRequest) (*model.Applicant, error)
	GetApplicantByID(ctx context.Context, id string) (*model.Applicant, error)
	ListApplicants(ctx context.Context, query model.ListApplicantsQuery) ([]model.Applicant, *response.PaginationMeta, error)
	UpdateApplicantStage(ctx context.Context, id string, req model.UpdateStageRequest) (*model.Applicant, error)
}

type applicantService struct {
	repo repository.ApplicantRepository
}

// NewApplicantService creates a new ApplicantService instance
func NewApplicantService(repo repository.ApplicantRepository) ApplicantService {
	return &applicantService{repo: repo}
}

func (s *applicantService) CreateApplicant(ctx context.Context, req model.CreateApplicantRequest) (*model.Applicant, error) {
	// Check email uniqueness
	existing, err := s.repo.FindByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, ErrEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	applicant := &model.Applicant{
		ID:              uuid.NewString(),
		Name:            strings.TrimSpace(req.Name),
		Email:           strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:           strings.TrimSpace(req.Phone),
		AppliedPosition: strings.TrimSpace(req.AppliedPosition),
		Stage:           model.StageApplied,
		Notes:           strings.TrimSpace(req.Notes),
	}

	if err := s.repo.Create(ctx, applicant); err != nil {
		return nil, err
	}

	return applicant, nil
}

func (s *applicantService) GetApplicantByID(ctx context.Context, id string) (*model.Applicant, error) {
	applicant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicantNotFound
		}
		return nil, err
	}
	return applicant, nil
}

func (s *applicantService) ListApplicants(ctx context.Context, query model.ListApplicantsQuery) ([]model.Applicant, *response.PaginationMeta, error) {
	// Sanitize pagination defaults
	if query.Page < 1 {
		query.Page = 1
	}
	if query.Limit < 1 {
		query.Limit = 10
	}
	if query.Limit > 100 {
		query.Limit = 100
	}

	// Validate stage filter if provided
	if query.Stage != "" && !model.IsValidStage(query.Stage) {
		return nil, nil, ErrInvalidStage
	}

	applicants, totalRecords, err := s.repo.FindAll(ctx, query)
	if err != nil {
		return nil, nil, err
	}

	totalPages := 0
	if totalRecords > 0 {
		totalPages = int(math.Ceil(float64(totalRecords) / float64(query.Limit)))
	}

	meta := &response.PaginationMeta{
		Page:         query.Page,
		Limit:        query.Limit,
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
	}

	return applicants, meta, nil
}

func (s *applicantService) UpdateApplicantStage(ctx context.Context, id string, req model.UpdateStageRequest) (*model.Applicant, error) {
	// Validate stage enum
	if !model.IsValidStage(req.Stage) {
		return nil, ErrInvalidStage
	}

	// Check if applicant exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicantNotFound
		}
		return nil, err
	}

	updated, err := s.repo.UpdateStage(ctx, id, req.Stage, strings.TrimSpace(req.Notes))
	if err != nil {
		return nil, err
	}

	return updated, nil
}

