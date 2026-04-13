package bootstrap

import (
	"fmt"
	"log/slog"
	"strings"

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
	if err := ensureExperimentProblemIDNullable(db, logger); err != nil {
		return fmt.Errorf("relax experiments.problem_id nullability: %w", err)
	}

	logger.Info("database auto migration completed")
	return nil
}

func ensureExperimentProblemIDNullable(db *gorm.DB, logger *slog.Logger) error {
	columnTypes, err := db.Migrator().ColumnTypes(&model.Experiment{})
	if err != nil {
		return fmt.Errorf("inspect experiment columns: %w", err)
	}

	for _, columnType := range columnTypes {
		if !strings.EqualFold(columnType.Name(), "problem_id") {
			continue
		}

		if nullable, ok := columnType.Nullable(); ok && nullable {
			return nil
		}

		if err := db.Exec("ALTER TABLE `experiments` MODIFY COLUMN `problem_id` bigint unsigned NULL").Error; err != nil {
			return err
		}

		if logger != nil {
			logger.Info("relaxed legacy experiments.problem_id column to nullable for batch experiments")
		}
		return nil
	}

	return nil
}
