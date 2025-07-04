// Package database contains methods and models to interact with the database.
package database

import (
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is a shared database handle.
var DB *gorm.DB

// Init sets up the database connection and migrates the schema.
func Init(dbPath string) error {
	var err error
	if DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DefaultTransactionTimeout: 1 * time.Minute,
		Logger:                    logger.Default.LogMode(logger.Info),
		TranslateError:            true,
	}); err != nil {
		return err
	}

	// Run schema migrations.
	err = DB.AutoMigrate(&DeviceRecord{}, &MetricRecord{}, &ConfigRecord{}, &AlertRecord{})
	if err != nil {
		return err
	}

	// Enforce foreign key constraints.
	if result := DB.Exec("PRAGMA foreign_keys = ON"); result.Error != nil {
		return result.Error
	}

	log.Printf("Database initialised successfully: %q", dbPath)

	return nil
}
