package config

import "testing"

func TestLoadAppliesGLMEnvOverrides(t *testing.T) {
	t.Setenv("LLM_GLM_BASE_URL", "https://open.bigmodel.cn/api/paas/v4")
	t.Setenv("LLM_GLM_API_KEY", "glm-test-key")
	t.Setenv("LLM_GLM_MODEL_PREFIX", "glm-")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load config returned error: %v", err)
	}

	if cfg.LLM.GLMBaseURL != "https://open.bigmodel.cn/api/paas/v4" {
		t.Fatalf("unexpected glm base url: %q", cfg.LLM.GLMBaseURL)
	}
	if cfg.LLM.GLMAPIKey != "glm-test-key" {
		t.Fatalf("unexpected glm api key: %q", cfg.LLM.GLMAPIKey)
	}
	if cfg.LLM.GLMModelPrefix != "glm-" {
		t.Fatalf("unexpected glm model prefix: %q", cfg.LLM.GLMModelPrefix)
	}
}

func TestLoadAppliesDeepSeekEnvOverrides(t *testing.T) {
	t.Setenv("LLM_DEEPSEEK_BASE_URL", "https://api.deepseek.com")
	t.Setenv("LLM_DEEPSEEK_API_KEY", "deepseek-test-key")
	t.Setenv("LLM_DEEPSEEK_MODEL_PREFIX", "deepseek-")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.LLM.DeepSeekBaseURL != "https://api.deepseek.com" {
		t.Fatalf("unexpected deepseek base url: %q", cfg.LLM.DeepSeekBaseURL)
	}
	if cfg.LLM.DeepSeekAPIKey != "deepseek-test-key" {
		t.Fatalf("unexpected deepseek api key: %q", cfg.LLM.DeepSeekAPIKey)
	}
	if cfg.LLM.DeepSeekModelPrefix != "deepseek-" {
		t.Fatalf("unexpected deepseek model prefix: %q", cfg.LLM.DeepSeekModelPrefix)
	}
}
