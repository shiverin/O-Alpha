ALTER TABLE users RENAME COLUMN email TO username;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
