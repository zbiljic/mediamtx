package backup

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type BackupStore struct {
	Path        string
	S3Enabled   bool
	S3Bucket    string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
}

func NewBackupStore(path, bucket, region, accessKey, secretKey string) (*BackupStore, error) {
	// Validate required fields
	if path == "" && bucket == "" {
		return nil, fmt.Errorf("either local path or S3 bucket must be specified")
	}

	return &BackupStore{
		Path:        path,
		S3Enabled:   bucket != "",
		S3Bucket:    bucket,
		S3Region:    region,
		S3AccessKey: accessKey,
		S3SecretKey: secretKey,
	}, nil
}

func (b *BackupStore) UploadToS3(filePath string) error {
	if !b.S3Enabled {
		return fmt.Errorf("S3 upload is not enabled")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(b.S3Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(b.S3AccessKey, b.S3SecretKey, "")),
	)
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open file %q, %v", filePath, err)
	}
	defer file.Close()

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(b.S3Bucket),
		Key:    aws.String(filePath),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("unable to upload %q to S3 bucket %q, %v", filePath, b.S3Bucket, err)
	}

	log.Printf("Successfully uploaded %q to S3 bucket %q\n", filePath, b.S3Bucket)
	return nil
}
