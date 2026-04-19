package tooling

type Registry struct {
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: map[string]Tool{}}
}

func (r *Registry) Register(tool Tool) {
	if r == nil || tool == nil || tool.Name() == "" {
		return
	}
	r.tools[tool.Name()] = tool
}

func (r *Registry) NewRunner(cfg Config) *Runner {
	if r == nil {
		r = NewRegistry()
	}
	return &Runner{
		config:       cfg,
		tools:        r.tools,
		perToolCalls: map[string]int{},
	}
}
