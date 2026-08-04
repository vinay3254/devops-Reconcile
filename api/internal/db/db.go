package db

import (
	"fmt"
	"log"

	"github.com/statuspage/api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established. Running migrations...")
	if err := db.AutoMigrate(&Tenant{}, &User{}, &Monitor{}, &CheckResult{}); err != nil {
		return nil, fmt.Errorf("auto-migration failed: %w", err)
	}
	log.Println("Database migration completed successfully.")

	return db, nil
}
