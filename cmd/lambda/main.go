package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/adii1203/pixify/storage"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, request map[string]interface{}) (events.APIGatewayProxyResponse, error) {
	s3client, err := storage.NewS3Client()
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error: initializing s3 client",
		}, err
	}

	path := request["rawPath"]
	objectKey := path.(string)[1:]

	res, err := s3client.S3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String("pixify-raw-images-bucket"),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error: getting object from s3",
		}, err
	}

	defer res.Body.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, res.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Error: %v", err.Error()),
		}, err
	}

	_, err = s3client.S3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:        aws.String("pixify-transformed-images-bucket"),
		Key:           aws.String(objectKey),
		Body:          bytes.NewReader(buf.Bytes()),
		ContentLength: res.ContentLength,
		ContentType:   res.ContentType,
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("Error: %v", err.Error()),
		}, err
	}

	encodeedBody := base64.StdEncoding.EncodeToString(buf.Bytes())

	// p := request["rawPath"]
	// q := request["rawQueryString"]

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": *res.ContentType,
		},
		Body:            encodeedBody,
		IsBase64Encoded: true,
	}, nil
}
