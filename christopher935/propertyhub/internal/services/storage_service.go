package services

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/nfnt/resize"
)

// StorageService handles file uploads to DigitalOcean Spaces
type StorageService struct {
	s3Client *s3.S3
	bucket   string
	cdnURL   string
}

// NewStorageService creates a new storage service for DigitalOcean Spaces
func NewStorageService() (*StorageService, error) {
	endpoint := os.Getenv("DO_SPACES_ENDPOINT")
	region := os.Getenv("DO_SPACES_REGION")
	bucket := os.Getenv("DO_SPACES_BUCKET")
	accessKey := os.Getenv("DO_SPACES_ACCESS_KEY")
	secretKey := os.Getenv("DO_SPACES_SECRET_KEY")
	cdnURL := os.Getenv("DO_SPACES_CDN_URL")

	if endpoint == "" || region == "" || bucket == "" || accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("missing DigitalOcean Spaces configuration")
	}

	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(false),
	}

	sess, err := session.NewSession(s3Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	return &StorageService{
		s3Client: s3.New(sess),
		bucket:   bucket,
		cdnURL:   cdnURL,
	}, nil
}

// UploadProfilePhoto uploads a profile photo with image processing
func (s *StorageService) UploadProfilePhoto(file multipart.File, filename string, userID int64) (string, error) {
	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Decode image
	img, format, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("invalid image file: %v", err)
	}

	// Resize to 400x400 (profile photo size)
	resized := resize.Resize(400, 400, img, resize.Lanczos3)

	// Encode back to bytes
	var buf bytes.Buffer
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, resized)
	default:
		return "", fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return "", fmt.Errorf("failed to encode image: %v", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	timestamp := time.Now().Unix()
	key := fmt.Sprintf("profiles/%d/avatar_%d%s", userID, timestamp, ext)

	// Upload to Spaces
	_, err = s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf.Bytes()),
		ACL:         aws.String("public-read"),
		ContentType: aws.String(fmt.Sprintf("image/%s", format)),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to Spaces: %v", err)
	}

	// Return CDN URL
	url := fmt.Sprintf("%s/%s", s.cdnURL, key)
	return url, nil
}

// DeleteFile deletes a file from Spaces
func (s *StorageService) DeleteFile(url string) error {
	// Extract key from URL
	key := strings.TrimPrefix(url, s.cdnURL+"/")

	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	return err
}

// ValidateImageFile validates file type and size
func ValidateImageFile(header *multipart.FileHeader) error {
	// Check file size (max 5MB)
	if header.Size > 5*1024*1024 {
		return fmt.Errorf("file size exceeds 5MB limit")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !validExts[ext] {
		return fmt.Errorf("invalid file type, only JPG and PNG allowed")
	}

	return nil
}
