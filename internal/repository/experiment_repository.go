package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

var (
	ErrExperimentNotFound    = errors.New("experiment not found")
	ErrExperimentRunNotFound = errors.New("experiment run not found")
)

type ExperimentRepository interface {
	Create(ctx context.Context, experiment *model.Experiment) error
	Update(ctx context.Context, experiment *model.Experiment) error
	CreateRun(ctx context.Context, run *model.ExperimentRun) error
	List(ctx context.Context, query ExperimentListQuery) ([]model.Experiment, int64, error)
	GetByIDWithRuns(ctx context.Context, experimentID uint) (*model.Experiment, error)
	GetRunTrace(ctx context.Context, runID uint) (*model.ExperimentRun, error)
}

type ExperimentListQuery struct {
	Page     int
	PageSize int
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

func (r *GORMExperimentRepository) List(ctx context.Context, query ExperimentListQuery) ([]model.Experiment, int64, error) {
	var experiments []model.Experiment
	db := r.db.WithContext(ctx).Model(&model.Experiment{})

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.
		Order("created_at DESC, id DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&experiments).
		Error
	if err != nil {
		return nil, 0, err
	}
	return experiments, total, nil
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

func (r *GORMExperimentRepository) GetRunTrace(ctx context.Context, runID uint) (*model.ExperimentRun, error) {
	var run model.ExperimentRun
	if err := r.db.WithContext(ctx).
		Preload("TraceEvents", func(db *gorm.DB) *gorm.DB {
			return db.Order("sequence_no ASC, id ASC")
		}).
		Preload("AISolveRun").
		Preload("Submission").
		Preload("Submission.JudgeResult").
		First(&run, runID).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExperimentRunNotFound
		}
		return nil, err
	}
	return &run, nil
}
