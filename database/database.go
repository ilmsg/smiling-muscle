package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabaseWithSqlite(filename string) *gorm.DB {
	return NewDatabase(sqlite.Open(filename))
}

func NewDatabase(dialector gorm.Dialector) *gorm.DB {
	conn, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
