package entity

import "time"

type File struct {
	ID           int64     `db:"id"`
	UserID       int64     `db:"user_id"`
	OriginalName string    `db:"original_name"`
	MimeType     string    `db:"mime_type"`
	SizeInBytes  int64     `db:"size_in_bytes"`
	S3Bucket     string    `db:"s3_bucket"`
	S3Key        string    `db:"s3_key"`
	Status       int       `db:"status"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	IsPublic     bool      `db:"is_public"`
}
