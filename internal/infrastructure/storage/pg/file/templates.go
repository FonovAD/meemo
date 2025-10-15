package file

const (
	SaveFileTemplate = `
INSERT INTO files (user_id, original_name, mime_type, size_in_bytes, s3_bucket, s3_key, status, created_at, updated_at, is_public)
VALUES (:user_id, :original_name, :mime_type, :size_in_bytes, :s3_bucket, :s3_key, :status, :created_at, :updated_at, :is_public) 
RETURNING id;`

	DeleteFileTemplate = `
DELETE FROM files f
USING users u
WHERE f.user_id = u.id 
  AND u.email = $1 
  AND f.original_name = $2
RETURNING f.id;`

	// TODO: убрать *

	GetFileTemplate = `
SELECT f.*
FROM files f
INNER JOIN users u ON f.user_id = u.id
WHERE u.email = $1 
  AND f.original_name = $2;`

	ChangeVisibilityTemplate = `
UPDATE files f
SET is_public = $1
FROM users u
WHERE f.user_id = u.id 
  AND u.email = $2 
  AND f.original_name = $3
RETURNING f.id, f.is_public;`

	SetStatusTemplate = `
UPDATE files f
SET status = $1
FROM users u
WHERE f.user_id = u.id 
  AND u.email = $2 
  AND f.original_name = $3
RETURNING f.id, f.status;`
)
