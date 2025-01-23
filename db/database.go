package db

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Db *gorm.DB

func connectDatabase() (*gorm.DB, error) {
	dns := os.Getenv("DB_DNS")
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database connection fail: %v", err)
	}

	return db, nil
}

func Init() *gorm.DB {
	if err := godotenv.Load(); err != nil {
		log.Panic("error loading env", err)
	}
	db, err := connectDatabase()
	if err != nil {
		log.Fatal(err)
	}
	return db
}
