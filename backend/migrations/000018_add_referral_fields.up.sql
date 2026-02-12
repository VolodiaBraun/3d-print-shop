ALTER TABLE users
  ADD COLUMN referral_code VARCHAR(12) UNIQUE,
  ADD COLUMN referred_by_user_id INTEGER REFERENCES users(id),
  ADD COLUMN bonus_balance DECIMAL(10,2) NOT NULL DEFAULT 0;

CREATE INDEX idx_users_referral_code ON users(referral_code) WHERE referral_code IS NOT NULL;
