package config

import (
	"log"

	"github.com/glebarez/sqlite"
	"github.com/lesmana/sekolah-karir-applicant-api/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes SQLite database connection and runs AutoMigrate
func InitDB(dbPath string) (*gorm.DB, error) {
	if dbPath == "" {
		dbPath = "applicants.db"
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(sqlite.Open(dbPath), gormConfig)
	if err != nil {
		return nil, err
	}

	// Auto-migrate tables
	if err := db.AutoMigrate(&model.Applicant{}); err != nil {
		return nil, err
	}

	log.Printf("[Database] Connected successfully to SQLite (%s)\n", dbPath)
	return db, nil
}

