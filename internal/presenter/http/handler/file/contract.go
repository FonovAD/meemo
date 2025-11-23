package file

type SaveFileMetadata struct {
	UserID       int64  `json:"user_id"`
	UserEmail    string `json:"user_email"`
	OriginalName string `json:"original_name"`
	MimeType     string `json:"mime_type"`
	SizeInBytes  int64  `json:"size_in_bytes"`
	S3Bucket     string `json:"s3_bucket"`
	S3Key        string `json:"s3_key"`
	Status       int    `json:"status"`
	IsPublic     bool   `json:"is_public"`
}

//curl -X POST http://localhost:8080/files/metadata \
//-H "Content-Type: application/json" \
//-d '{
//"user_id": 12345,
//"user_email": "aleksandra@example.com",
//"original_name": "report.pdf",
//"mime_type": "application/pdf",
//"size_in_bytes": 102400,
//"s3_bucket": "bucket",
//"s3_key": "uploads/report.pdf",
//"status": 1,
//"is_public": false
//}'
