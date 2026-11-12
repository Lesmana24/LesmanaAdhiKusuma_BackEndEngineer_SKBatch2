package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response standardizes API JSON output
type Response struct {
	Success bool            `json:"success" example:"true"`
	Message string          `json:"message" example:"Operation successful"`
	Data    interface{}     `json:"data,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
	Errors  []string        `json:"errors,omitempty"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page         int   `json:"page" example:"1"`
	Limit        int   `json:"limit" example:"10"`
	TotalRecords int64 `json:"total_records" example:"50"`
	TotalPages   int   `json:"total_pages" example:"5"`
}

// JSON sends a generic JSON response
func JSON(c *gin.Context, statusCode int, success bool, message string, data interface{}, meta *PaginationMeta, errors []string) {
	c.JSON(statusCode, Response{
		Success: success,
		Message: message,
		Data:    data,
		Meta:    meta,
		Errors:  errors,
	})
}

// Success sends an OK (200) response with data
func Success(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusOK, true, message, data, nil, nil)
}

// SuccessWithMeta sends an OK (200) response with data and pagination metadata
func SuccessWithMeta(c *gin.Context, message string, data interface{}, meta *PaginationMeta) {
	JSON(c, http.StatusOK, true, message, data, meta, nil)
}

// Created sends a Created (201) response with data
func Created(c *gin.Context, message string, data interface{}) {
	JSON(c, http.StatusCreated, true, message, data, nil, nil)
}

// Error sends an error response with given status code
func Error(c *gin.Context, statusCode int, message string, errors ...string) {
	var errList []string
	if len(errors) > 0 {
		errList = errors
	}
	JSON(c, statusCode, false, message, nil, nil, errList)
}

// BadRequest sends a 400 Bad Request error
func BadRequest(c *gin.Context, message string, errors ...string) {
	Error(c, http.StatusBadRequest, message, errors...)
}

// NotFound sends a 404 Not Found error
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// Conflict sends a 409 Conflict error
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}

// InternalServerError sends a 500 Internal Server Error
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

