package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/adii1203/pixify/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
)

type PutObjectKey struct {
	File *multipart.FileHeader `form:"file" json:"file" validate:"required"`
}

func HandelPutImage(s *storage.BucketBasics) echo.HandlerFunc {
	return func(c echo.Context) error {
		var objectKey PutObjectKey
		err := c.Bind(&objectKey)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return c.JSON(http.StatusBadRequest, map[string]interface{}{
					"message": "empty body",
				})
			}
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": err,
			})
		}
		err = c.Validate(objectKey)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"message": err.Error(),
			})
		}

		f, err := objectKey.File.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "error file opening",
			})
		}
		f.Close()

		_, err = s.S3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String("pixify-raw-images-bucket"),
			Key:    aws.String(objectKey.File.Filename),
			Body:   f,
		})
		if err != nil {
			log.Fatal(err)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
		})
	}
}

func HandelGetImage() echo.HandlerFunc {
	return func(c echo.Context) error {
		imageKey := c.Param("key")
		url := fmt.Sprintf("https://d28dzh6o5yikbs.cloudfront.net/%s", imageKey)
		res, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
		}

		defer res.Body.Close()

		return c.Stream(200, "image/png", res.Body)
	}
}
