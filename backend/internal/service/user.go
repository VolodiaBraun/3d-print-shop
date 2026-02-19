package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/cache"
	"github.com/brown/3d-print-shop/internal/domain"
)

var (
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrVerificationCodeExpired = errors.New("verification code expired")
	ErrNoEmailToVerify         = errors.New("no email to verify")
)

const verifyCodeTTL = 15 * time.Minute

type UserService struct {
	userRepo     domain.UserRepository
	cache        *cache.Store
	emailService *EmailService
	log          *zap.Logger
}

func NewUserService(userRepo domain.UserRepository, log *zap.Logger) *UserService {
	return &UserService{userRepo: userRepo, log: log}
}

func (s *UserService) SetCache(c *cache.Store) {
	s.cache = c
}

func (s *UserService) SetEmailService(es *EmailService) {
	s.emailService = es
}

type ProfileResponse struct {
	ID            int     `json:"id"`
	Email         string  `json:"email,omitempty"`
	EmailVerified bool    `json:"emailVerified"`
	Phone         string  `json:"phone,omitempty"`
	FirstName     string  `json:"firstName,omitempty"`
	LastName      string  `json:"lastName,omitempty"`
	Username      string  `json:"username,omitempty"`
	TelegramID    int64   `json:"telegramId,omitempty"`
	Role          string  `json:"role"`
	ReferralCode  string  `json:"referralCode,omitempty"`
	BonusBalance  float64 `json:"bonusBalance"`
	CreatedAt     string  `json:"createdAt"`
}

func (s *UserService) GetProfile(ctx context.Context, userID int) (*ProfileResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return userToProfile(user), nil
}

type UpdateProfileInput struct {
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Phone     *string `json:"phone"`
	Email     *string `json:"email"`
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int, input UpdateProfileInput) (*ProfileResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	emailChanged := false

	// Check email uniqueness if changing
	if input.Email != nil && *input.Email != "" {
		currentEmail := ""
		if user.Email != nil {
			currentEmail = *user.Email
		}
		if *input.Email != currentEmail {
			existing, err := s.userRepo.FindByEmail(ctx, *input.Email)
			if err == nil && existing.ID != userID {
				return nil, domain.ErrEmailAlreadyExists
			}
			if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
				return nil, fmt.Errorf("check email: %w", err)
			}
			user.Email = input.Email
			user.EmailVerified = false
			emailChanged = true
		}
	}

	if input.FirstName != nil {
		user.FirstName = input.FirstName
	}
	if input.LastName != nil {
		user.LastName = input.LastName
	}
	if input.Phone != nil {
		user.Phone = input.Phone
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	s.log.Info("profile updated", zap.Int("userID", userID))

	// Auto-send verification code when email changed
	if emailChanged {
		go s.sendVerificationCode(context.Background(), userID, *input.Email)
	}

	return userToProfile(user), nil
}

// SendVerificationCode generates and sends a 6-digit code to user's email.
func (s *UserService) SendVerificationCode(ctx context.Context, userID int) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if user.Email == nil || *user.Email == "" {
		return ErrNoEmailToVerify
	}

	if user.EmailVerified {
		return nil // already verified
	}

	return s.sendVerificationCode(ctx, userID, *user.Email)
}

func (s *UserService) sendVerificationCode(ctx context.Context, userID int, email string) error {
	if s.cache == nil || s.emailService == nil {
		return nil
	}

	code := generateCode()
	cacheKey := fmt.Sprintf("email_verify:%d", userID)

	data := map[string]string{
		"code":  code,
		"email": email,
	}

	if err := s.cache.Set(ctx, cacheKey, data, verifyCodeTTL); err != nil {
		s.log.Warn("failed to store verification code", zap.Error(err))
		return err
	}

	s.emailService.SendVerificationCode(email, code)
	s.log.Info("verification code sent", zap.Int("userID", userID), zap.String("email", email))
	return nil
}

// ConfirmVerificationCode checks the 6-digit code and marks email as verified.
func (s *UserService) ConfirmVerificationCode(ctx context.Context, userID int, code string) error {
	if s.cache == nil {
		return ErrVerificationCodeExpired
	}

	cacheKey := fmt.Sprintf("email_verify:%d", userID)

	var data map[string]string
	found, err := s.cache.Get(ctx, cacheKey, &data)
	if err != nil || !found {
		return ErrVerificationCodeExpired
	}

	if data["code"] != code {
		return ErrInvalidVerificationCode
	}

	// Code matches — verify email
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	// Make sure email hasn't changed since code was sent
	if user.Email == nil || *user.Email != data["email"] {
		return ErrVerificationCodeExpired
	}

	user.EmailVerified = true
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	// Clean up code
	_ = s.cache.Delete(ctx, cacheKey)

	s.log.Info("email verified", zap.Int("userID", userID), zap.String("email", *user.Email))
	return nil
}

func generateCode() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return fmt.Sprintf("%06d", n.Int64())
}

func userToProfile(u *domain.User) *ProfileResponse {
	p := &ProfileResponse{
		ID:            u.ID,
		EmailVerified: u.EmailVerified,
		Role:          u.Role,
		CreatedAt:     u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if u.Email != nil {
		p.Email = *u.Email
	}
	if u.Phone != nil {
		p.Phone = *u.Phone
	}
	if u.FirstName != nil {
		p.FirstName = *u.FirstName
	}
	if u.LastName != nil {
		p.LastName = *u.LastName
	}
	if u.Username != nil {
		p.Username = *u.Username
	}
	if u.TelegramID != nil {
		p.TelegramID = *u.TelegramID
	}
	if u.ReferralCode != nil {
		p.ReferralCode = *u.ReferralCode
	}
	p.BonusBalance = u.BonusBalance
	return p
}
