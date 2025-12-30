package file

type SaveFileMetadata struct {
	OriginalName string `json:"original_name"`
	IsPublic     bool   `json:"is_public"`
	MimeType     string `json:"mime_type"`
	SizeInBytes  int64  `json:"size_in_bytes"`
}

type RenameFileRequest struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

type ChangeVisibilityRequest struct {
	OriginalName string `json:"original_name"`
	IsPublic     bool   `json:"is_public"`
}

type SetStatusRequest struct {
	OriginalName string `json:"original_name"`
	Status       int    `json:"status"`
}
