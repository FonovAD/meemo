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

	GetFileTemplate = `
SELECT f.id, f.user_id, f.original_name, f.mime_type, f.size_in_bytes, 
       f.s3_bucket, f.s3_key, f.status, f.created_at, f.updated_at, f.is_public
FROM files f
INNER JOIN users u ON f.user_id = u.id
WHERE f.id = $1`

	GetFileByOriginalNameAndUserEmailTemplate = `
SELECT f.id, f.user_id, f.original_name, f.mime_type, f.size_in_bytes, 
       f.s3_bucket, f.s3_key, f.status, f.created_at, f.updated_at, f.is_public
FROM files f
INNER JOIN users u ON f.user_id = u.id
WHERE u.email = $1 AND f.original_name = $2`

	ChangeVisibilityTemplate = `
UPDATE files f
SET is_public = $1, updated_at = CURRENT_TIMESTAMP
FROM users u
WHERE f.user_id = u.id 
  AND u.email = $2 
  AND f.original_name = $3
RETURNING f.id, f.is_public, f.updated_at;`

	SetStatusTemplate = `
UPDATE files f
SET status = $1, updated_at = CURRENT_TIMESTAMP
FROM users u
WHERE f.user_id = u.id 
  AND u.email = $2 
  AND f.original_name = $3
RETURNING f.id, f.status, f.updated_at;`

	RenameFileTemplate = `
UPDATE files f
SET original_name = $1, updated_at = CURRENT_TIMESTAMP
FROM users u
WHERE f.user_id = u.id 
  AND u.email = $2 
  AND f.original_name = $3
RETURNING f.id, f.original_name, f.updated_at;`

	ListUserFilesTemplate = `
SELECT f.id, f.user_id, f.original_name, f.mime_type, f.size_in_bytes, 
       f.s3_bucket, f.s3_key, f.status, f.created_at, f.updated_at, f.is_public
FROM files f
INNER JOIN users u ON f.user_id = u.id
WHERE u.email = $1
ORDER BY f.created_at DESC;`

	GetTotalUsedSpaceTemplate = `
SELECT COALESCE(SUM(f.size_in_bytes), 0)
FROM files f
INNER JOIN users u ON f.user_id = u.id
WHERE u.email = $1;`
)
