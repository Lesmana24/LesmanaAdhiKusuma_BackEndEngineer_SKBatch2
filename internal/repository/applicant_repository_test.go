package repository_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/model"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRepoDB(t *testing.T) (*gorm.DB, repository.ApplicantRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.Applicant{})
	require.NoError(t, err)

	db.Exec("DELETE FROM applicants")

	repo := repository.NewApplicantRepository(db)
	return db, repo
}

func TestApplicantRepository_CRUD(t *testing.T) {
	_, repo := setupRepoDB(t)
	ctx := context.Background()

	app := &model.Applicant{
		ID:              uuid.NewString(),
		Name:            "Kurnia Meiga",
		Email:           "kurnia@example.com",
		Phone:           "081234567890",
		AppliedPosition: "Go Developer",
		Stage:           model.StageApplied,
		Notes:           "Initial note",
	}

	t.Run("Create and FindByID", func(t *testing.T) {
		err := repo.Create(ctx, app)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, app.ID)
		require.NoError(t, err)
		assert.Equal(t, app.ID, found.ID)
		assert.Equal(t, app.Name, found.Name)
	})

	t.Run("FindByEmail case-insensitive", func(t *testing.T) {
		found, err := repo.FindByEmail(ctx, "KURNIA@EXAMPLE.COM")
		require.NoError(t, err)
		assert.Equal(t, app.ID, found.ID)
	})

	t.Run("UpdateStage", func(t *testing.T) {
		updated, err := repo.UpdateStage(ctx, app.ID, model.StageInterview, "Passed HR")
		require.NoError(t, err)
		assert.Equal(t, model.StageInterview, updated.Stage)
		assert.Equal(t, "Passed HR", updated.Notes)
	})

	t.Run("FindAll with pagination, filters and search", func(t *testing.T) {
		// Insert another record
		_ = repo.Create(ctx, &model.Applicant{
			ID:              uuid.NewString(),
			Name:            "Bambang Pamungkas",
			Email:           "bambang@example.com",
			Phone:           "082222222222",
			AppliedPosition: "Frontend Developer",
			Stage:           model.StageApplied,
		})

		// Filter by stage
		list, total, err := repo.FindAll(ctx, model.ListApplicantsQuery{
			Page:  1,
			Limit: 10,
			Stage: model.StageInterview,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, list, 1)
		assert.Equal(t, "Kurnia Meiga", list[0].Name)

		// Search
		list, total, err = repo.FindAll(ctx, model.ListApplicantsQuery{
			Page:   1,
			Limit:  10,
			Search: "bambang",
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, list, 1)
		assert.Equal(t, "Bambang Pamungkas", list[0].Name)
	})
}
