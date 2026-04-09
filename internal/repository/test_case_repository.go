package repository

import (
	"context"

	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

type TestCaseRepository interface {
	Create(ctx context.Context, testCase *model.TestCase) error
	ListByProblemID(ctx context.Context, problemID uint) ([]model.TestCase, error)
}

type GORMTestCaseRepository struct {
	db *gorm.DB
}

func NewTestCaseRepository(db *gorm.DB) *GORMTestCaseRepository {
	return &GORMTestCaseRepository{db: db}
}

func (r *GORMTestCaseRepository) Create(ctx context.Context, testCase *model.TestCase) error {
	return r.db.WithContext(ctx).Create(testCase).Error
}

func (r *GORMTestCaseRepository) ListByProblemID(ctx context.Context, problemID uint) ([]model.TestCase, error) {
	var testCases []model.TestCase
	if err := r.db.WithContext(ctx).
		Where("problem_id = ?", problemID).
		Order("id ASC").
		Find(&testCases).
		Error; err != nil {
		return nil, err
	}

	return testCases, nil
}
