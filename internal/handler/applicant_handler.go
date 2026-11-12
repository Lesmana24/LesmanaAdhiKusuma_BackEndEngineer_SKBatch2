package handler

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/model"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/service"
	"github.com/lesmana/sekolah-karir-applicant-api/pkg/response"
)

// ApplicantHandler handles HTTP requests for applicants
type ApplicantHandler struct {
	svc service.ApplicantService
}

// NewApplicantHandler creates a new ApplicantHandler instance
func NewApplicantHandler(svc service.ApplicantService) *ApplicantHandler {
	return &ApplicantHandler{svc: svc}
}

// HealthCheck godoc
// @Summary      Health Check
// @Description  Check service availability and health status
// @Tags         System
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /health [get]
func (h *ApplicantHandler) HealthCheck(c *gin.Context) {
	response.Success(c, "Service is healthy and ready to serve requests", gin.H{
		"status": "UP",
	})
}

// CreateApplicant godoc
// @Summary      Create Applicant
// @Description  Register a new job applicant record with initial stage APPLIED
// @Tags         Applicants
// @Accept       json
// @Produce      json
// @Param        request body model.CreateApplicantRequest true "Applicant payload"
// @Success      201  {object}  response.Response{data=model.Applicant}
// @Failure      400  {object}  response.Response
// @Failure      409  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/applicants [post]
func (h *ApplicantHandler) CreateApplicant(c *gin.Context) {
	var req model.CreateApplicantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMessages := formatValidationErrors(err)
		response.BadRequest(c, "Validation failed", errMessages...)
		return
	}

	applicant, err := h.svc.CreateApplicant(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			response.Conflict(c, err.Error())
			return
		}
		response.InternalServerError(c, "Failed to create applicant: "+err.Error())
		return
	}

	response.Created(c, "Applicant created successfully", applicant)
}

// GetApplicantByID godoc
// @Summary      Get Applicant Detail
// @Description  Retrieve an applicant record by its UUID
// @Tags         Applicants
// @Produce      json
// @Param        id   path      string  true  "Applicant UUID"
// @Success      200  {object}  response.Response{data=model.Applicant}
// @Failure      404  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/applicants/{id} [get]
func (h *ApplicantHandler) GetApplicantByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Applicant ID is required")
		return
	}

	applicant, err := h.svc.GetApplicantByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrApplicantNotFound) {
			response.NotFound(c, "Applicant not found with the specified ID")
			return
		}
		response.InternalServerError(c, "Failed to retrieve applicant: "+err.Error())
		return
	}

	response.Success(c, "Applicant retrieved successfully", applicant)
}

// ListApplicants godoc
// @Summary      List Applicants
// @Description  Get a paginated list of applicants with optional filtering and keyword search
// @Tags         Applicants
// @Produce      json
// @Param        page      query     int     false  "Page number (default: 1)"
// @Param        limit     query     int     false  "Records per page (default: 10, max: 100)"
// @Param        stage     query     string  false  "Filter by stage (e.g. APPLIED, SCREENING, INTERVIEW, etc.)"
// @Param        position  query     string  false  "Filter by position name (case-insensitive substring)"
// @Param        search    query     string  false  "Search keyword in name or email"
// @Success      200  {object}  response.Response{data=[]model.Applicant,meta=response.PaginationMeta}
// @Failure      400  {object}  response.Response
// @Failure      500  {object}  response.Response
// @Router       /api/v1/applicants [get]
func (h *ApplicantHandler) ListApplicants(c *gin.Context) {
	var query model.ListApplicantsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.BadRequest(c, "Invalid query parameters: "+err.Error())
		return
	}

	applicants, meta, err := h.svc.ListApplicants(c.Request.Context(), query)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStage) {
			response.BadRequest(c, "Invalid stage filter. Allowed stages: "+fmt.Sprint(model.AllowedStages()))
			return
		}
		response.InternalServerError(c, "Failed to list applicants: "+err.Error())
		return
	}

	response.SuccessWithMeta(c, "Applicants retrieved successfully", applicants, meta)
}

// UpdateApplicantStage godoc
// @Summary      Update Applicant Stage
// @Description  Update the recruitment stage and notes of an applicant
// @Tags         Applicants
// @Accept       json
// @Produce      json
// @Param        id       path      string                   true  "Applicant UUID"
// @Param        request  body      model.UpdateStageRequest true  "Update stage payload"
// @Success      200      {object}  response.Response{data=model.Applicant}
// @Failure      400      {object}  response.Response
// @Failure      404      {object}  response.Response
// @Failure      500      {object}  response.Response
// @Router       /api/v1/applicants/{id}/stage [patch]
func (h *ApplicantHandler) UpdateApplicantStage(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "Applicant ID is required")
		return
	}

	var req model.UpdateStageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMessages := formatValidationErrors(err)
		response.BadRequest(c, "Validation failed", errMessages...)
		return
	}

	updated, err := h.svc.UpdateApplicantStage(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrApplicantNotFound) {
			response.NotFound(c, "Applicant not found with the specified ID")
			return
		}
		if errors.Is(err, service.ErrInvalidStage) {
			response.BadRequest(c, "Invalid recruitment stage. Allowed stages: "+fmt.Sprint(model.AllowedStages()))
			return
		}
		response.InternalServerError(c, "Failed to update stage: "+err.Error())
		return
	}

	response.Success(c, "Applicant stage updated successfully", updated)
}

// formatValidationErrors translates struct validation errors into human-readable messages
func formatValidationErrors(err error) []string {
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		result := make([]string, 0, len(errs))
		for _, fe := range errs {
			var msg string
			field := fe.Field()
			switch fe.Tag() {
			case "required":
				msg = fmt.Sprintf("field '%s' is required", field)
			case "email":
				msg = fmt.Sprintf("field '%s' must be a valid email address", field)
			case "min":
				msg = fmt.Sprintf("field '%s' must have at least %s characters", field, fe.Param())
			case "max":
				msg = fmt.Sprintf("field '%s' cannot exceed %s characters", field, fe.Param())
			default:
				msg = fmt.Sprintf("field '%s' failed validation for tag '%s'", field, fe.Tag())
			}
			result = append(result, msg)
		}
		return result
	}
	return []string{err.Error()}
}

