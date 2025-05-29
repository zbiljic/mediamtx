package backup

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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

	// Initialize minio client object.
	s3Client, err := minio.New(b.S3Bucket, &minio.Options{
		Creds:  credentials.NewStaticV4(b.S3AccessKey, b.S3SecretKey, ""),
		Region: b.S3Region,
	})
	if err != nil {
		return fmt.Errorf("unable to initialize minio client, %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("unable to open file %q, %v", filePath, err)
	}
	defer file.Close()

	_, err = s3Client.PutObject(context.TODO(), b.S3Bucket, filePath, file, -1, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("unable to upload %q to S3 bucket %q, %v", filePath, b.S3Bucket, err)
	}

	log.Printf("Successfully uploaded %q to S3 bucket %q\n", filePath, b.S3Bucket)
	return nil
}
