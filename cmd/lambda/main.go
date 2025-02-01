package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Input struct {
	Message string `json:"message"`
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	path := request.Path
	fmt.Println(path)
	response := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "Hello from Lambda@Edge! and Path: " + path,
	}

	return response, nil
}
