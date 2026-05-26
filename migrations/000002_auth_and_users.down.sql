-- Drop in reverse order to respect Foreign Key dependencies
DROP TABLE IF EXISTS agent_settings CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS users CASCADE;
