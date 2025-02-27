package db

import (
	"fmt"
	"log"
	"os"

	"github.com/adii1203/pixify/data"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Store interface {
	InsertFile(file *data.File) error
}

type GormDB struct {
	Db *gorm.DB
}

func (g *GormDB) InsertFile(file *data.File) error {
	tx := g.Db.Create(file)
	if tx.Error != nil {
		return fmt.Errorf("error inserting file: %v", tx.Error)
	}
	return nil
}

func connectDatabase() (*gorm.DB, error) {
	dns := os.Getenv("DB_DNS")
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database connection fail: %v", err)
	}

	return db, nil
}

func Init() *GormDB {
	if os.Getenv("ENV") == "prod" {

	} else {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file")
		}
	}
	db, err := connectDatabase()
	if err != nil {
		log.Fatal(err)
	}
	return &GormDB{
		Db: db,
	}
}
