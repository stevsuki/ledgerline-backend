-- Per-account brute force protection. The per-IP rate limiter does not cover
-- this: an attacker rotating addresses still gets unlimited tries per account.
ALTER TABLE users ADD COLUMN failed_login_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN locked_until TIMESTAMPTZ;
