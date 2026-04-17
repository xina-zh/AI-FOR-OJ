package model

import (
	"reflect"
	"testing"
)

func TestAllModelsIncludesAISolveAttempt(t *testing.T) {
	t.Helper()

	found := false
	for _, model := range AllModels() {
		typ := reflect.TypeOf(model)
		if typ != nil && typ.Kind() == reflect.Ptr && typ.Elem().Name() == "AISolveAttempt" {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("AllModels() does not include AISolveAttempt")
	}
}

func TestAISolveRunHasAttemptsRelation(t *testing.T) {
	t.Helper()

	typ := reflect.TypeOf(AISolveRun{})
	if _, ok := typ.FieldByName("Attempts"); !ok {
		t.Fatalf("AISolveRun does not expose an Attempts relation")
	}
}
