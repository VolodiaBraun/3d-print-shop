ALTER TABLE users ADD COLUMN email_verified BOOLEAN NOT NULL DEFAULT FALSE;

-- Mark existing users with email as verified (they registered via email)
UPDATE users SET email_verified = TRUE WHERE email IS NOT NULL AND password_hash != '';
