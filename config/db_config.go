package config

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)


func ConnectDB(db_url string) (*gorm.DB, error) {

    db, err := gorm.Open(postgres.Open(db_url), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
    if err != nil {
        return nil,fmt.Errorf(
            "failed to connect to PostgreSQL ('renet' database): %w",
            err,
        )
    }

	fmt.Printf("Successfully connected to PostgreSQL database\n")

    return db, nil
}


