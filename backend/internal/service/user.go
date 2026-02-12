package service

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/brown/3d-print-shop/internal/domain"
)

type UserService struct {
	userRepo domain.UserRepository
	log      *zap.Logger
}

func NewUserService(userRepo domain.UserRepository, log *zap.Logger) *UserService {
	return &UserService{userRepo: userRepo, log: log}
}

type ProfileResponse struct {
	ID           int     `json:"id"`
	Email        string  `json:"email,omitempty"`
	Phone        string  `json:"phone,omitempty"`
	FirstName    string  `json:"firstName,omitempty"`
	LastName     string  `json:"lastName,omitempty"`
	Username     string  `json:"username,omitempty"`
	TelegramID   int64   `json:"telegramId,omitempty"`
	Role         string  `json:"role"`
	ReferralCode string  `json:"referralCode,omitempty"`
	BonusBalance float64 `json:"bonusBalance"`
	CreatedAt    string  `json:"createdAt"`
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
	return userToProfile(user), nil
}

func userToProfile(u *domain.User) *ProfileResponse {
	p := &ProfileResponse{
		ID:        u.ID,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
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
