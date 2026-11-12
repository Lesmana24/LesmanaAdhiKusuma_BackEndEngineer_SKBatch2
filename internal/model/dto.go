package model

// CreateApplicantRequest represents the payload to register a new applicant
type CreateApplicantRequest struct {
	Name            string `json:"name" binding:"required,min=2,max=100" example:"Budi Santoso"`
	Email           string `json:"email" binding:"required,email,max=100" example:"budi.santoso@example.com"`
	Phone           string `json:"phone" binding:"required,min=8,max=20" example:"081234567890"`
	AppliedPosition string `json:"applied_position" binding:"required,min=2,max=100" example:"Back-End Engineer"`
	Notes           string `json:"notes" binding:"omitempty,max=500" example:"Kandidat direferensikan oleh tim engineer"`
}

// UpdateStageRequest represents the payload to update an applicant's stage
type UpdateStageRequest struct {
	Stage string `json:"stage" binding:"required" example:"INTERVIEW"`
	Notes string `json:"notes" binding:"omitempty,max=500" example:"Lolos screening CV, lanjut interview teknis"`
}

// ListApplicantsQuery represents query parameters for filtering and pagination
type ListApplicantsQuery struct {
	Page     int    `form:"page" example:"1"`
	Limit    int    `form:"limit" example:"10"`
	Stage    string `form:"stage" example:"APPLIED"`
	Position string `form:"position" example:"Back-End Engineer"`
	Search   string `form:"search" example:"Budi"`
}

