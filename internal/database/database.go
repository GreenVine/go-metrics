// Package database contains methods and models to interact with the database.
package database

import (
	"context"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// gormDB is the database handle.
var gormDB *gorm.DB

var sqlitePragmas = []string{
	"PRAGMA foreign_keys = ON;",
	"PRAGMA busy_timeout = 10000;",
	"PRAGMA journal_mode = WAL;",
	"PRAGMA synchronous = NORMAL;",
	"PRAGMA cache_size = 1000;",
}

// Init sets up the database connection and migrates the schema.
func Init(dbPath string) error {
	var err error
	if gormDB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DefaultTransactionTimeout: 1 * time.Minute,
		Logger:                    logger.Default.LogMode(logger.Silent),
		TranslateError:            true,
	}); err != nil {
		return err
	}

	// Run schema migrations.
	err = gormDB.AutoMigrate(&DeviceRecord{}, &MetricRecord{}, &ConfigRecord{}, &AlertRecord{})
	if err != nil {
		return err
	}

	for _, pragma := range sqlitePragmas {
		if result := gormDB.Exec(pragma); result.Error != nil {
			return result.Error
		}
	}

	log.Printf("Database initialised successfully: %q", dbPath)

	return nil
}

// WithContext returns a gormDB handle with the provided context.
func WithContext(ctx context.Context) *gorm.DB {
	return gormDB.WithContext(ctx)
}
