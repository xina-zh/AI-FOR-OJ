package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

var ErrExperimentRepeatNotFound = errors.New("experiment repeat not found")

type ExperimentRepeatRepository interface {
	Create(ctx context.Context, repeat *model.ExperimentRepeat) error
	Update(ctx context.Context, repeat *model.ExperimentRepeat) error
	GetByID(ctx context.Context, repeatID uint) (*model.ExperimentRepeat, error)
}

type GORMExperimentRepeatRepository struct {
	db *gorm.DB
}

func NewExperimentRepeatRepository(db *gorm.DB) *GORMExperimentRepeatRepository {
	return &GORMExperimentRepeatRepository{db: db}
}

func (r *GORMExperimentRepeatRepository) Create(ctx context.Context, repeat *model.ExperimentRepeat) error {
	return r.db.WithContext(ctx).Create(repeat).Error
}

func (r *GORMExperimentRepeatRepository) Update(ctx context.Context, repeat *model.ExperimentRepeat) error {
	return r.db.WithContext(ctx).Save(repeat).Error
}

func (r *GORMExperimentRepeatRepository) GetByID(ctx context.Context, repeatID uint) (*model.ExperimentRepeat, error) {
	var repeat model.ExperimentRepeat
	if err := r.db.WithContext(ctx).First(&repeat, repeatID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExperimentRepeatNotFound
		}
		return nil, err
	}
	return &repeat, nil
}
