package service

import (
	"context"
	"testing"
	"time"

	"ai-for-oj/internal/repository"
)

func TestSubmissionQueryServiceAggregateByProblem(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeSubmissionRepository{
		stats: []repository.SubmissionProblemStatsRow{
			{
				ProblemID:          7,
				ProblemTitle:       "A + B",
				TotalSubmissions:   10,
				ACCount:            4,
				WACount:            2,
				CECount:            1,
				RECount:            2,
				TLECount:           1,
				LatestSubmissionAt: &now,
			},
		},
	}

	service := NewSubmissionQueryService(repo)
	outputs, err := service.AggregateByProblem(context.Background())
	if err != nil {
		t.Fatalf("aggregate by problem returned error: %v", err)
	}

	if len(outputs) != 1 {
		t.Fatalf("expected 1 stats row, got %d", len(outputs))
	}

	if outputs[0].ProblemID != 7 || outputs[0].ACCount != 4 || outputs[0].TLECount != 1 {
		t.Fatalf("unexpected stats row: %+v", outputs[0])
	}
}
