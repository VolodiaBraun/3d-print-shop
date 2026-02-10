package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"github.com/brown/3d-print-shop/internal/domain"
	"github.com/brown/3d-print-shop/pkg/jwt"
)

// AuthService handles user authentication.
type AuthService struct {
	userRepo domain.UserRepository
	tokens   *AuthTokenService
	botToken string
	log      *zap.Logger
}

// NewAuthService creates a new authentication service.
func NewAuthService(userRepo domain.UserRepository, tokens *AuthTokenService, botToken string, log *zap.Logger) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		tokens:   tokens,
		botToken: botToken,
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

// TelegramUser represents user data extracted from Telegram initData.
type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// TelegramLoginResponse is returned after a successful Telegram login.
type TelegramLoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         struct {
		ID         int    `json:"id"`
		FirstName  string `json:"firstName"`
		TelegramID int64  `json:"telegramId"`
		Role       string `json:"role"`
	} `json:"user"`
}

var (
	ErrInvalidInitData = errors.New("invalid initData")
	ErrInitDataExpired = errors.New("initData expired")
)

// LoginTelegram validates Telegram initData and returns tokens.
func (s *AuthService) LoginTelegram(ctx context.Context, initData string) (*TelegramLoginResponse, error) {
	if s.botToken == "" {
		return nil, fmt.Errorf("telegram bot token not configured")
	}

	// Parse initData (URL-encoded query string)
	values, err := url.ParseQuery(initData)
	if err != nil {
		return nil, ErrInvalidInitData
	}

	receivedHash := values.Get("hash")
	if receivedHash == "" {
		return nil, ErrInvalidInitData
	}

	// Build data_check_string: sort all params except "hash"
	var pairs []string
	for k, v := range values {
		if k == "hash" {
			continue
		}
		pairs = append(pairs, k+"="+v[0])
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// HMAC-SHA256 validation per Telegram docs
	secretKey := hmacSHA256([]byte("WebAppData"), []byte(s.botToken))
	computedHash := hex.EncodeToString(hmacSHA256(secretKey, []byte(dataCheckString)))

	if !hmac.Equal([]byte(computedHash), []byte(receivedHash)) {
		s.log.Warn("telegram initData hash mismatch")
		return nil, ErrInvalidInitData
	}

	// Check auth_date (not older than 5 minutes)
	authDateStr := values.Get("auth_date")
	if authDateStr == "" {
		return nil, ErrInvalidInitData
	}
	authDate, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInitData
	}
	if time.Now().Unix()-authDate > 300 {
		return nil, ErrInitDataExpired
	}

	// Extract user JSON
	userJSON := values.Get("user")
	if userJSON == "" {
		return nil, ErrInvalidInitData
	}
	var tgUser TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &tgUser); err != nil {
		return nil, ErrInvalidInitData
	}
	if tgUser.ID == 0 {
		return nil, ErrInvalidInitData
	}

	// Find or create user
	user, err := s.userRepo.FindByTelegramID(ctx, tgUser.ID)
	if errors.Is(err, domain.ErrUserNotFound) {
		firstName := tgUser.FirstName
		lastName := tgUser.LastName
		username := tgUser.Username
		user = &domain.User{
			TelegramID: &tgUser.ID,
			FirstName:  &firstName,
			LastName:   &lastName,
			Username:   &username,
			Role:       "customer",
			IsActive:   true,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("create telegram user: %w", err)
		}
		s.log.Info("created user from telegram", zap.Int("userID", user.ID), zap.Int64("telegramID", tgUser.ID))
	} else if err != nil {
		return nil, fmt.Errorf("find user by telegram id: %w", err)
	} else {
		// Update user info if changed
		changed := false
		if user.FirstName == nil || *user.FirstName != tgUser.FirstName {
			user.FirstName = &tgUser.FirstName
			changed = true
		}
		if user.LastName == nil || *user.LastName != tgUser.LastName {
			user.LastName = &tgUser.LastName
			changed = true
		}
		if tgUser.Username != "" && (user.Username == nil || *user.Username != tgUser.Username) {
			user.Username = &tgUser.Username
			changed = true
		}
		if changed {
			_ = s.userRepo.Update(ctx, user)
		}
	}

	if !user.IsActive {
		return nil, domain.ErrAccountDisabled
	}

	// Generate JWT
	pair, err := s.tokens.GenerateTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	s.log.Info("telegram user logged in", zap.Int("userID", user.ID), zap.Int64("telegramID", tgUser.ID))

	resp := &TelegramLoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}
	resp.User.ID = user.ID
	if user.FirstName != nil {
		resp.User.FirstName = *user.FirstName
	}
	resp.User.TelegramID = tgUser.ID
	resp.User.Role = user.Role

	return resp, nil
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
