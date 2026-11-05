package file

import (
	"context"
	"strings"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/storage"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

const (
	ExpirePresignedURL = 365 * 24 * time.Hour // 1 year
	ExpiresDatetime    = 1 * time.Hour        // 1 Hour
)

// FileService handles file upload operations
type FileService struct {
	storageRepo   storage.FileRepository
	profileRepo   profiles.ProfileRepository
	presignExpiry time.Duration
}

// NewFileService creates a new FileService instance
func NewFileService(storageRepo storage.FileRepository, profileRepo profiles.ProfileRepository) *FileService {
	return &FileService{
		storageRepo:   storageRepo,
		profileRepo:   profileRepo,
		presignExpiry: ExpiresDatetime,
	}
}

// GeneratePresignedURL generates a presigned URL for file upload
func (s *FileService) GeneratePresignedURL(ctx context.Context, req *PresignRequest) (*PresignResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	// Validate file metadata
	if err := ValidateFileMetadata(req.FileName, req.FileSize, req.ContentType); err != nil {
		return nil, apperror.NewBadRequestError(err.Error(), "FILE_VALIDATION_ERROR", nil)
	}

	// Generate unique filename with UUID
	uniqueFileName := GenerateUniqueFileName(req.FileName)

	// Generate storage key
	key := GenerateStorageKey(uniqueFileName)

	// Generate presigned URL via storage repository
	presignedURL, err := s.storageRepo.GeneratePresignedURL(ctx, key, req.ContentType, s.presignExpiry)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to generate presigned URL", "PRESIGN_URL_ERROR", nil)
	}

	// Return presigned URL and key
	return &PresignResponse{
		PresignedURL:    presignedURL,
		Key:             key,
		ExpiresDatetime: time.Now().Add(s.presignExpiry),
	}, nil
}

// UploadFileS3 completes the upload process and updates the profile
func (s *FileService) UploadFileS3(ctx context.Context, req *UploadRequest) (*UploadResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	if req.Key == "" {
		return nil, apperror.NewBadRequestError("key is required", "INVALID_KEY", nil)
	}

	var fileURL string
	signedURL, err := s.storageRepo.GetSignedURL(ctx, req.Key, ExpirePresignedURL)
	if err != nil {
		fileURL = s.storageRepo.GetPresignedURL(req.Key)
	} else {
		fileURL = signedURL
	}

	// Get old avatar_path from profile
	existingProfile, err := s.profileRepo.GetProfile(ctx, req.AccountID, req.NodeID)
	if err != nil {
		return nil, apperror.NewNotFoundError("profile not found", "PROFILE_NOT_FOUND", nil)
	}

	// Store old avatar path for cleanup
	var oldAvatarPath *string
	if existingProfile.AvatarPath != nil && *existingProfile.AvatarPath != "" {
		oldAvatarPath = existingProfile.AvatarPath
	}

	// Update profile with new avatar_path
	existingProfile.AvatarPath = &fileURL
	existingProfile.UpdatedAt = time.Now()

	err = s.profileRepo.UpdateProfile(ctx, existingProfile, req.NodeID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to update profile", "UPDATE_PROFILE_ERROR", nil)
	}

	// Delete old avatar
	if oldAvatarPath != nil && s.isSupabaseStorageURL(*oldAvatarPath) {
		oldKey := s.extractKeyFromURL(*oldAvatarPath)
		if oldKey != "" {
			_ = s.storageRepo.DeleteFile(ctx, oldKey)
		}
	}

	// Return file URL
	return &UploadResponse{
		URL:       fileURL,
		UpdatedAt: existingProfile.UpdatedAt,
	}, nil
}

// isSupabaseStorageURL checks if the URL is from Supabase Storage
func (s *FileService) isSupabaseStorageURL(url string) bool {
	// Check if URL contains storage path pattern
	return strings.Contains(url, "/storage/v1/object/") || strings.Contains(url, "avatars/")
}

// extractKeyFromURL extracts the storage key from a Supabase Storage URL
func (s *FileService) extractKeyFromURL(url string) string {
	if strings.Contains(url, "/storage/v1/object/public/") {
		parts := strings.Split(url, "/storage/v1/object/public/")
		if len(parts) == 2 {
			// Remove bucket name and get the key
			keyParts := strings.SplitN(parts[1], "/", 2)
			if len(keyParts) == 2 {
				return keyParts[1]
			}
		}
	}

	// Fallback: if URL contains avatars/ pattern, extract from there
	if strings.Contains(url, "avatars/") {
		idx := strings.Index(url, "avatars/")
		key := url[idx:]
		if qIdx := strings.Index(key, "?"); qIdx != -1 {
			key = key[:qIdx]
		}
		return key
	}

	return ""
}
