package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/pkg/jwt"
)

// AuthService handles user authentication.
type AuthService struct {
	userRepo domain.UserRepository
	tokens   *AuthTokenService
	log      *zap.Logger
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo domain.UserRepository, tokens *AuthTokenService, log *zap.Logger) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		tokens:   tokens,
		log:      log,
	}
}

// LoginInput represents the login request data.
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// TokenResponse is the response with access and refresh tokens.
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// Login authenticates a user by email and password, returning a token pair.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("find user: %w", err)
	}

	if !user.IsActive {
		return nil, domain.ErrAccountDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	pair, err := s.tokens.GenerateTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	s.log.Info("user logged in", zap.Int("userID", user.ID), zap.String("role", user.Role))

	return &TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}

// Refresh exchanges a valid refresh token for a new token pair.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	pair, err := s.tokens.RefreshTokenPair(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, jwt.ErrInvalidToken) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("refresh tokens: %w", err)
	}

	return &TokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}, nil
}
