package tooling

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Config struct {
	EnabledTools    []string       `json:"enabled"`
	MaxCalls        int            `json:"max_calls"`
	PerToolMaxCalls map[string]int `json:"per_tool_max_calls"`
}

func ResolveConfig(raw string) (Config, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return canonicalizeConfig(Config{})
	}

	var cfg Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return Config{}, "", fmt.Errorf("parse tooling config: %w", err)
	}
	return canonicalizeConfig(cfg)
}

func (c Config) Enabled(name string) bool {
	name = strings.TrimSpace(name)
	for _, item := range c.EnabledTools {
		if item == name {
			return true
		}
	}
	return false
}

func (c Config) LimitFor(name string) int {
	if c.PerToolMaxCalls == nil {
		return 0
	}
	return c.PerToolMaxCalls[strings.TrimSpace(name)]
}

func canonicalizeConfig(cfg Config) (Config, string, error) {
	seen := map[string]struct{}{}
	enabled := make([]string, 0, len(cfg.EnabledTools))
	for _, item := range cfg.EnabledTools {
		name := strings.TrimSpace(item)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		enabled = append(enabled, name)
	}
	sort.Strings(enabled)

	if cfg.MaxCalls < 0 {
		cfg.MaxCalls = 0
	}

	limits := map[string]int{}
	for name, limit := range cfg.PerToolMaxCalls {
		name = strings.TrimSpace(name)
		if name == "" || limit <= 0 {
			continue
		}
		limits[name] = limit
	}

	normalized := Config{
		EnabledTools:    enabled,
		MaxCalls:        cfg.MaxCalls,
		PerToolMaxCalls: limits,
	}
	data, err := json.Marshal(normalized)
	if err != nil {
		return Config{}, "", fmt.Errorf("marshal tooling config: %w", err)
	}
	return normalized, string(data), nil
}
