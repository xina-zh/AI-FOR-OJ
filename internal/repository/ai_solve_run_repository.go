package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

var ErrAISolveRunNotFound = errors.New("ai solve run not found")

type AISolveRunRepository interface {
	Create(ctx context.Context, run *model.AISolveRun) error
	Update(ctx context.Context, run *model.AISolveRun) error
	GetByID(ctx context.Context, runID uint) (*model.AISolveRun, error)
}

type GORMAISolveRunRepository struct {
	db *gorm.DB
}

func NewAISolveRunRepository(db *gorm.DB) *GORMAISolveRunRepository {
	return &GORMAISolveRunRepository{db: db}
}

func (r *GORMAISolveRunRepository) Create(ctx context.Context, run *model.AISolveRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *GORMAISolveRunRepository) Update(ctx context.Context, run *model.AISolveRun) error {
	return r.db.WithContext(ctx).Save(run).Error
}

func (r *GORMAISolveRunRepository) GetByID(ctx context.Context, runID uint) (*model.AISolveRun, error) {
	var run model.AISolveRun
	if err := r.db.WithContext(ctx).
		Preload("Attempts", func(db *gorm.DB) *gorm.DB {
			return db.Order("attempt_no ASC, id ASC")
		}).
		First(&run, runID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAISolveRunNotFound
		}
		return nil, err
	}
	return &run, nil
}
