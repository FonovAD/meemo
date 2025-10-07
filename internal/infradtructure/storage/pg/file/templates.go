package file

const (
	SaveFileTemplate = `
INSERT INTO file
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
RETURNING id;`

	DeleteFileTemplate = `
DELETE FROM file f
USING users u
WHERE f.user_id = u.id 
  AND u.email = $1 
  AND f.original_name = $2;`

	GetFileTemplate = `
SELECT f.*, u.* 
FROM file f
INNER JOIN users u ON f.user_id = u.id
WHERE u.email = $1 
  AND f.original_name = $2;`

	ChangeVisibilityTemplate = `
UPDATE file f
SET is_public = $1
FROM users u
WHERE f.user_id = u.id 
  AND u.email = $2 
  AND f.original_name = $3;`

	SetStatusTemplate = `
UPDATE file f
SET status = $1
FROM users u
WHERE f.user_id = u.id 
  AND u.email = $2 
  AND f.original_name = $3
RETURNING f.id;`
)
