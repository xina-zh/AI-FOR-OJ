package repository

import (
	"context"

	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

type AISolveAttemptRepository interface {
	Create(ctx context.Context, attempt *model.AISolveAttempt) error
	ListByRunID(ctx context.Context, runID uint) ([]model.AISolveAttempt, error)
}

type GORMAISolveAttemptRepository struct {
	db *gorm.DB
}

func NewAISolveAttemptRepository(db *gorm.DB) *GORMAISolveAttemptRepository {
	return &GORMAISolveAttemptRepository{db: db}
}

func (r *GORMAISolveAttemptRepository) Create(ctx context.Context, attempt *model.AISolveAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

func (r *GORMAISolveAttemptRepository) ListByRunID(ctx context.Context, runID uint) ([]model.AISolveAttempt, error) {
	var attempts []model.AISolveAttempt
	if err := r.db.WithContext(ctx).
		Where("ai_solve_run_id = ?", runID).
		Order("attempt_no ASC, id ASC").
		Find(&attempts).
		Error; err != nil {
		return nil, err
	}
	return attempts, nil
}
