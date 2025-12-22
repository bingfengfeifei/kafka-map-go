package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"github.com/ncruces/go-sqlite3/gormlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	_ "github.com/ncruces/go-sqlite3/embed"
)

func InitDB(dbPath string) (*gorm.DB, error) {
	// Ensure data directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database connection using ncruces/go-sqlite3 via gormlite
	db, err := gorm.Open(gormlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate schemas
	if err := db.AutoMigrate(&model.User{}, &model.Cluster{}, &model.TopicStats{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}
