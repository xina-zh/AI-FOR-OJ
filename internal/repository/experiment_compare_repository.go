package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

var ErrExperimentCompareNotFound = errors.New("experiment compare not found")

type ExperimentCompareRepository interface {
	Create(ctx context.Context, compare *model.ExperimentCompare) error
	Update(ctx context.Context, compare *model.ExperimentCompare) error
	GetByID(ctx context.Context, compareID uint) (*model.ExperimentCompare, error)
}

type GORMExperimentCompareRepository struct {
	db *gorm.DB
}

func NewExperimentCompareRepository(db *gorm.DB) *GORMExperimentCompareRepository {
	return &GORMExperimentCompareRepository{db: db}
}

func (r *GORMExperimentCompareRepository) Create(ctx context.Context, compare *model.ExperimentCompare) error {
	return r.db.WithContext(ctx).Create(compare).Error
}

func (r *GORMExperimentCompareRepository) Update(ctx context.Context, compare *model.ExperimentCompare) error {
	return r.db.WithContext(ctx).Save(compare).Error
}

func (r *GORMExperimentCompareRepository) GetByID(ctx context.Context, compareID uint) (*model.ExperimentCompare, error) {
	var compare model.ExperimentCompare
	if err := r.db.WithContext(ctx).First(&compare, compareID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrExperimentCompareNotFound
		}
		return nil, err
	}
	return &compare, nil
}
