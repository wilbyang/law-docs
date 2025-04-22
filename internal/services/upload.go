package services

import (
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Uploader struct {
	S3Client *s3.Client
	Bucket   string
}

func NewS3Uploader(ctx context.Context, cfg aws.Config, bucketName string) (*S3Uploader, error) {
	s3Client := s3.NewFromConfig(cfg)
	return &S3Uploader{S3Client: s3Client, Bucket: bucketName}, nil
}

func (uploader *S3Uploader) UploadFile(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	_, err = uploader.S3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(uploader.Bucket),
		Key:    aws.String(fileHeader.Filename),
		Body:   file,
	})
	if err != nil {
		slog.Error("Failed to upload file", "error", err, "filePath", fmt.Sprintf("s3://%s/%s", uploader.Bucket, fileHeader.Filename))
		return "", err
	}

	return fmt.Sprintf("s3://%s/%s", uploader.Bucket, fileHeader.Filename), nil
}
