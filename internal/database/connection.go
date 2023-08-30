package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseConnection struct {
	DSN string
	DB  *gorm.DB
}

func (d *DatabaseConnection) InitConnection() error {
	db, err := gorm.Open(postgres.Open(d.DSN), &gorm.Config{})
	if err == nil {
		d.DB = db
	}
	return err
}
