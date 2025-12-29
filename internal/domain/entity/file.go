package entity

import (
	"io"
	"time"
)

type File struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	OriginalName string `json:"original_name"`
	name         string
	MimeType     string    `json:"mime_type"`
	SizeInBytes  int64     `json:"size_in_bytes"`
	S3Bucket     string    `json:"s3_bucket"`
	S3Key        string    `json:"s3_key"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	IsPublic     bool      `json:"is_public"`
	R            io.Reader `json:"-"`
	W            io.Writer `json:"-"`
}
