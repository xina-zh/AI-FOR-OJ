package model

import (
	"reflect"
	"slices"
	"testing"
)

func TestSchemaIncludesMemoryExceededFields(t *testing.T) {
	requireModelField(t, JudgeResult{}, "MemoryExceeded")
	requireModelField(t, SubmissionTestCaseResult{}, "MemoryExceeded")
}

func TestSchemaIncludesAISolveAttempt(t *testing.T) {
	models := AllModels()
	if !slices.ContainsFunc(models, func(item any) bool {
		_, ok := item.(*AISolveAttempt)
		return ok
	}) {
		t.Fatal("expected AllModels to include AISolveAttempt")
	}

	requireModelField(t, AISolveRun{}, "AttemptCount")
	requireModelField(t, AISolveRun{}, "FailureType")
	requireModelField(t, AISolveRun{}, "StrategyPath")
}

func requireModelField(t *testing.T, model any, fieldName string) {
	t.Helper()

	modelType := reflect.TypeOf(model)
	if _, ok := modelType.FieldByName(fieldName); !ok {
		t.Fatalf("expected %s to define %s", modelType.Name(), fieldName)
	}
}
