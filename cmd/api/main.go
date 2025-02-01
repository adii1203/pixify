package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/adii1203/pixify/handler"
	"github.com/adii1203/pixify/storage"
	"github.com/adii1203/pixify/utils"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	// loading environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	// initializing s3 client
	fmt.Println("initializing s3 client...")
	S3, err := storage.NewS3Client()
	if err != nil {
		log.Printf("error : %v", err.Error())
	}

	// initializing echo server
	e := echo.New()
	e.Validator = utils.NewValidator(context.TODO())
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10,
		LogLevel:  log.ERROR,
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latencu=${latency}\n",
	}))

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World!")
	})

	e.POST("/api/upload", handler.HandelPutObjectPresignURL(S3))
	e.GET("/images/:key", handler.HandelGetImage())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// start server
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// gracefully shut down the server with a timeout
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}
