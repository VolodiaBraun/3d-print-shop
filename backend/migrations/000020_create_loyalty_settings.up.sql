CREATE TABLE loyalty_settings (
  id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  referrer_bonus_percent DECIMAL(5,2) NOT NULL DEFAULT 5.00,
  referral_welcome_bonus DECIMAL(10,2) NOT NULL DEFAULT 100.00,
  is_active BOOLEAN NOT NULL DEFAULT true,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO loyalty_settings DEFAULT VALUES;
