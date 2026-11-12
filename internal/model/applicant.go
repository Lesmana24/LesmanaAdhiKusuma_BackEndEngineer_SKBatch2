package model

import (
	"strings"
	"time"
)

// Recruitment stages
const (
	StageApplied       = "APPLIED"
	StageScreening     = "SCREENING"
	StageInterview     = "INTERVIEW"
	StageTechnicalTest = "TECHNICAL_TEST"
	StageOffered       = "OFFERED"
	StageAccepted      = "ACCEPTED"
	StageRejected      = "REJECTED"
)

// AllowedStages returns all valid recruitment stages
func AllowedStages() []string {
	return []string{
		StageApplied,
		StageScreening,
		StageInterview,
		StageTechnicalTest,
		StageOffered,
		StageAccepted,
		StageRejected,
	}
}

// IsValidStage checks whether a given stage is valid
func IsValidStage(stage string) bool {
	upper := strings.ToUpper(strings.TrimSpace(stage))
	for _, s := range AllowedStages() {
		if s == upper {
			return true
		}
	}
	return false
}

// Applicant represents the domain model for an applicant record
type Applicant struct {
	ID              string    `gorm:"primaryKey;type:varchar(36)" json:"id" example:"8f123456-789a-bcde-f012-3456789abcde"`
	Name            string    `gorm:"type:varchar(100);not null" json:"name" example:"Budi Santoso"`
	Email           string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" example:"budi.santoso@example.com"`
	Phone           string    `gorm:"type:varchar(20);not null" json:"phone" example:"081234567890"`
	AppliedPosition string    `gorm:"type:varchar(100);not null" json:"applied_position" example:"Back-End Engineer"`
	Stage           string    `gorm:"type:varchar(30);not null;default:'APPLIED'" json:"stage" example:"APPLIED"`
	Notes           string    `gorm:"type:text" json:"notes" example:"Kandidat berpengalaman Go & PostgreSQL"`
	CreatedAt       time.Time `json:"created_at" example:"2026-11-12T19:30:00Z"`
	UpdatedAt       time.Time `json:"updated_at" example:"2026-11-12T19:30:00Z"`
}

