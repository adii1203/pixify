package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/adii1203/pixify/db"
	"github.com/adii1203/pixify/handler"
	"github.com/adii1203/pixify/storage"
	"github.com/adii1203/pixify/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {

	// initializing database
	db := db.Init()

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
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// health check for uptime robot
	e.HEAD("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	e.POST("/api/upload", handler.HandlePutImage(S3, db))
	e.GET("/images/:key", handler.HandelGetImage())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// start server
	go func() {
		if err := e.Start(":8000"); err != nil && err != http.ErrServerClosed {
			log.Print(err.Error())
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
