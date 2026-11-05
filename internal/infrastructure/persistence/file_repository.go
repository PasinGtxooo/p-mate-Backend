package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// FileRepository implements storage.FileRepository using S3-compatible storage
type FileRepository struct {
	s3Client   *s3.Client
	bucketName string
	publicURL  string
	region     string
	endpoint   string
}

// NewFileRepository creates a new instance of FileRepository
func NewFileRepository(endpoint, region, bucketName, accessKeyID, secretAccessKey, publicURL string) (*FileRepository, error) {
	// Load AWS SDK configuration with credentials
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &FileRepository{
		s3Client:   s3Client,
		bucketName: bucketName,
		publicURL:  publicURL,
		region:     region,
		endpoint:   endpoint,
	}, nil
}

// GeneratePresignedURL generates a presigned URL for uploading a file to storage
func (r *FileRepository) GeneratePresignedURL(ctx context.Context, key string, contentType string, expiresIn time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(r.s3Client)

	putObjectInput := &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	presignedRequest, err := presignClient.PresignPutObject(ctx, putObjectInput, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedRequest.URL, nil
}

// DeleteFile deletes a file from storage
func (r *FileRepository) DeleteFile(ctx context.Context, key string) error {
	_, err := r.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// FileExists checks if a file exists in storage
func (r *FileRepository) FileExists(ctx context.Context, key string) bool {
	_, err := r.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})
	return err == nil
}

// GetPresignedURL returns the public URL for accessing a file
func (r *FileRepository) GetPresignedURL(key string) string {
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", r.publicURL, r.bucketName, key)
}

// GetSignedURL generates a signed URL for accessing a file
func (r *FileRepository) GetSignedURL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(r.s3Client)

	getObjectInput := &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	}

	presignedRequest, err := presignClient.PresignGetObject(ctx, getObjectInput, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return presignedRequest.URL, nil
}
