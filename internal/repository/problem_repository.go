package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

var ErrProblemNotFound = errors.New("problem not found")

type ProblemRepository interface {
	Create(ctx context.Context, problem *model.Problem) error
	List(ctx context.Context) ([]model.Problem, error)
	GetByID(ctx context.Context, problemID uint) (*model.Problem, error)
	GetByIDWithTestCases(ctx context.Context, problemID uint) (*model.Problem, error)
}

type GORMProblemRepository struct {
	db *gorm.DB
}

func NewProblemRepository(db *gorm.DB) *GORMProblemRepository {
	return &GORMProblemRepository{db: db}
}

func (r *GORMProblemRepository) Create(ctx context.Context, problem *model.Problem) error {
	return r.db.WithContext(ctx).Create(problem).Error
}

func (r *GORMProblemRepository) List(ctx context.Context) ([]model.Problem, error) {
	var problems []model.Problem
	if err := r.db.WithContext(ctx).Order("id DESC").Find(&problems).Error; err != nil {
		return nil, err
	}
	return problems, nil
}

func (r *GORMProblemRepository) GetByID(ctx context.Context, problemID uint) (*model.Problem, error) {
	var problem model.Problem
	if err := r.db.WithContext(ctx).First(&problem, problemID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}

	return &problem, nil
}

func (r *GORMProblemRepository) GetByIDWithTestCases(ctx context.Context, problemID uint) (*model.Problem, error) {
	var problem model.Problem
	err := r.db.WithContext(ctx).
		Preload("TestCases", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		First(&problem, problemID).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}

	return &problem, nil
}
