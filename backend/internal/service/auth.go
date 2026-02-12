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
	userRepo       domain.UserRepository
	tokens         *AuthTokenService
	botToken       string
	loyaltyService *LoyaltyService
	log            *zap.Logger
}

// SetLoyaltyService sets the loyalty service (used to break circular dependency).
func (s *AuthService) SetLoyaltyService(ls *LoyaltyService) {
	s.loyaltyService = ls
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

	// Find or create/link user
	user, err := s.findOrLinkTelegramUser(ctx, tgUser.ID, "", tgUser.FirstName, tgUser.LastName, tgUser.Username)
	if err != nil {
		return nil, err
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

// RegisterInput represents the registration request data.
type RegisterInput struct {
	Name         string `json:"name" binding:"required,min=2"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	ReferralCode string `json:"referralCode"`
}

// RegisterResponse is returned after successful registration.
type RegisterResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         struct {
		ID        int    `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		Role      string `json:"role"`
	} `json:"user"`
}

// Register creates a new user account or links to existing TG account.
func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*RegisterResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Check if email already exists
	existing, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("find user: %w", err)
	}

	var user *domain.User
	if existing != nil {
		// Email exists — check if it's a TG-only account (no password)
		if existing.PasswordHash != "" {
			return nil, domain.ErrEmailAlreadyExists
		}
		// Link: add password to existing TG account
		existing.PasswordHash = string(hash)
		existing.FirstName = &input.Name
		if err := s.userRepo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("link account: %w", err)
		}
		user = existing
		s.log.Info("linked email registration to existing TG account",
			zap.Int("userID", user.ID), zap.String("email", input.Email))
	} else {
		// Create new user
		user = &domain.User{
			Email:        &input.Email,
			PasswordHash: string(hash),
			FirstName:    &input.Name,
			Role:         "customer",
			IsActive:     true,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}
		s.log.Info("user registered", zap.Int("userID", user.ID), zap.String("email", input.Email))
	}

	// Apply referral code if provided (non-blocking)
	if input.ReferralCode != "" && s.loyaltyService != nil {
		if err := s.loyaltyService.ApplyReferralCode(ctx, user.ID, input.ReferralCode); err != nil {
			s.log.Warn("failed to apply referral code during registration",
				zap.Error(err), zap.Int("userID", user.ID), zap.String("code", input.ReferralCode))
		}
	}

	pair, err := s.tokens.GenerateTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	resp := &RegisterResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}
	resp.User.ID = user.ID
	resp.User.Email = input.Email
	resp.User.FirstName = input.Name
	resp.User.Role = user.Role
	return resp, nil
}

// TelegramWidgetInput represents the Telegram Login Widget callback data.
type TelegramWidgetInput struct {
	ID        int64  `json:"id" binding:"required"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photoUrl"`
	AuthDate  int64  `json:"authDate" binding:"required"`
	Hash      string `json:"hash" binding:"required"`
	Email     string `json:"email"`
}

