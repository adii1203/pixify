package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/adii1203/pixify/data"
	"github.com/adii1203/pixify/db"
	"github.com/adii1203/pixify/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/h2non/bimg"
	"github.com/labstack/echo/v4"
)

type PutObjectKey struct {
	File *multipart.FileHeader `form:"file" json:"file" validate:"required"`
}

func HandlePutImage(s *storage.BucketBasics, db *db.GormDB) echo.HandlerFunc {
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

		if objectKey.File == nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "file is required",
			})
		}

		f, err := objectKey.File.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Errorf("error opening file: %v", err).Error(),
			})
		}
		defer f.Close()

		buf, err := io.ReadAll(f)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Errorf("error reading file: %v", err).Error(),
			})
		}

		img := bimg.NewImage(buf)
		meta, err := img.Metadata()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Errorf("error getting image size: %v", err).Error(),
			})
		}

		_, err = s.S3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: aws.String("pixify-raw-images-bucket"),
			Key:    aws.String(objectKey.File.Filename),
			Body:   bytes.NewReader(buf),
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Errorf("error uploading file: %v", err).Error(),
			})
		}

		file := data.File{
			Name:     objectKey.File.Filename,
			Path:     fmt.Sprint("/", objectKey.File.Filename),
			Size:     uint64(len(buf)),
			Width:    uint64(meta.Size.Width),
			Height:   uint64(meta.Size.Height),
			MimeType: meta.Type,
			Url:      fmt.Sprintf("https://d28dzh6o5yikbs.cloudfront.net/%s", objectKey.File.Filename),
			FileType: objectKey.File.Header.Get("Content-Type"),
		}
		err = db.InsertFile(&file)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": fmt.Errorf("error saving file: %v", err).Error(),
			})
		}
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
			"file":   file,
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
