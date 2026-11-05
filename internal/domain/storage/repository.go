package storage

import (
	"context"
	"time"
)

// FileRepository represent interface for storing files
type FileRepository interface {
	GeneratePresignedURL(ctx context.Context, key string, contentType string, expiresIn time.Duration) (string, error)
	DeleteFile(ctx context.Context, key string) error
	FileExists(ctx context.Context, key string) bool
	GetPresignedURL(key string) string
	GetSignedURL(ctx context.Context, key string, expiresIn time.Duration) (string, error)
}
