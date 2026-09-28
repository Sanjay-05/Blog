package driver

import (
	"blog/config"
	"blog/models"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB(cfg *config.Configuration) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		return nil, err
	}

	err = CreateTable(db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func CreateTable(db *gorm.DB) error {
	if err := db.AutoMigrate(&models.User{},&models.Category{},&models.Comment{},&models.Post{}); err != nil {
		return err
	}
	return nil
}
