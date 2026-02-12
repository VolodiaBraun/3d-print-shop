package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/brown/3d-print-shop/internal/domain"
)

var (
	ErrLoyaltyInactive      = errors.New("loyalty program is not active")
	ErrAlreadyReferred      = errors.New("user already has a referrer")
	ErrReferralCodeNotFound = errors.New("referral code not found")
	ErrSelfReferral         = errors.New("cannot use own referral code")
	ErrInsufficientBonuses  = errors.New("insufficient bonus balance")
)

type LoyaltyService struct {
	userRepo     domain.UserRepository
	bonusRepo    domain.BonusTransactionRepository
	settingsRepo domain.LoyaltySettingsRepository
	db           *gorm.DB
	log          *zap.Logger
}

func NewLoyaltyService(
	userRepo domain.UserRepository,
	bonusRepo domain.BonusTransactionRepository,
	settingsRepo domain.LoyaltySettingsRepository,
	db *gorm.DB,
	log *zap.Logger,
) *LoyaltyService {
	return &LoyaltyService{
		userRepo:     userRepo,
		bonusRepo:    bonusRepo,
		settingsRepo: settingsRepo,
		db:           db,
		log:          log,
	}
}

// GetOrCreateReferralCode returns the user's referral code, generating one if needed.
func (s *LoyaltyService) GetOrCreateReferralCode(ctx context.Context, userID int) (string, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return "", err
	}

	if user.ReferralCode != nil && *user.ReferralCode != "" {
		return *user.ReferralCode, nil
	}

	// Generate unique code
	for i := 0; i < 10; i++ {
		code := generateReferralCode()
		user.ReferralCode = &code
		if err := s.userRepo.Update(ctx, user); err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
				continue // retry with new code
			}
			return "", fmt.Errorf("save referral code: %w", err)
		}
		s.log.Info("referral code generated", zap.Int("userID", userID), zap.String("code", code))
		return code, nil
	}

	return "", fmt.Errorf("failed to generate unique referral code after 10 attempts")
}

// ApplyReferralCode links a user to a referrer and credits welcome bonus.
func (s *LoyaltyService) ApplyReferralCode(ctx context.Context, userID int, code string) error {
	code = strings.TrimSpace(strings.ToUpper(code))
	if code == "" {
		return ErrReferralCodeNotFound
	}

	settings, err := s.settingsRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("get settings: %w", err)
	}
	if !settings.IsActive {
		return ErrLoyaltyInactive
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.ReferredByUserID != nil {
		return ErrAlreadyReferred
	}

	referrer, err := s.userRepo.FindByReferralCode(ctx, code)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return ErrReferralCodeNotFound
		}
		return fmt.Errorf("find referrer: %w", err)
	}
	if referrer.ID == userID {
		return ErrSelfReferral
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Link user to referrer
		if err := tx.Model(&domain.User{}).Where("id = ?", userID).
			Updates(map[string]interface{}{
				"referred_by_user_id": referrer.ID,
			}).Error; err != nil {
			return fmt.Errorf("link referrer: %w", err)
		}

		// Credit welcome bonus to the new user
		if settings.ReferralWelcomeBonus > 0 {
			if err := tx.Model(&domain.User{}).Where("id = ?", userID).
				UpdateColumn("bonus_balance", gorm.Expr("bonus_balance + ?", settings.ReferralWelcomeBonus)).Error; err != nil {
				return fmt.Errorf("credit welcome bonus: %w", err)
			}

			desc := fmt.Sprintf("Приветственный бонус по реферальному коду %s", code)
			bonusTx := &domain.BonusTransaction{
				UserID:      userID,
				Amount:      settings.ReferralWelcomeBonus,
				Type:        "referral_welcome",
				ReferenceID: &referrer.ID,
				Description: &desc,
			}
			if err := tx.Create(bonusTx).Error; err != nil {
				return fmt.Errorf("create welcome bonus tx: %w", err)
			}
		}

		s.log.Info("referral code applied",
			zap.Int("userID", userID),
			zap.Int("referrerID", referrer.ID),
			zap.String("code", code),
			zap.Float64("welcomeBonus", settings.ReferralWelcomeBonus),
		)
		return nil
	})
}

