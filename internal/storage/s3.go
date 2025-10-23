package storage

import (
	"context"
	"io"
	"spotifybackend/internal/config"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client struct {
	Client  *s3.Client
	Presign *s3.PresignClient
	Bucket  string
	UseSSL  bool
	Region  string
}

func NewS3Client(cfg *config.Config) (*S3Client, error) {
	// custom endpoint for MinIO local dev
	customResolver := aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
		if cfg.S3Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.S3Endpoint,
				SigningRegion:     cfg.S3Region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	awsCfg, err := awscfg.LoadDefaultConfig(context.TODO(),
		awscfg.WithRegion(cfg.S3Region),
		awscfg.WithEndpointResolver(customResolver),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		// nothing
	})

	presigner := s3.NewPresignClient(client)

	// ensure bucket exists (best-effort)
	_, _ = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{Bucket: &cfg.S3Bucket})

	return &S3Client{
		Client:  client,
		Presign: presigner,
		Bucket:  cfg.S3Bucket,
		UseSSL:  cfg.S3UseSSL,
		Region:  cfg.S3Region,
	}, nil
}

func (s *S3Client) PresignedPutURL(ctx context.Context, key string, ttl time.Duration, contentType string) (string, error) {
	params := &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	}
	req, err := s.Presign.PresignPutObject(ctx, params, func(po *s3.PresignOptions) {
		po.Expires = ttl
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (s *S3Client) PresignedGetURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(s.Bucket),
		Key:    aws.String(key),
	}
	req, err := s.Presign.PresignGetObject(ctx, params, func(po *s3.PresignOptions) {
		po.Expires = ttl
	})
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// Optional helper to upload small files server side
func (s *S3Client) UploadFromReader(ctx context.Context, key string, body io.ReadSeekCloser, contentType string) error {
	uploader := manager.NewUploader(s.Client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.Bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
		ACL:         types.ObjectCannedACLPrivate,
	})
	return err
}
