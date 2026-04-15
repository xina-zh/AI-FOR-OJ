package bootstrap

import (
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"ai-for-oj/internal/config"
	"ai-for-oj/internal/model"
)

const (
	targetDBCharset   = "utf8mb4"
	targetDBCollation = "utf8mb4_unicode_ci"
)

var utf8mb4MigrationTables = []string{
	"problems",
	"test_cases",
	"submissions",
	"judge_results",
	"submission_test_case_results",
	"ai_solve_runs",
	"experiments",
	"experiment_repeats",
	"experiment_compares",
	"experiment_configs",
	"experiment_runs",
	"trace_events",
}

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
	if err := ensureDatabaseUTF8MB4(db, cfg.Name, logger); err != nil {
		return fmt.Errorf("ensure database utf8mb4 defaults: %w", err)
	}
	if err := ensureTablesUTF8MB4(db, logger, utf8mb4MigrationTables); err != nil {
		return fmt.Errorf("ensure table utf8mb4 charset: %w", err)
	}
	if err := ensureExperimentProblemIDNullable(db, logger); err != nil {
		return fmt.Errorf("relax experiments.problem_id nullability: %w", err)
	}

	logger.Info("database auto migration completed")
	return nil
}

func ensureDatabaseUTF8MB4(db *gorm.DB, dbName string, logger *slog.Logger) error {
	dbName = strings.TrimSpace(dbName)
	if dbName == "" {
		return nil
	}

	type schemaInfo struct {
		DefaultCharacterSetName string `gorm:"column:DEFAULT_CHARACTER_SET_NAME"`
		DefaultCollationName    string `gorm:"column:DEFAULT_COLLATION_NAME"`
	}

	var info schemaInfo
	if err := db.Raw(
		"SELECT DEFAULT_CHARACTER_SET_NAME, DEFAULT_COLLATION_NAME FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?",
		dbName,
	).Scan(&info).Error; err != nil {
		return fmt.Errorf("inspect database defaults: %w", err)
	}

	if strings.EqualFold(info.DefaultCharacterSetName, targetDBCharset) &&
		strings.EqualFold(info.DefaultCollationName, targetDBCollation) {
		return nil
	}

	if err := db.Exec(
		fmt.Sprintf(
			"ALTER DATABASE %s CHARACTER SET %s COLLATE %s",
			quoteIdentifier(dbName),
			targetDBCharset,
			targetDBCollation,
		),
	).Error; err != nil {
		return err
	}

	if logger != nil {
		logger.Info("updated database defaults to utf8mb4", "database", dbName, "collation", targetDBCollation)
	}
	return nil
}

func ensureTablesUTF8MB4(db *gorm.DB, logger *slog.Logger, tableNames []string) error {
	type tableInfo struct {
		TableCollation string `gorm:"column:TABLE_COLLATION"`
	}

	for _, tableName := range tableNames {
		if !db.Migrator().HasTable(tableName) {
			continue
		}

		var info tableInfo
		if err := db.Raw(
			"SELECT TABLE_COLLATION FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
			tableName,
		).Scan(&info).Error; err != nil {
			return fmt.Errorf("inspect table collation for %s: %w", tableName, err)
		}

		if strings.EqualFold(info.TableCollation, targetDBCollation) {
			continue
		}

		if err := db.Exec(
			fmt.Sprintf(
				"ALTER TABLE %s CONVERT TO CHARACTER SET %s COLLATE %s",
				quoteIdentifier(tableName),
				targetDBCharset,
				targetDBCollation,
			),
		).Error; err != nil {
			return fmt.Errorf("convert table %s to utf8mb4: %w", tableName, err)
		}

		if logger != nil {
			logger.Info("converted legacy table to utf8mb4", "table", tableName, "collation", targetDBCollation)
		}
	}

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

func quoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
