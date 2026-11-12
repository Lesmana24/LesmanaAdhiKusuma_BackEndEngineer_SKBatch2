package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/handler"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/model"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/repository"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/service"
	"github.com/lesmana/sekolah-karir-applicant-api/pkg/response"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.Applicant{})
	require.NoError(t, err)

	db.Exec("DELETE FROM applicants")

	repo := repository.NewApplicantRepository(db)
	svc := service.NewApplicantService(repo)
	h := handler.NewApplicantHandler(svc)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", h.HealthCheck)

	v1 := r.Group("/api/v1")
	{
		applicants := v1.Group("/applicants")
		{
			applicants.POST("", h.CreateApplicant)
			applicants.GET("", h.ListApplicants)
			applicants.GET("/:id", h.GetApplicantByID)
			applicants.PATCH("/:id/stage", h.UpdateApplicantStage)
		}
	}

	return r, db
}

func TestHandler_HealthCheck(t *testing.T) {
	r, _ := setupTestRouter(t)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var res response.Response
	err := json.Unmarshal(w.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.True(t, res.Success)
	assert.Contains(t, res.Message, "healthy")
}

func TestHandler_CreateApplicant(t *testing.T) {
	r, _ := setupTestRouter(t)

	t.Run("success create applicant returns 201", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":             "Lesmana Adh Kusuma",
			"email":            "lesmana@example.com",
			"phone":            "081234567890",
			"applied_position": "Back-End Engineer",
			"notes":            "Candidate SKR26-0005",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/applicants", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var res struct {
			Success bool            `json:"success"`
			Message string          `json:"message"`
			Data    model.Applicant `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.True(t, res.Success)
		assert.NotEmpty(t, res.Data.ID)
		assert.Equal(t, "Lesmana Adh Kusuma", res.Data.Name)
		assert.Equal(t, model.StageApplied, res.Data.Stage)
	})

	t.Run("validation failure returns 400 with details", func(t *testing.T) {
		// Missing name and invalid email
		payload := map[string]interface{}{
			"email":            "invalid-email-format",
			"phone":            "123", // too short
			"applied_position": "",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/applicants", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var res response.Response
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.False(t, res.Success)
		assert.NotEmpty(t, res.Errors)
	})

	t.Run("duplicate email returns 409 conflict", func(t *testing.T) {
		payload := map[string]interface{}{
			"name":             "Duplicate Person",
			"email":            "lesmana@example.com", // existing email from first test
			"phone":            "081199887766",
			"applied_position": "Fullstack Engineer",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/applicants", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

func TestHandler_ListApplicants(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create 2 applicants
	for _, app := range []map[string]string{
		{"name": "Dev One", "email": "dev1@example.com", "phone": "081111111111", "applied_position": "Go Engineer"},
		{"name": "Dev Two", "email": "dev2@example.com", "phone": "082222222222", "applied_position": "React Engineer"},
	} {
		body, _ := json.Marshal(app)
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/applicants", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	t.Run("success get list applicants with pagination meta", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/applicants?page=1&limit=10", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res struct {
			Success bool                    `json:"success"`
			Data    []model.Applicant       `json:"data"`
			Meta    response.PaginationMeta `json:"meta"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.True(t, res.Success)
		assert.Len(t, res.Data, 2)
		assert.Equal(t, int64(2), res.Meta.TotalRecords)
	})

	t.Run("invalid stage filter returns 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/applicants?stage=UNKNOWN_STAGE", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestHandler_GetApplicantByID(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create an applicant
	payload := map[string]interface{}{
		"name":             "John Doe",
		"email":            "john@example.com",
		"phone":            "081234567890",
		"applied_position": "DevOps Engineer",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/applicants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createRes struct {
		Data model.Applicant `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createRes)
	createdID := createRes.Data.ID

	t.Run("found applicant by id returns 200", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/applicants/"+createdID, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var res response.Response
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		assert.True(t, res.Success)
	})

	t.Run("non-existent id returns 404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/applicants/random-id-12345", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandler_UpdateApplicantStage(t *testing.T) {
	r, _ := setupTestRouter(t)

	// Create an applicant
	payload := map[string]interface{}{
		"name":             "Eka Wijaya",
		"email":            "eka@example.com",
		"phone":            "081234567890",
		"applied_position": "Back-End Engineer",
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/applicants", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var createRes struct {
		Data model.Applicant `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createRes)
	createdID := createRes.Data.ID

	t.Run("success update stage returns 200 with updated model", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"stage": "INTERVIEW",
			"notes": "Passed HR screening, scheduled user interview",
		}
		updateBody, _ := json.Marshal(updatePayload)
		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/applicants/"+createdID+"/stage", bytes.NewBuffer(updateBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var res struct {
			Success bool            `json:"success"`
			Data    model.Applicant `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &res)
		require.NoError(t, err)
		assert.True(t, res.Success)
		assert.Equal(t, "INTERVIEW", res.Data.Stage)
		assert.Equal(t, "Passed HR screening, scheduled user interview", res.Data.Notes)
	})

	t.Run("invalid stage enum returns 400", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"stage": "INVALID_STATE",
		}
		updateBody, _ := json.Marshal(updatePayload)
		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/applicants/"+createdID+"/stage", bytes.NewBuffer(updateBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("non-existent id returns 404", func(t *testing.T) {
		updatePayload := map[string]interface{}{
			"stage": "ACCEPTED",
		}
		updateBody, _ := json.Marshal(updatePayload)
		req, _ := http.NewRequest(http.MethodPatch, "/api/v1/applicants/non-existent-uuid/stage", bytes.NewBuffer(updateBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

