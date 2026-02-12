CREATE TABLE bonus_transactions (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id),
  amount DECIMAL(10,2) NOT NULL,
  type VARCHAR(30) NOT NULL,
  reference_id INTEGER,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bonus_transactions_user_id ON bonus_transactions(user_id);
