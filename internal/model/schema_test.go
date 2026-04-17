package model

import (
	"reflect"
	"sync"
	"testing"

	"gorm.io/gorm/schema"
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
	field, ok := typ.FieldByName("Attempts")
	if !ok {
		t.Fatalf("AISolveRun does not expose an Attempts relation")
	}

	if field.Type != reflect.TypeOf([]AISolveAttempt{}) {
		t.Fatalf("AISolveRun.Attempts has type %v, want []AISolveAttempt", field.Type)
	}

	sch, err := schema.Parse(&AISolveRun{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse AISolveRun schema: %v", err)
	}

	rel, ok := sch.Relationships.Relations["Attempts"]
	if !ok {
		t.Fatalf("AISolveRun schema does not register Attempts relation")
	}

	if rel.Type != schema.HasMany {
		t.Fatalf("AISolveRun.Attempts relation type = %v, want %v", rel.Type, schema.HasMany)
	}

	if rel.Field == nil || rel.Field.Name != "Attempts" {
		t.Fatalf("AISolveRun.Attempts relation field = %+v, want Attempts", rel.Field)
	}
}

func TestAISolveAttemptHasCompositeUniqueIndex(t *testing.T) {
	t.Helper()

	typ := reflect.TypeOf(AISolveAttempt{})
	runIDField, ok := typ.FieldByName("AISolveRunID")
	if !ok {
		t.Fatalf("AISolveAttempt does not expose AISolveRunID")
	}
	attemptNoField, ok := typ.FieldByName("AttemptNo")
	if !ok {
		t.Fatalf("AISolveAttempt does not expose AttemptNo")
	}

	const uniqueIndex = "uniqueIndex:idx_ai_solve_attempt_run_no"
	if tag := runIDField.Tag.Get("gorm"); !containsTag(tag, uniqueIndex) {
		t.Fatalf("AISolveAttempt.AISolveRunID gorm tag %q does not include %q", tag, uniqueIndex)
	}
	if tag := attemptNoField.Tag.Get("gorm"); !containsTag(tag, uniqueIndex) {
		t.Fatalf("AISolveAttempt.AttemptNo gorm tag %q does not include %q", tag, uniqueIndex)
	}
}

func containsTag(tag, want string) bool {
	for _, part := range splitTag(tag) {
		if part == want {
			return true
		}
	}
	return false
}

func splitTag(tag string) []string {
	if tag == "" {
		return nil
	}

	parts := make([]string, 0, 8)
	start := 0
	for i := 0; i < len(tag); i++ {
		if tag[i] == ';' {
			if start < i {
				parts = append(parts, tag[start:i])
			}
			start = i + 1
		}
	}
	if start < len(tag) {
		parts = append(parts, tag[start:])
	}
	return parts
}
