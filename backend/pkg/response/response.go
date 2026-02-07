package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type ErrorBody struct {
	Code    string        `json:"code"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

type errorResponse struct {
	Error ErrorBody `json:"error"`
}

type successResponse struct {
	Data interface{} `json:"data"`
}

type paginatedResponse struct {
	Data interface{} `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type PaginationMeta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, successResponse{Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, successResponse{Data: data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func Paginated(c *gin.Context, data interface{}, meta PaginationMeta) {
	c.JSON(http.StatusOK, paginatedResponse{Data: data, Meta: meta})
}

func Error(c *gin.Context, status int, code string, message string) {
	c.JSON(status, errorResponse{
		Error: ErrorBody{Code: code, Message: message},
	})
}

func ValidationError(c *gin.Context, details []ErrorDetail) {
	c.JSON(http.StatusBadRequest, errorResponse{
		Error: ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: "Некорректные данные",
			Details: details,
		},
	})
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Внутренняя ошибка сервера")
}
