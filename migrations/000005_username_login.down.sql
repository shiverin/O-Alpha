ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
ALTER TABLE users RENAME COLUMN username TO email;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
