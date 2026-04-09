package bootstrap

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ai-for-oj/internal/config"
)

func NewDatabase(cfg config.DatabaseConfig, logger *slog.Logger) (*gorm.DB, *sql.DB, error) {
	retries := cfg.InitMaxRetries
	if retries < 1 {
		retries = 1
	}

	var lastErr error

	for attempt := 1; attempt <= retries; attempt++ {
		db, sqlDB, err := openDatabase(cfg)
		if err == nil {
			return db, sqlDB, nil
		}

		lastErr = err
		logger.Warn("database connection attempt failed",
			"attempt", attempt,
			"max_attempts", retries,
			"error", err,
		)

		if attempt < retries {
			time.Sleep(cfg.InitRetryInterval)
		}
	}

	return nil, nil, fmt.Errorf("connect database after %d attempts: %w", retries, lastErr)
}

func openDatabase(cfg config.DatabaseConfig) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, nil, err
	}

	return db, sqlDB, nil
}
