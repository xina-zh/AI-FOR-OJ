package tooling

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Config struct {
	Enabled         []string       `json:"enabled"`
	MaxCalls        int            `json:"max_calls"`
	PerToolMaxCalls map[string]int `json:"per_tool_max_calls"`
}

func ResolveConfig(raw string, registries ...*Registry) (Config, string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return canonicalizeConfig(Config{}, registries...)
	}

	var cfg Config
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return Config{}, "", fmt.Errorf("parse tooling config: %w", err)
	}
	return canonicalizeConfig(cfg, registries...)
}

func (c Config) EnabledTool(name string) bool {
	name = strings.TrimSpace(name)
	for _, item := range c.Enabled {
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

func canonicalizeConfig(cfg Config, registries ...*Registry) (Config, string, error) {
	seen := map[string]struct{}{}
	enabled := make([]string, 0, len(cfg.Enabled))
	for _, item := range cfg.Enabled {
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
		Enabled:         enabled,
		MaxCalls:        cfg.MaxCalls,
		PerToolMaxCalls: limits,
	}
	if len(registries) > 0 && registries[0] != nil {
		for _, name := range normalized.Enabled {
			if _, ok := registries[0].Lookup(name); !ok {
				return Config{}, "", fmt.Errorf("%w: %s", ErrToolNotFound, name)
			}
		}
	}

	data, err := json.Marshal(normalized)
	if err != nil {
		return Config{}, "", fmt.Errorf("marshal tooling config: %w", err)
	}
	return normalized, string(data), nil
}
