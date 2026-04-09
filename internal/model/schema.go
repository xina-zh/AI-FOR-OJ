package model

import "gorm.io/gorm"

func AutoMigrate(db *gorm.DB) error {
	return db.
		Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci").
		AutoMigrate(AllModels()...)
}

func AllModels() []any {
	return []any{
		&Problem{},
		&TestCase{},
		&Submission{},
		&JudgeResult{},
		&SubmissionTestCaseResult{},
		&Experiment{},
		&ExperimentConfig{},
		&ExperimentRun{},
		&TraceEvent{},
	}
}
