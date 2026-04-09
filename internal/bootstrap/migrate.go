package bootstrap

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"ai-for-oj/internal/config"
	"ai-for-oj/internal/model"
)

// RunMigrations keeps the current project in AutoMigrate mode for fast iteration.
// When the schema stabilizes, this can be replaced with a versioned migration flow
// without changing the rest of the bootstrap path.
func RunMigrations(db *gorm.DB, cfg config.DatabaseConfig, logger *slog.Logger) error {
	if !cfg.AutoMigrate {
		logger.Info("database auto migration disabled")
		return nil
	}

	logger.Info("running database auto migration")
	if err := model.AutoMigrate(db); err != nil {
		return fmt.Errorf("auto migrate schema: %w", err)
	}

	logger.Info("database auto migration completed")
	return nil
}
