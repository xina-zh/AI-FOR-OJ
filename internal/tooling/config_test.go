package tooling

import (
	"errors"
	"testing"
)

func TestResolveConfigDefaultsToDisabled(t *testing.T) {
	tests := []string{"", "   ", "null"}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			cfg, canonical, err := ResolveConfig(raw)
			if err != nil {
				t.Fatalf("ResolveConfig returned error: %v", err)
			}
			if cfg.EnabledTool("sample_judge") {
				t.Fatal("expected sample_judge to be disabled by default")
			}
			if cfg.MaxCalls != 0 {
				t.Fatalf("expected max calls 0, got %d", cfg.MaxCalls)
			}
			if canonical != `{"enabled":[],"max_calls":0,"per_tool_max_calls":{}}` {
				t.Fatalf("unexpected canonical config: %s", canonical)
			}
		})
	}
}

func TestResolveConfigCanonicalizesStableJSON(t *testing.T) {
	raw := `{"enabled":[" sample_judge ","trace_lookup","sample_judge",""],"max_calls":2,"per_tool_max_calls":{"sample_judge":1,"":9,"trace_lookup":0}}`
	cfg, canonical, err := ResolveConfig(raw)
	if err != nil {
		t.Fatalf("ResolveConfig returned error: %v", err)
	}
	if !cfg.EnabledTool("sample_judge") || !cfg.EnabledTool("trace_lookup") {
		t.Fatalf("expected tools to be enabled, got %+v", cfg)
	}
	if cfg.LimitFor("sample_judge") != 1 {
		t.Fatalf("expected sample_judge limit 1, got %d", cfg.LimitFor("sample_judge"))
	}
	if canonical != `{"enabled":["sample_judge","trace_lookup"],"max_calls":2,"per_tool_max_calls":{"sample_judge":1}}` {
		t.Fatalf("unexpected canonical config: %s", canonical)
	}
}

func TestResolveConfigRejectsInvalidJSON(t *testing.T) {
	_, _, err := ResolveConfig(`{"enabled":`)
	if err == nil {
		t.Fatal("expected invalid JSON to fail")
	}
}

func TestResolveConfigRejectsUnknownEnabledTool(t *testing.T) {
	registry := NewRegistry()
	registry.Register(fakeTool{name: "sample_judge"})

	_, _, err := ResolveConfig(`{"enabled":["missing_tool"],"max_calls":1}`, registry)
	if !errors.Is(err, ErrToolNotFound) {
		t.Fatalf("expected ErrToolNotFound, got %v", err)
	}
}
