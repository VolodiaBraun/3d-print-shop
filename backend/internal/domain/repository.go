package domain

import "context"

// UserRepository defines the interface for user data access.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id int) (*User, error)
	FindByTelegramID(ctx context.Context, telegramID int64) (*User, error)
	FindByReferralCode(ctx context.Context, code string) (*User, error)
	CountByReferrer(ctx context.Context, referrerID int) (int, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
}
