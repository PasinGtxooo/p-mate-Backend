package file

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	MaxFileSize = 5 * 1024 * 1024 // 5MB
)

var (
	AllowedMimeTypes = map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	AllowedExtensions = map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
	}
)

// PresignRequest represents the request to generate a presigned URL
type PresignRequest struct {
	FileName    string `json:"file_name" validate:"required"`
	FileSize    int64  `json:"file_size" validate:"required,gt=0"`
	ContentType string `json:"content_type" validate:"required"`
	AccountID   string `json:"account_id"`
	NodeID      string `json:"node_id"`
}

// PresignResponse represents the response containing the presigned URL
type PresignResponse struct {
	PresignedURL    string    `json:"presigned_url"`
	Key             string    `json:"key"`
	ExpiresDatetime time.Time `json:"expires_datetime"`
}

// UploadRequest represents the request to complete the upload
type UploadRequest struct {
	Key       string `json:"key" validate:"required"`
	AccountID string `json:"account_id"`
	NodeID    string `json:"node_id"`
}

// UploadResponse represents the response after completing the upload
type UploadResponse struct {
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FileValidationError represents a file validation error
type FileValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

func (e *FileValidationError) Error() string {
	return e.Message
}

// ValidateFileMetadata validates file metadata (size, MIME type, extension)
func ValidateFileMetadata(fileName string, fileSize int64, contentType string) error {
	// Validate file size
	if fileSize <= 0 {
		return &FileValidationError{
			Code:    "FILE_SIZE_INVALID",
			Message: "ขนาดไฟล์ไม่ถูกต้อง",
			Field:   "file_size",
		}
	}

	if fileSize > MaxFileSize {
		return &FileValidationError{
			Code:    "FILE_SIZE_EXCEEDED",
			Message: "ไฟล์มีขนาดใหญ่เกินไป กรุณาเลือกไฟล์ที่มีขนาดไม่เกิน 5MB",
			Field:   "file_size",
		}
	}

	// Validate MIME type
	if !AllowedMimeTypes[contentType] {
		return &FileValidationError{
			Code:    "MIME_TYPE_INVALID",
			Message: "ประเภทไฟล์ไม่ถูกต้อง กรุณาเลือกไฟล์รูปภาพ (JPG, PNG, GIF, WebP)",
			Field:   "content_type",
		}
	}

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(fileName))
	expectedMimeType, validExt := AllowedExtensions[ext]
	if !validExt {
		return &FileValidationError{
			Code:    "EXTENSION_INVALID",
			Message: "นามสกุลไฟล์ไม่ถูกต้อง กรุณาเลือกไฟล์รูปภาพ (JPG, PNG, GIF, WebP)",
			Field:   "file_name",
		}
	}

	// Validate extension matches MIME type
	if expectedMimeType != contentType {
		return &FileValidationError{
			Code:    "EXTENSION_MISMATCH",
			Message: "นามสกุลไฟล์ไม่ตรงกับประเภทไฟล์",
			Field:   "file_name",
		}
	}

	return nil
}

// GenerateUniqueFileName generates a unique filename using UUID
func GenerateUniqueFileName(originalFileName string) string {
	ext := strings.ToLower(filepath.Ext(originalFileName))
	uniqueID := uuid.New().String()
	return fmt.Sprintf("%s%s", uniqueID, ext)
}

// GenerateStorageKey generates the full storage key for the file
func GenerateStorageKey(fileName string) string {
	return fmt.Sprintf("avatars/%s", fileName)
}
