package tooling

import "testing"

func TestResolveConfigDefaultsToDisabled(t *testing.T) {
	cfg, canonical, err := ResolveConfig("")
	if err != nil {
		t.Fatalf("ResolveConfig returned error: %v", err)
	}
	if cfg.Enabled("sample_judge") {
		t.Fatal("expected sample_judge to be disabled by default")
	}
	if cfg.MaxCalls != 0 {
		t.Fatalf("expected max calls 0, got %d", cfg.MaxCalls)
	}
	if canonical != `{"enabled":[],"max_calls":0,"per_tool_max_calls":{}}` {
		t.Fatalf("unexpected canonical config: %s", canonical)
	}
}

func TestResolveConfigNormalizesEnabledTools(t *testing.T) {
	raw := `{"enabled":[" sample_judge ","sample_judge",""],"max_calls":2,"per_tool_max_calls":{"sample_judge":1}}`
	cfg, canonical, err := ResolveConfig(raw)
	if err != nil {
		t.Fatalf("ResolveConfig returned error: %v", err)
	}
	if !cfg.Enabled("sample_judge") {
		t.Fatal("expected sample_judge to be enabled")
	}
	if cfg.MaxCalls != 2 {
		t.Fatalf("expected max calls 2, got %d", cfg.MaxCalls)
	}
	if cfg.LimitFor("sample_judge") != 1 {
		t.Fatalf("expected sample_judge limit 1, got %d", cfg.LimitFor("sample_judge"))
	}
	if canonical != `{"enabled":["sample_judge"],"max_calls":2,"per_tool_max_calls":{"sample_judge":1}}` {
		t.Fatalf("unexpected canonical config: %s", canonical)
	}
}

func TestResolveConfigRejectsInvalidJSON(t *testing.T) {
	_, _, err := ResolveConfig(`{"enabled":`)
	if err == nil {
		t.Fatal("expected invalid JSON to fail")
	}
}
