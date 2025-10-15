package model

import (
	"errors"
	"meemo/internal/domain/entity"
	"time"
)

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

func (m *File) ModelToEntity() *entity.File {
	return &entity.File{
		ID:           m.ID,
		UserID:       m.UserID,
		OriginalName: m.OriginalName,
		MimeType:     m.MimeType,
		SizeInBytes:  m.SizeInBytes,
		S3Bucket:     m.S3Bucket,
		S3Key:        m.S3Key,
		Status:       m.Status,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		IsPublic:     m.IsPublic,
	}
}

func (m *File) EntityToModel(entity *entity.File) error {
	if entity == nil {
		return errors.New("entity is nil")
	}
	m.ID = entity.ID
	m.UserID = entity.UserID
	m.OriginalName = entity.OriginalName
	m.MimeType = entity.MimeType
	m.SizeInBytes = entity.SizeInBytes
	m.S3Bucket = entity.S3Bucket
	m.S3Key = entity.S3Key
	m.Status = entity.Status
	m.CreatedAt = entity.CreatedAt
	m.UpdatedAt = entity.UpdatedAt
	m.IsPublic = entity.IsPublic
	return nil
}
