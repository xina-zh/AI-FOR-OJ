package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"ai-for-oj/internal/model"
)

var ErrSubmissionNotFound = errors.New("submission not found")

type SubmissionRepository interface {
	Create(ctx context.Context, submission *model.Submission) error
	CreateJudgeResult(ctx context.Context, result *model.JudgeResult) error
	CreateTestCaseResults(ctx context.Context, results []model.SubmissionTestCaseResult) error
	List(ctx context.Context, query SubmissionListQuery) ([]model.Submission, int64, error)
	GetByID(ctx context.Context, submissionID uint) (*model.Submission, error)
	AggregateByProblem(ctx context.Context) ([]SubmissionProblemStatsRow, error)
}

type SubmissionListQuery struct {
	Page      int
	PageSize  int
	ProblemID *uint
}

type SubmissionProblemStatsRow struct {
	ProblemID          uint       `json:"problem_id"`
	ProblemTitle       string     `json:"problem_title"`
	TotalSubmissions   int64      `json:"total_submissions"`
	ACCount            int64      `json:"ac_count"`
	WACount            int64      `json:"wa_count"`
	CECount            int64      `json:"ce_count"`
	RECount            int64      `json:"re_count"`
	TLECount           int64      `json:"tle_count"`
	LatestSubmissionAt *time.Time `json:"latest_submission_at,omitempty"`
}

type GORMSubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *GORMSubmissionRepository {
	return &GORMSubmissionRepository{db: db}
}

func (r *GORMSubmissionRepository) Create(ctx context.Context, submission *model.Submission) error {
	return r.db.WithContext(ctx).Create(submission).Error
}

func (r *GORMSubmissionRepository) CreateJudgeResult(ctx context.Context, result *model.JudgeResult) error {
	return r.db.WithContext(ctx).Create(result).Error
}

func (r *GORMSubmissionRepository) CreateTestCaseResults(ctx context.Context, results []model.SubmissionTestCaseResult) error {
	if len(results) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&results).Error
}

func (r *GORMSubmissionRepository) List(ctx context.Context, query SubmissionListQuery) ([]model.Submission, int64, error) {
	var submissions []model.Submission
	db := r.db.WithContext(ctx).Model(&model.Submission{})
	if query.ProblemID != nil {
		db = db.Where("problem_id = ?", *query.ProblemID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := db.
		Preload("Problem").
		Preload("JudgeResult").
		Order("id DESC").
		Offset((query.Page - 1) * query.PageSize).
		Limit(query.PageSize).
		Find(&submissions).
		Error
	if err != nil {
		return nil, 0, err
	}
	return submissions, total, nil
}

func (r *GORMSubmissionRepository) GetByID(ctx context.Context, submissionID uint) (*model.Submission, error) {
	var submission model.Submission
	if err := r.db.WithContext(ctx).
		Preload("Problem").
		Preload("JudgeResult").
		Preload("TestCaseResults", func(db *gorm.DB) *gorm.DB {
			return db.Order("case_index ASC, id ASC")
		}).
		First(&submission, submissionID).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSubmissionNotFound
		}
		return nil, err
	}
	return &submission, nil
}

func (r *GORMSubmissionRepository) AggregateByProblem(ctx context.Context) ([]SubmissionProblemStatsRow, error) {
	var rows []SubmissionProblemStatsRow

	err := r.db.WithContext(ctx).
		Table("submissions").
		Select(`
			submissions.problem_id AS problem_id,
			problems.title AS problem_title,
			COUNT(submissions.id) AS total_submissions,
			SUM(CASE WHEN judge_results.verdict = 'AC' THEN 1 ELSE 0 END) AS ac_count,
			SUM(CASE WHEN judge_results.verdict = 'WA' THEN 1 ELSE 0 END) AS wa_count,
			SUM(CASE WHEN judge_results.verdict = 'CE' THEN 1 ELSE 0 END) AS ce_count,
			SUM(CASE WHEN judge_results.verdict = 'RE' THEN 1 ELSE 0 END) AS re_count,
			SUM(CASE WHEN judge_results.verdict = 'TLE' THEN 1 ELSE 0 END) AS tle_count,
			MAX(submissions.created_at) AS latest_submission_at
		`).
		Joins("JOIN problems ON problems.id = submissions.problem_id").
		Joins("LEFT JOIN judge_results ON judge_results.submission_id = submissions.id").
		Group("submissions.problem_id, problems.title").
		Order("latest_submission_at DESC").
		Scan(&rows).
		Error
	if err != nil {
		return nil, err
	}

	return rows, nil
}
