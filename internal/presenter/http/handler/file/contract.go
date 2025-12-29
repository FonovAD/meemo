package file

type SaveFileMetadata struct {
	OriginalName string `json:"original_name"`
	IsPublic     bool   `json:"is_public"`
	MimeType     string `json:"mime_type"`
	SizeInBytes  int64  `json:"size_in_bytes"`
}
