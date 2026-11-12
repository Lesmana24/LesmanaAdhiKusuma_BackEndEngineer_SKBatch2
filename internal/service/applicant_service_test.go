package service_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/model"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/repository"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) (*gorm.DB, service.ApplicantService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.Applicant{})
	require.NoError(t, err)

	// Clean table
	db.Exec("DELETE FROM applicants")

	repo := repository.NewApplicantRepository(db)
	svc := service.NewApplicantService(repo)

	return db, svc
}

func TestApplicantService_CreateApplicant(t *testing.T) {
	_, svc := setupTestDB(t)
	ctx := context.Background()

	t.Run("success create applicant", func(t *testing.T) {
		req := model.CreateApplicantRequest{
			Name:            "Lesmana Adh Kusuma",
			Email:           "lesmana@example.com",
			Phone:           "081234567890",
			AppliedPosition: "Back-End Engineer",
			Notes:           "Experienced Go developer",
		}

		applicant, err := svc.CreateApplicant(ctx, req)
		require.NoError(t, err)
		assert.NotEmpty(t, applicant.ID)
		assert.Equal(t, "Lesmana Adh Kusuma", applicant.Name)
		assert.Equal(t, "lesmana@example.com", applicant.Email)
		assert.Equal(t, "081234567890", applicant.Phone)
		assert.Equal(t, "Back-End Engineer", applicant.AppliedPosition)
		assert.Equal(t, model.StageApplied, applicant.Stage)
		assert.Equal(t, "Experienced Go developer", applicant.Notes)
		assert.False(t, applicant.CreatedAt.IsZero())
	})

	t.Run("duplicate email returns ErrEmailAlreadyExists", func(t *testing.T) {
		req := model.CreateApplicantRequest{
			Name:            "Lesmana Duplicate",
			Email:           "lesmana@example.com", // same email
			Phone:           "089999999999",
			AppliedPosition: "Back-End Engineer",
		}

		applicant, err := svc.CreateApplicant(ctx, req)
		assert.ErrorIs(t, err, service.ErrEmailAlreadyExists)
		assert.Nil(t, applicant)
	})
}

func TestApplicantService_GetApplicantByID(t *testing.T) {
	_, svc := setupTestDB(t)
	ctx := context.Background()

	created, err := svc.CreateApplicant(ctx, model.CreateApplicantRequest{
		Name:            "Budi Santoso",
		Email:           "budi@example.com",
		Phone:           "081122334455",
		AppliedPosition: "Mobile Developer",
	})
	require.NoError(t, err)

	t.Run("success get applicant by id", func(t *testing.T) {
		found, err := svc.GetApplicantByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, "Budi Santoso", found.Name)
	})

	t.Run("not found for non-existent id", func(t *testing.T) {
		found, err := svc.GetApplicantByID(ctx, "non-existent-uuid")
		assert.ErrorIs(t, err, service.ErrApplicantNotFound)
		assert.Nil(t, found)
	})
}

func TestApplicantService_ListApplicants(t *testing.T) {
	_, svc := setupTestDB(t)
	ctx := context.Background()

	// Seed 3 applicants
	_, err := svc.CreateApplicant(ctx, model.CreateApplicantRequest{
		Name:            "Alice Doe",
		Email:           "alice@example.com",
		Phone:           "081111111111",
		AppliedPosition: "Back-End Engineer",
	})
	require.NoError(t, err)

	b, err := svc.CreateApplicant(ctx, model.CreateApplicantRequest{
		Name:            "Bob Smith",
		Email:           "bob@example.com",
		Phone:           "082222222222",
		AppliedPosition: "Frontend Engineer",
	})
	require.NoError(t, err)

	// Move Bob to INTERVIEW stage
	_, err = svc.UpdateApplicantStage(ctx, b.ID, model.UpdateStageRequest{
		Stage: model.StageInterview,
	})
	require.NoError(t, err)

	_, err = svc.CreateApplicant(ctx, model.CreateApplicantRequest{
		Name:            "Charlie Brown",
		Email:           "charlie@example.com",
		Phone:           "083333333333",
		AppliedPosition: "Back-End Engineer",
	})
	require.NoError(t, err)

	t.Run("list all with pagination meta", func(t *testing.T) {
		list, meta, err := svc.ListApplicants(ctx, model.ListApplicantsQuery{
			Page:  1,
			Limit: 2,
		})
		require.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, int64(3), meta.TotalRecords)
		assert.Equal(t, 2, meta.TotalPages)
		assert.Equal(t, 1, meta.Page)
		assert.Equal(t, 2, meta.Limit)
	})

	t.Run("filter by stage", func(t *testing.T) {
		list, meta, err := svc.ListApplicants(ctx, model.ListApplicantsQuery{
			Stage: model.StageInterview,
		})
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "Bob Smith", list[0].Name)
		assert.Equal(t, int64(1), meta.TotalRecords)
	})

	t.Run("filter by position", func(t *testing.T) {
		list, meta, err := svc.ListApplicants(ctx, model.ListApplicantsQuery{
			Position: "Frontend",
		})
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "Bob Smith", list[0].Name)
		assert.Equal(t, int64(1), meta.TotalRecords)
	})

	t.Run("search keyword in name or email", func(t *testing.T) {
		list, meta, err := svc.ListApplicants(ctx, model.ListApplicantsQuery{
			Search: "alice",
		})
		require.NoError(t, err)
		assert.Len(t, list, 1)
		assert.Equal(t, "Alice Doe", list[0].Name)
		assert.Equal(t, int64(1), meta.TotalRecords)
	})

	t.Run("invalid stage returns ErrInvalidStage", func(t *testing.T) {
		list, meta, err := svc.ListApplicants(ctx, model.ListApplicantsQuery{
			Stage: "INVALID_STAGE",
		})
		assert.ErrorIs(t, err, service.ErrInvalidStage)
		assert.Nil(t, list)
		assert.Nil(t, meta)
	})
}

func TestApplicantService_UpdateApplicantStage(t *testing.T) {
	_, svc := setupTestDB(t)
	ctx := context.Background()

	created, err := svc.CreateApplicant(ctx, model.CreateApplicantRequest{
		Name:            "Diana Prince",
		Email:           "diana@example.com",
		Phone:           "085555555555",
		AppliedPosition: "QA Engineer",
	})
	require.NoError(t, err)
	assert.Equal(t, model.StageApplied, created.Stage)

	t.Run("success update stage to TECHNICAL_TEST with notes", func(t *testing.T) {
		updated, err := svc.UpdateApplicantStage(ctx, created.ID, model.UpdateStageRequest{
			Stage: model.StageTechnicalTest,
			Notes: "Passed screening, assigned take-home test",
		})
		require.NoError(t, err)
		assert.Equal(t, model.StageTechnicalTest, updated.Stage)
		assert.Equal(t, "Passed screening, assigned take-home test", updated.Notes)
	})

	t.Run("reject invalid stage enum", func(t *testing.T) {
		updated, err := svc.UpdateApplicantStage(ctx, created.ID, model.UpdateStageRequest{
			Stage: "RANDOM_STAGE",
		})
		assert.ErrorIs(t, err, service.ErrInvalidStage)
		assert.Nil(t, updated)
	})

	t.Run("error when applicant id not found", func(t *testing.T) {
		updated, err := svc.UpdateApplicantStage(ctx, "00000000-0000-0000-0000-000000000000", model.UpdateStageRequest{
			Stage: model.StageOffered,
		})
		assert.ErrorIs(t, err, service.ErrApplicantNotFound)
		assert.Nil(t, updated)
	})
}