// LoginTelegramWidget validates Telegram Login Widget data and returns tokens.
func (s *AuthService) LoginTelegramWidget(ctx context.Context, input TelegramWidgetInput) (*TelegramLoginResponse, error) {
	if s.botToken == "" {
		return nil, fmt.Errorf("telegram bot token not configured")
	}

	// Build data_check_string: sort all non-empty fields except hash
	var pairs []string
	if input.AuthDate != 0 {
		pairs = append(pairs, fmt.Sprintf("auth_date=%d", input.AuthDate))
	}
	if input.Email != "" {
		pairs = append(pairs, "email="+input.Email)
	}
	if input.FirstName != "" {
		pairs = append(pairs, "first_name="+input.FirstName)
	}
	if input.ID != 0 {
		pairs = append(pairs, fmt.Sprintf("id=%d", input.ID))
	}
	if input.LastName != "" {
		pairs = append(pairs, "last_name="+input.LastName)
	}
	if input.PhotoURL != "" {
		pairs = append(pairs, "photo_url="+input.PhotoURL)
	}
	if input.Username != "" {
		pairs = append(pairs, "username="+input.Username)
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Telegram Login Widget uses SHA256(bot_token) as secret key (different from WebApp)
	secretKey := sha256Sum([]byte(s.botToken))
	computedHash := hex.EncodeToString(hmacSHA256(secretKey, []byte(dataCheckString)))

	if !hmac.Equal([]byte(computedHash), []byte(input.Hash)) {
		s.log.Warn("telegram widget hash mismatch")
		return nil, ErrInvalidInitData
	}

	// Check auth_date (not older than 5 minutes)
	if time.Now().Unix()-input.AuthDate > 300 {
		return nil, ErrInitDataExpired
	}

	// Find or link user
	user, err := s.findOrLinkTelegramUser(ctx, input.ID, input.Email, input.FirstName, input.LastName, input.Username)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, domain.ErrAccountDisabled
	}

	pair, err := s.tokens.GenerateTokenPair(ctx, user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	s.log.Info("telegram widget user logged in", zap.Int("userID", user.ID), zap.Int64("telegramID", input.ID))

	resp := &TelegramLoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}
	resp.User.ID = user.ID
	if user.FirstName != nil {
		resp.User.FirstName = *user.FirstName
	}
	resp.User.TelegramID = input.ID
	resp.User.Role = user.Role
	return resp, nil
}

// findOrLinkTelegramUser finds existing user by telegram_id or email, or creates new.
func (s *AuthService) findOrLinkTelegramUser(ctx context.Context, telegramID int64, email, firstName, lastName, username string) (*domain.User, error) {
	// 1. Try find by telegram_id
	user, err := s.userRepo.FindByTelegramID(ctx, telegramID)
	if err == nil {
		// Found by TG ID — update info
		changed := false
		if firstName != "" && (user.FirstName == nil || *user.FirstName != firstName) {
			user.FirstName = &firstName
			changed = true
		}
		if lastName != "" && (user.LastName == nil || *user.LastName != lastName) {
			user.LastName = &lastName
			changed = true
		}
		if username != "" && (user.Username == nil || *user.Username != username) {
			user.Username = &username
			changed = true
		}
		if email != "" && user.Email == nil {
			user.Email = &email
			changed = true
		}
		if changed {
			_ = s.userRepo.Update(ctx, user)
		}
		return user, nil
	}
	if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("find by telegram id: %w", err)
	}

	// 2. Try find by email (link TG to existing email account)
	if email != "" {
		user, err = s.userRepo.FindByEmail(ctx, email)
		if err == nil {
			// Found by email — link telegram_id
			user.TelegramID = &telegramID
			if firstName != "" && user.FirstName == nil {
				user.FirstName = &firstName
			}
			if lastName != "" && user.LastName == nil {
				user.LastName = &lastName
			}
			if username != "" && user.Username == nil {
				user.Username = &username
			}
			if err := s.userRepo.Update(ctx, user); err != nil {
				return nil, fmt.Errorf("link telegram to email account: %w", err)
			}
			s.log.Info("linked telegram to existing email account",
				zap.Int("userID", user.ID), zap.Int64("telegramID", telegramID), zap.String("email", email))
			return user, nil
		}
		if !errors.Is(err, domain.ErrUserNotFound) {
			return nil, fmt.Errorf("find by email: %w", err)
		}
	}

	// 3. Create new user
	user = &domain.User{
		TelegramID: &telegramID,
		FirstName:  &firstName,
		LastName:   &lastName,
		Username:   &username,
		Role:       "customer",
		IsActive:   true,
	}
	if email != "" {
		user.Email = &email
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create telegram user: %w", err)
	}
	s.log.Info("created user from telegram", zap.Int("userID", user.ID), zap.Int64("telegramID", telegramID))
	return user, nil
}

func sha256Sum(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
