DROP INDEX IF EXISTS idx_files_is_public;
DROP INDEX IF EXISTS idx_files_created_at;
DROP INDEX IF EXISTS idx_files_status;
DROP INDEX IF EXISTS idx_files_user_id;

DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS files;

DROP TABLE IF EXISTS users;