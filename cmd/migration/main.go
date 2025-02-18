package main

import (
	"github.com/adii1203/pixify/data"
	"github.com/adii1203/pixify/db"
)

func main() {
	db := db.Init()
	db.Db.AutoMigrate(&data.File{})
}
