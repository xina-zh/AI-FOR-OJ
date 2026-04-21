package repository

import (
	"context"
	"encoding/json"
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
	Delete(ctx context.Context, problemID uint) error
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

func (r *GORMProblemRepository) Delete(ctx context.Context, problemID uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var problem model.Problem
		if err := tx.First(&problem, problemID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProblemNotFound
			}
			return err
		}

		var testCaseIDs []uint
		if err := tx.Model(&model.TestCase{}).Where("problem_id = ?", problemID).Pluck("id", &testCaseIDs).Error; err != nil {
			return err
		}

		var submissionIDs []uint
		if err := tx.Model(&model.Submission{}).Where("problem_id = ?", problemID).Pluck("id", &submissionIDs).Error; err != nil {
			return err
		}

		var aiSolveRunIDs []uint
		if err := tx.Model(&model.AISolveRun{}).Where("problem_id = ?", problemID).Pluck("id", &aiSolveRunIDs).Error; err != nil {
			return err
		}

		var affectedExperimentIDs []uint
		if err := tx.Model(&model.Experiment{}).Where("problem_id = ?", problemID).Pluck("id", &affectedExperimentIDs).Error; err != nil {
			return err
		}

		var runExperimentIDs []uint
		if err := tx.Model(&model.ExperimentRun{}).Where("problem_id = ?", problemID).Pluck("experiment_id", &runExperimentIDs).Error; err != nil {
			return err
		}
		affectedExperimentIDs = appendUniqueUint(affectedExperimentIDs, runExperimentIDs...)
		affectedExperimentSet := uintSet(affectedExperimentIDs)

		var compares []model.ExperimentCompare
		if err := tx.Find(&compares).Error; err != nil {
			return err
		}
		compareIDs := make([]uint, 0)
		for _, compare := range compares {
			baselineAffected := false
			if compare.BaselineExperimentID != nil {
				_, baselineAffected = affectedExperimentSet[*compare.BaselineExperimentID]
			}
			candidateAffected := false
			if compare.CandidateExperimentID != nil {
				_, candidateAffected = affectedExperimentSet[*compare.CandidateExperimentID]
			}
			if containsUintJSON(compare.ProblemIDs, problemID) || baselineAffected || candidateAffected {
				compareIDs = append(compareIDs, compare.ID)
			}
		}

		var repeats []model.ExperimentRepeat
		if err := tx.Find(&repeats).Error; err != nil {
			return err
		}
		repeatIDs := make([]uint, 0)
		for _, repeat := range repeats {
			if containsUintJSON(repeat.ProblemIDs, problemID) || intersectsUintJSON(repeat.ExperimentIDs, affectedExperimentSet) {
				repeatIDs = append(repeatIDs, repeat.ID)
			}
		}

		var experimentRunIDs []uint
		if len(affectedExperimentIDs) > 0 {
			if err := tx.Model(&model.ExperimentRun{}).Where("experiment_id IN ?", affectedExperimentIDs).Pluck("id", &experimentRunIDs).Error; err != nil {
				return err
			}
		}

		if len(compareIDs) > 0 {
			if err := tx.Where("id IN ?", compareIDs).Delete(&model.ExperimentCompare{}).Error; err != nil {
				return err
			}
		}
		if len(repeatIDs) > 0 {
			if err := tx.Where("id IN ?", repeatIDs).Delete(&model.ExperimentRepeat{}).Error; err != nil {
				return err
			}
		}
		if len(experimentRunIDs) > 0 {
			if err := tx.Where("experiment_run_id IN ?", experimentRunIDs).Delete(&model.TraceEvent{}).Error; err != nil {
				return err
			}
		}
		if len(affectedExperimentIDs) > 0 {
			if err := tx.Where("experiment_id IN ?", affectedExperimentIDs).Delete(&model.ExperimentConfig{}).Error; err != nil {
				return err
			}
			if err := tx.Where("experiment_id IN ?", affectedExperimentIDs).Delete(&model.ExperimentRun{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id IN ?", affectedExperimentIDs).Delete(&model.Experiment{}).Error; err != nil {
				return err
			}
		}
		if len(submissionIDs) > 0 {
			if err := tx.Where("submission_id IN ?", submissionIDs).Delete(&model.JudgeResult{}).Error; err != nil {
				return err
			}
			if err := tx.Where("submission_id IN ?", submissionIDs).Delete(&model.SubmissionTestCaseResult{}).Error; err != nil {
				return err
			}
		}
		if len(testCaseIDs) > 0 {
			if err := tx.Where("test_case_id IN ?", testCaseIDs).Delete(&model.SubmissionTestCaseResult{}).Error; err != nil {
				return err
			}
		}
		if len(aiSolveRunIDs) > 0 {
			if err := tx.Where("id IN ?", aiSolveRunIDs).Delete(&model.AISolveRun{}).Error; err != nil {
				return err
			}
		}
		if len(submissionIDs) > 0 {
			if err := tx.Where("id IN ?", submissionIDs).Delete(&model.Submission{}).Error; err != nil {
				return err
			}
		}
		if len(testCaseIDs) > 0 {
			if err := tx.Where("id IN ?", testCaseIDs).Delete(&model.TestCase{}).Error; err != nil {
				return err
			}
		}

		return tx.Delete(&model.Problem{}, problemID).Error
	})
}

func appendUniqueUint(dst []uint, values ...uint) []uint {
	seen := make(map[uint]struct{}, len(dst)+len(values))
	out := make([]uint, 0, len(dst)+len(values))
	for _, value := range dst {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func containsUintJSON(raw string, target uint) bool {
	var values []uint
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return false
	}
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func intersectsUintJSON(raw string, targets map[uint]struct{}) bool {
	if len(targets) == 0 {
		return false
	}
	var values []uint
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return false
	}
	for _, value := range values {
		if _, ok := targets[value]; ok {
			return true
		}
	}
	return false
}

func uintSet(values []uint) map[uint]struct{} {
	out := make(map[uint]struct{}, len(values))
	for _, value := range values {
		if value != 0 {
			out[value] = struct{}{}
		}
	}
	return out
}
