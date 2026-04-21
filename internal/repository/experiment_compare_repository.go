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
	List(ctx context.Context, query ExperimentCompareListQuery) ([]model.ExperimentCompare, int64, error)
	GetByID(ctx context.Context, compareID uint) (*model.ExperimentCompare, error)
}

type ExperimentCompareListQuery struct {
	Page     int
	PageSize int
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

func (r *GORMExperimentCompareRepository) List(ctx context.Context, query ExperimentCompareListQuery) ([]model.ExperimentCompare, int64, error) {
	var compares []model.ExperimentCompare
	db := r.db.WithContext(ctx).Model(&model.ExperimentCompare{})

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.
		Order("created_at DESC, id DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&compares).
		Error
	if err != nil {
		return nil, 0, err
	}
	return compares, total, nil
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
