package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/internal/middleware"
	"github.com/brown/3d-print-shop/internal/service"
	"github.com/brown/3d-print-shop/pkg/response"
)

type ReviewHandler struct {
	reviewService *service.ReviewService
}

func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// RegisterPublicRoutes registers public review routes.
// Note: uses :slug param name to match existing /products/:slug route in Gin.
func (h *ReviewHandler) RegisterPublicRoutes(rg *gin.RouterGroup) {
	rg.GET("/products/:slug/reviews", h.GetProductReviews)
}

// RegisterProtectedRoutes registers auth-required review routes.
func (h *ReviewHandler) RegisterProtectedRoutes(rg *gin.RouterGroup) {
	rg.POST("/products/:slug/reviews", h.CreateReview)
	rg.GET("/reviews/my", h.GetMyReviews)
}

// RegisterAdminRoutes registers admin review routes.
func (h *ReviewHandler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.GET("/reviews", h.ListReviews)
	rg.PUT("/reviews/:id/approve", h.ApproveReview)
	rg.PUT("/reviews/:id/reject", h.RejectReview)
	rg.DELETE("/reviews/:id", h.DeleteReview)
}

// GetProductReviews handles GET /api/v1/products/:slug/reviews
func (h *ReviewHandler) GetProductReviews(c *gin.Context) {
	productID, err := strconv.Atoi(c.Param("slug"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID товара")
		return
	}

	reviews, err := h.reviewService.GetProductReviews(c.Request.Context(), productID)
	if err != nil {
		response.InternalError(c)
		return
	}

	// Return safe user data (only name)
	type reviewUserResponse struct {
		ID        int    `json:"id"`
		FirstName string `json:"firstName,omitempty"`
		LastName  string `json:"lastName,omitempty"`
	}
	type reviewResponse struct {
		ID        int                `json:"id"`
		Rating    int                `json:"rating"`
		Comment   string             `json:"comment,omitempty"`
		User      reviewUserResponse `json:"user"`
		CreatedAt string             `json:"createdAt"`
	}

	result := make([]reviewResponse, 0, len(reviews))
	for _, r := range reviews {
		rr := reviewResponse{
			ID:        r.ID,
			Rating:    r.Rating,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if r.User != nil {
			rr.User.ID = r.User.ID
			if r.User.FirstName != nil {
				rr.User.FirstName = *r.User.FirstName
			}
			if r.User.LastName != nil {
				rr.User.LastName = *r.User.LastName
			}
		}
		result = append(result, rr)
	}

	response.OK(c, result)
}

// CreateReview handles POST /api/v1/products/:slug/reviews
func (h *ReviewHandler) CreateReview(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Не авторизован")
		return
	}

	productID, err := strconv.Atoi(c.Param("slug"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID товара")
		return
	}

	var input service.CreateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.ValidationError(c, []response.ErrorDetail{
			{Message: "Рейтинг (1-5) и ID заказа обязательны"},
		})
		return
	}
	input.ProductID = productID

	review, err := h.reviewService.CreateReview(c.Request.Context(), userID, input)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrAlreadyReviewed):
			response.Error(c, http.StatusConflict, "ALREADY_REVIEWED", "Вы уже оставили отзыв на этот товар")
		case errors.Is(err, domain.ErrOrderNotDelivered):
			response.Error(c, http.StatusBadRequest, "ORDER_NOT_DELIVERED", "Заказ не доставлен или товар не в заказе")
		default:
			response.InternalError(c)
		}
		return
	}

	response.Created(c, review)
}

// GetMyReviews handles GET /api/v1/reviews/my
func (h *ReviewHandler) GetMyReviews(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Не авторизован")
		return
	}

	reviews, err := h.reviewService.GetMyReviews(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c)
		return
	}

	response.OK(c, reviews)
}

// ListReviews handles GET /api/v1/admin/reviews
func (h *ReviewHandler) ListReviews(c *gin.Context) {
	filter := domain.ReviewFilter{
		Status: c.Query("status"),
	}
	if p, _ := strconv.Atoi(c.DefaultQuery("page", "1")); p > 0 {
		filter.Page = p
	}
	if l, _ := strconv.Atoi(c.DefaultQuery("limit", "20")); l > 0 {
		filter.Limit = l
	}
	if pid, _ := strconv.Atoi(c.Query("productId")); pid > 0 {
		filter.ProductID = pid
	}

	reviews, total, err := h.reviewService.ListReviews(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c)
		return
	}

	totalPages := int(total) / filter.Limit
	if int(total)%filter.Limit > 0 {
		totalPages++
	}
	response.Paginated(c, reviews, response.PaginationMeta{
		Page:       filter.Page,
		Limit:      filter.Limit,
		Total:      total,
		TotalPages: totalPages,
	})
}

// ApproveReview handles PUT /api/v1/admin/reviews/:id/approve
func (h *ReviewHandler) ApproveReview(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	if err := h.reviewService.ApproveReview(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			response.NotFound(c, "Отзыв не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{"message": "Отзыв одобрен"})
}

// RejectReview handles PUT /api/v1/admin/reviews/:id/reject
func (h *ReviewHandler) RejectReview(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	if err := h.reviewService.RejectReview(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			response.NotFound(c, "Отзыв не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{"message": "Отзыв отклонён"})
}

// DeleteReview handles DELETE /api/v1/admin/reviews/:id
func (h *ReviewHandler) DeleteReview(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "INVALID_ID", "Некорректный ID")
		return
	}

	if err := h.reviewService.DeleteReview(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrReviewNotFound) {
			response.NotFound(c, "Отзыв не найден")
			return
		}
		response.InternalError(c)
		return
	}

	response.OK(c, gin.H{"message": "Отзыв удалён"})
}
