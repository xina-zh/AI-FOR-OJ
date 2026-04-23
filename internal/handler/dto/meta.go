package dto

type ExperimentOptionResponse struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type ExperimentOptionsResponse struct {
	DefaultModel   string                     `json:"default_model"`
	Models         []ExperimentOptionResponse `json:"models"`
	Prompts        []ExperimentOptionResponse `json:"prompts"`
	Agents         []ExperimentOptionResponse `json:"agents"`
	ToolingOptions []ExperimentOptionResponse `json:"tooling_options"`
}
