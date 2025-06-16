package storage

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

type Config struct {
	AccessKey string // Access Key, предоставленный провайдером
	SecretKey string // Secret Key
	Region    string // Регион, например, "ru-central1"
	Endpoint  string // Endpoint, например, "https://storage.yandexcloud.net"
	Bucket    string // Имя бакета, куда будут загружаться файлы
}

type S3Client struct {
	svc    *s3.S3
	bucket string
}

func (c *S3Client) PresignGet(key string, expires time.Duration) (string, error) {

	req, _ := c.svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	urlStr, err := req.Presign(expires)
	if err != nil {
		return "", fmt.Errorf("failed to presign GET for %q: %w", key, err)
	}
	return urlStr, nil
}

type GetObjectInput struct {
	Key   string
	Range *string // например "bytes=0-1023"
}

func (c *S3Client) GetObject(ctx context.Context, in *GetObjectInput) (*s3.GetObjectOutput, error) {
	req := &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(in.Key),
	}
	if in.Range != nil {
		req.Range = in.Range
	}
	return c.svc.GetObjectWithContext(ctx, req)
}

func NewS3Client(cfg Config) (*S3Client, error) {
	awsCfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		Region:           aws.String(cfg.Region),
		Endpoint:         aws.String(cfg.Endpoint),
		S3ForcePathStyle: aws.Bool(true), // Часто необходимо для S3-совместимых сервисов
	}
	sess, err := session.NewSession(awsCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %v", err)
	}

	svc := s3.New(sess)
	return &S3Client{
		svc:    svc,
		bucket: cfg.Bucket,
	}, nil
}

func (c *S3Client) UploadObject(objectData []byte, contentType string) (string, error) {
	key := uuid.New().String()

	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(objectData),
		ContentType: aws.String(contentType),
	}

	_, err := c.svc.PutObject(input)
	if err != nil {
		return "", fmt.Errorf("failed to upload object: %v", err)
	}

	return key, nil
}
