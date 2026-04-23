package tooling

import "strings"

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

func (r *Registry) Register(tool Tool) {
	if r == nil || tool == nil {
		return
	}
	name := strings.TrimSpace(tool.Name())
	if name == "" {
		return
	}
	r.tools[name] = tool
}

func (r *Registry) Lookup(name string) (Tool, bool) {
	if r == nil {
		return nil, false
	}
	tool, ok := r.tools[strings.TrimSpace(name)]
	return tool, ok
}

func (r *Registry) NewRunner(cfg Config) *Runner {
	if r == nil {
		r = NewRegistry()
	}
	return &Runner{
		config:       cfg,
		registry:     r,
		perToolCalls: map[string]int{},
	}
}
