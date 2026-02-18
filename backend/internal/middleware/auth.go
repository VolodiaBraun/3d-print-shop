package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	jwtpkg "github.com/brown/3d-print-shop/pkg/jwt"
	"github.com/brown/3d-print-shop/pkg/response"
)

const (
	ContextKeyUserID = "userID"
	ContextKeyRole   = "role"
)

// AuthRequired validates the JWT access token from the Authorization header.
func AuthRequired(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", "Неверный формат токена")
			c.Abort()
			return
		}

		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			if errors.Is(err, jwtpkg.ErrTokenExpired) {
				response.Error(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "Токен истёк")
			} else {
				response.Error(c, http.StatusUnauthorized, "INVALID_TOKEN", "Невалидный токен")
			}
			c.Abort()
			return
		}

		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// RequireRole checks that the authenticated user has one of the allowed roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			response.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Требуется авторизация")
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			response.InternalError(c)
			c.Abort()
			return
		}

		for _, r := range roles {
			if userRole == r {
				c.Next()
				return
			}
		}

		response.Error(c, http.StatusForbidden, "FORBIDDEN", "Недостаточно прав")
		c.Abort()
	}
}

// OptionalAuth is like AuthRequired but never blocks the request.
// If a valid Bearer token is present it sets userID/role in context, otherwise just calls Next().
func OptionalAuth(jwtManager *jwtpkg.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.Next()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.Next()
			return
		}
		claims, err := jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			c.Next()
			return
		}
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// GetUserID extracts the user ID from the Gin context.
func GetUserID(c *gin.Context) (int, bool) {
	id, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	userID, ok := id.(int)
	return userID, ok
}

// GetRole extracts the user role from the Gin context.
func GetRole(c *gin.Context) (string, bool) {
	role, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}
	r, ok := role.(string)
	return r, ok
}
