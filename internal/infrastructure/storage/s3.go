package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	appconfig "github.com/EduGoGroup/edugo-api-mobile-new/internal/config"
)

// S3Client wraps AWS S3 presign operations.
type S3Client struct {
	presigner *s3.PresignClient
	bucket    string
	expiry    time.Duration
}

// NewS3Client creates an S3Client from application config.
func NewS3Client(ctx context.Context, cfg appconfig.S3Config) (*S3Client, error) {
	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("loading AWS config: %w", err)
	}

	var s3Opts []func(*s3.Options)
	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)
	presigner := s3.NewPresignClient(client)

	expiry := cfg.PresignExpiry
	if expiry == 0 {
		expiry = 15 * time.Minute
	}

	return &S3Client{
		presigner: presigner,
		bucket:    cfg.Bucket,
		expiry:    expiry,
	}, nil
}

// GenerateUploadURL creates a presigned PUT URL for uploading an object.
func (c *S3Client) GenerateUploadURL(ctx context.Context, key string) (string, time.Time, error) {
	result, err := c.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(c.expiry))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("presigning PUT: %w", err)
	}
	return result.URL, time.Now().Add(c.expiry), nil
}

// GenerateDownloadURL creates a presigned GET URL for downloading an object.
func (c *S3Client) GenerateDownloadURL(ctx context.Context, key string) (string, time.Time, error) {
	result, err := c.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(c.expiry))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("presigning GET: %w", err)
	}
	return result.URL, time.Now().Add(c.expiry), nil
}
