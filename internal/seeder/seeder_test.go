package seeder_test

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/model"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/seeder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSeedApplicants(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.Applicant{})
	require.NoError(t, err)

	db.Exec("DELETE FROM applicants")

	t.Run("successfully seed 10 applicants on empty database", func(t *testing.T) {
		err := seeder.SeedApplicants(db)
		require.NoError(t, err)

		var count int64
		err = db.Model(&model.Applicant{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(10), count)

		// Verify distinct stages are present
		var appliedCount int64
		db.Model(&model.Applicant{}).Where("stage = ?", model.StageApplied).Count(&appliedCount)
		assert.GreaterOrEqual(t, appliedCount, int64(1))

		var acceptedCount int64
		db.Model(&model.Applicant{}).Where("stage = ?", model.StageAccepted).Count(&acceptedCount)
		assert.Equal(t, int64(1), acceptedCount)
	})

	t.Run("idempotency: calling seeder again should not duplicate records", func(t *testing.T) {
		// Run seeder again
		err := seeder.SeedApplicants(db)
		require.NoError(t, err)

		var count int64
		err = db.Model(&model.Applicant{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(10), count)
	})
}