// CreditReferrerBonus credits the referrer when a referred user's order is delivered.
func (s *LoyaltyService) CreditReferrerBonus(ctx context.Context, order *domain.Order) {
	if order == nil || order.UserID == nil {
		return
	}

	settings, err := s.settingsRepo.Get(ctx)
	if err != nil || !settings.IsActive || settings.ReferrerBonusPercent <= 0 {
		return
	}

	user, err := s.userRepo.FindByID(ctx, *order.UserID)
	if err != nil || user.ReferredByUserID == nil {
		return
	}

	referrerID := *user.ReferredByUserID
	bonus := math.Round(order.TotalPrice*settings.ReferrerBonusPercent) / 100

	if bonus <= 0 {
		return
	}

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&domain.User{}).Where("id = ?", referrerID).
			UpdateColumn("bonus_balance", gorm.Expr("bonus_balance + ?", bonus)).Error; err != nil {
			return err
		}

		desc := fmt.Sprintf("Бонус за заказ #%s реферала", order.OrderNumber)
		orderID := order.ID
		bonusTx := &domain.BonusTransaction{
			UserID:      referrerID,
			Amount:      bonus,
			Type:        "referral_reward",
			ReferenceID: &orderID,
			Description: &desc,
		}
		return tx.Create(bonusTx).Error
	})

	if err != nil {
		s.log.Warn("failed to credit referrer bonus", zap.Error(err), zap.Int("referrerID", referrerID))
		return
	}

	s.log.Info("referrer bonus credited",
		zap.Int("referrerID", referrerID),
		zap.Float64("bonus", bonus),
		zap.String("orderNumber", order.OrderNumber),
	)
}

// DeductBonuses deducts bonuses from user balance within a GORM transaction.
func (s *LoyaltyService) DeductBonuses(ctx context.Context, tx *gorm.DB, userID int, amount float64, orderID int) error {
	if amount <= 0 {
		return nil
	}

	// Check balance
	var user domain.User
	if err := tx.WithContext(ctx).First(&user, userID).Error; err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	if user.BonusBalance < amount {
		return ErrInsufficientBonuses
	}

	// Deduct
	if err := tx.Model(&domain.User{}).Where("id = ? AND bonus_balance >= ?", userID, amount).
		UpdateColumn("bonus_balance", gorm.Expr("bonus_balance - ?", amount)).Error; err != nil {
		return fmt.Errorf("deduct bonus: %w", err)
	}

	desc := "Списание бонусов при оформлении заказа"
	bonusTx := &domain.BonusTransaction{
		UserID:      userID,
		Amount:      -amount,
		Type:        "order_deduction",
		ReferenceID: &orderID,
		Description: &desc,
	}
	if err := tx.Create(bonusTx).Error; err != nil {
		return fmt.Errorf("create deduction tx: %w", err)
	}

	return nil
}

// ReferralInfoResponse is returned by GetReferralInfo.
type ReferralInfoResponse struct {
	ReferralCode   string  `json:"referralCode"`
	ReferralLink   string  `json:"referralLink"`
	ReferralsCount int     `json:"referralsCount"`
	BonusBalance   float64 `json:"bonusBalance"`
}

// GetReferralInfo returns referral info for a user.
func (s *LoyaltyService) GetReferralInfo(ctx context.Context, userID int) (*ReferralInfoResponse, error) {
	code, err := s.GetOrCreateReferralCode(ctx, userID)
	if err != nil {
		return nil, err
	}

	count, err := s.userRepo.CountByReferrer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("count referrals: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &ReferralInfoResponse{
		ReferralCode:   code,
		ReferralLink:   fmt.Sprintf("https://avangard-print.ru/register?ref=%s", code),
		ReferralsCount: count,
		BonusBalance:   user.BonusBalance,
	}, nil
}

// GetBonusHistory returns bonus transactions for a user.
func (s *LoyaltyService) GetBonusHistory(ctx context.Context, userID int) ([]domain.BonusTransaction, error) {
	return s.bonusRepo.ListByUserID(ctx, userID)
}

// GetSettings returns loyalty settings.
func (s *LoyaltyService) GetSettings(ctx context.Context) (*domain.LoyaltySettings, error) {
	return s.settingsRepo.Get(ctx)
}

// UpdateSettingsInput is the input for updating loyalty settings.
type UpdateSettingsInput struct {
	ReferrerBonusPercent *float64 `json:"referrerBonusPercent"`
	ReferralWelcomeBonus *float64 `json:"referralWelcomeBonus"`
	IsActive             *bool    `json:"isActive"`
}

// UpdateSettings updates loyalty settings.
func (s *LoyaltyService) UpdateSettings(ctx context.Context, input UpdateSettingsInput) (*domain.LoyaltySettings, error) {
	settings, err := s.settingsRepo.Get(ctx)
	if err != nil {
		return nil, err
	}

	if input.ReferrerBonusPercent != nil {
		settings.ReferrerBonusPercent = *input.ReferrerBonusPercent
	}
	if input.ReferralWelcomeBonus != nil {
		settings.ReferralWelcomeBonus = *input.ReferralWelcomeBonus
	}
	if input.IsActive != nil {
		settings.IsActive = *input.IsActive
	}

	if err := s.settingsRepo.Update(ctx, settings); err != nil {
		return nil, err
	}

	s.log.Info("loyalty settings updated",
		zap.Float64("referrerPercent", settings.ReferrerBonusPercent),
		zap.Float64("welcomeBonus", settings.ReferralWelcomeBonus),
		zap.Bool("isActive", settings.IsActive),
	)
	return settings, nil
}

// generateReferralCode creates a random code like "REF-A1B2C3".
func generateReferralCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no O/0/1/I to avoid confusion
	b := make([]byte, 6)
	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[n.Int64()]
	}
	return "REF-" + string(b)
}
