DROP INDEX IF EXISTS idx_users_referral_code;
ALTER TABLE users
  DROP COLUMN IF EXISTS bonus_balance,
  DROP COLUMN IF EXISTS referred_by_user_id,
  DROP COLUMN IF EXISTS referral_code;
