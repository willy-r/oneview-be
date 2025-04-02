package config

import (
	"log"
	"oneview-be/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDatabase() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("oneview.sqlite3"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}
	db.AutoMigrate(&model.User{}, &model.Message{})
	return db
}
