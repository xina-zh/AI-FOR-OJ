package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

var ErrExperimentNotFound = errors.New("experiment not found")

type ExperimentRepository interface {
	Create(ctx context.Context, experiment *model.Experiment) error
	Update(ctx context.Context, experiment *model.Experiment) error
	CreateRun(ctx context.Context, run *model.ExperimentRun) error
	GetByIDWithRuns(ctx context.Context, experimentID uint) (*model.Experiment, error)
}

type GORMExperimentRepository struct {
	db *gorm.DB
}

func NewExperimentRepository(db *gorm.DB) *GORMExperimentRepository {
	return &GORMExperimentRepository{db: db}
}

func (r *GORMExperimentRepository) Create(ctx context.Context, experiment *model.Experiment) error {
	return r.db.WithContext(ctx).Create(experiment).Error
}

func (r *GORMExperimentRepository) Update(ctx context.Context, experiment *model.Experiment) error {
	return r.db.WithContext(ctx).Save(experiment).Error
}

func (r *GORMExperimentRepository) CreateRun(ctx context.Context, run *model.ExperimentRun) error {
	return r.db.WithContext(ctx).Create(run).Error
}

func (r *GORMExperimentRepository) GetByIDWithRuns(ctx context.Context, experimentID uint) (*model.Experiment, error) {
	var experiment model.Experiment
	if err := r.db.WithContext(ctx).
		Preload("Runs", func(db *gorm.DB) *gorm.DB {
			return db.Order("attempt_no ASC, id ASC")
		}).
		Preload("Runs.AISolveRun").
		First(&experiment, experimentID).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExperimentNotFound
		}
		return nil, err
	}
	return &experiment, nil
}
