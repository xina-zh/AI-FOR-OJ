package model

import (
	"reflect"
	"testing"
)

func TestSchemaIncludesMemoryExceededFields(t *testing.T) {
	requireModelField(t, JudgeResult{}, "MemoryExceeded")
	requireModelField(t, SubmissionTestCaseResult{}, "MemoryExceeded")
}

func requireModelField(t *testing.T, model any, fieldName string) {
	t.Helper()

	modelType := reflect.TypeOf(model)
	if _, ok := modelType.FieldByName(fieldName); !ok {
		t.Fatalf("expected %s to define %s", modelType.Name(), fieldName)
	}
}
