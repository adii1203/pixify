package storage

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type BucketBasics struct {
	S3Client      *s3.Client
	PresignClient *s3.PresignClient
}

func NewS3Client() (*BucketBasics, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg)
	return &BucketBasics{
		S3Client:      client,
		PresignClient: s3.NewPresignClient(client),
	}, nil
}
