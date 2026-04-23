package service

import (
	"context"
	"testing"
	"time"

	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

func TestSubmissionQueryServiceList(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeSubmissionRepository{
		list: []model.Submission{
			{
				BaseModel:  model.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
				ProblemID:  1,
				Language:   model.LanguageCPP17,
				SourceType: model.SourceTypeHuman,
				Problem:    model.Problem{BaseModel: model.BaseModel{ID: 1}, Title: "Echo"},
				JudgeResult: &model.JudgeResult{
					BaseModel:   model.BaseModel{ID: 11, CreatedAt: now, UpdatedAt: now},
					Verdict:     "AC",
					RuntimeMS:   3,
					PassedCount: 2,
					TotalCount:  2,
					RunStdout:   "ok\n",
					ExitCode:    0,
					ExecStage:   "run",
				},
			},
		},
	}

	service := NewSubmissionQueryService(repo)
	output, err := service.List(context.Background(), SubmissionListInput{
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("list submissions returned error: %v", err)
	}

	if len(output.Items) != 1 {
		t.Fatalf("expected 1 submission, got %d", len(output.Items))
	}

	if output.Total != 1 || output.TotalPages != 1 {
		t.Fatalf("expected total/totalPages 1/1, got %d/%d", output.Total, output.TotalPages)
	}

	if output.Items[0].Verdict != "AC" {
		t.Fatalf("expected verdict AC, got %s", output.Items[0].Verdict)
	}
}

func TestSubmissionQueryServiceGet(t *testing.T) {
	now := time.Now().UTC()
	repo := &fakeSubmissionRepository{
		getResult: &model.Submission{
			BaseModel:  model.BaseModel{ID: 1, CreatedAt: now, UpdatedAt: now},
			ProblemID:  1,
			Language:   model.LanguageCPP17,
			SourceType: model.SourceTypeHuman,
			SourceCode: "#include <bits/stdc++.h>",
			Problem:    model.Problem{BaseModel: model.BaseModel{ID: 1}, Title: "Echo"},
			JudgeResult: &model.JudgeResult{
				BaseModel:      model.BaseModel{ID: 11, CreatedAt: now, UpdatedAt: now},
				Verdict:        "WA",
				RuntimeMS:      5,
				MemoryKB:       0,
				MemoryExceeded: true,
				PassedCount:    1,
				TotalCount:     2,
				RunStdout:      "wrong\n",
				RunStderr:      "",
				ExitCode:       0,
				TimedOut:       false,
				ExecStage:      "run",
				ErrorMessage:   "wrong answer: output mismatch",
			},
			TestCaseResults: []model.SubmissionTestCaseResult{
				{
					CreatedModel: model.CreatedModel{ID: 21, CreatedAt: now},
					SubmissionID: 1,
					TestCaseID:   101,
					CaseIndex:    1,
					Verdict:      "AC",
					RuntimeMS:    2,
					Stdout:       "1\n",
					ExitCode:     0,
				},
				{
					CreatedModel:   model.CreatedModel{ID: 22, CreatedAt: now},
					SubmissionID:   1,
					TestCaseID:     102,
					CaseIndex:      2,
					Verdict:        "WA",
					RuntimeMS:      5,
					Stdout:         "wrong\n",
					ExitCode:       0,
					MemoryExceeded: true,
				},
			},
		},
	}

	service := NewSubmissionQueryService(repo)
	output, err := service.Get(context.Background(), 1)
	if err != nil {
		t.Fatalf("get submission returned error: %v", err)
	}

	if output.SourceCode == "" {
		t.Fatal("expected source code in submission detail")
	}

	if output.JudgeResult == nil || output.JudgeResult.Verdict != "WA" {
		t.Fatal("expected nested judge result in submission detail")
	}

	if output.RunStdout == "" || output.ExecStage != "run" {
		t.Fatalf("expected observability fields in detail output, got %+v", output)
	}

	if !output.MemoryExceeded || output.JudgeResult == nil || !output.JudgeResult.MemoryExceeded {
		t.Fatalf("expected memory exceeded flag in detail output, got %+v", output)
	}

	if len(output.TestCaseResults) != 2 {
		t.Fatalf("expected 2 testcase results, got %d", len(output.TestCaseResults))
	}

	if output.TestCaseResults[1].CaseIndex != 2 || output.TestCaseResults[1].Verdict != "WA" {
		t.Fatalf("expected testcase-level result summary in detail output, got %+v", output.TestCaseResults[1])
	}

	if !output.TestCaseResults[1].MemoryExceeded {
		t.Fatalf("expected testcase memory flag in detail output, got %+v", output.TestCaseResults[1])
	}
}

func TestSubmissionQueryServiceGetReturnsNotFound(t *testing.T) {
	service := NewSubmissionQueryService(&fakeSubmissionRepository{})

	_, err := service.Get(context.Background(), 999)
	if err != repository.ErrSubmissionNotFound {
		t.Fatalf("expected err %v, got %v", repository.ErrSubmissionNotFound, err)
	}
}
