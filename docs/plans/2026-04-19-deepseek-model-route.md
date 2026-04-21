# DeepSeek Model Route Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add DeepSeek model routing to the existing OpenAI-compatible LLM client without changing current mock, default OpenAI-compatible, GLM, AI solve, experiment, compare, or repeat behavior.

**Architecture:** Keep the current `openai_compatible` provider as the single real-provider path. Add a DeepSeek-specific endpoint route selected by model prefix, mirroring the existing GLM route design. Requests using `deepseek-chat` or `deepseek-reasoner` go to DeepSeek; all other models continue to use the existing default endpoint or GLM route.

**Tech Stack:** Go, existing `internal/config` YAML/env loading, existing `internal/llm` OpenAI-compatible client, `httptest`, Go testing package.

---

## Context Review

当前项目的模型接入主线已经是 `openai_compatible`：

- `llm.NewClient` 支持 `mock` 和 `openai_compatible`。
- 默认 OpenAI-compatible endpoint 使用 `llm.base_url`、`llm.api_key`、`llm.model`。
- 当前已有一条 GLM 特殊路由：
  - `llm.glm_base_url`
  - `llm.glm_api_key`
  - `llm.glm_model_prefix`
- `OpenAICompatibleClient.endpointForModel` 会根据 request-level `model` 前缀选择不同 endpoint。
- AI solve / experiment / compare / repeat 已经支持请求级 `model` 变量透传。

这意味着 DeepSeek 不需要新增一个全新的 provider，也不应该改 AI solve 主链。最小安全做法是新增一条 DeepSeek endpoint route。

## DeepSeek API Facts

基于官方 DeepSeek API 文档：

- DeepSeek API 兼容 OpenAI API 格式。
- 推荐 `base_url` 是 `https://api.deepseek.com`。
- 也兼容 `https://api.deepseek.com/v1`，但这里的 `v1` 和模型版本无关。
- Chat endpoint 是 `/chat/completions`。
- 当前官方模型名包括：
  - `deepseek-chat`
  - `deepseek-reasoner`

本项目当前 client 会拼接：

```text
endpoint.baseURL + "/chat/completions"
```

所以配置 `deepseek_base_url: https://api.deepseek.com` 时，最终请求是：

```text
https://api.deepseek.com/chat/completions
```

## Scope

本计划只做 DeepSeek 路由接入：

- 新增 DeepSeek 配置字段。
- 新增 DeepSeek 环境变量覆盖。
- 在 `openai_compatible` provider 下增加 DeepSeek route。
- 默认 prefix 使用 `deepseek-`。
- 更新配置示例和 README 使用示例。
- 增加测试保证 DeepSeek 不影响默认 endpoint 和 GLM route。

## Non-Goals

本阶段不做这些内容：

- 不新增 `provider=deepseek`。
- 不改 AI solve / experiment / compare / repeat 的请求结构。
- 不改已有 GLM route 行为。
- 不接入 streaming。
- 不接入 DeepSeek beta prefix completion。
- 不解析 `reasoning_content`。
- 不新增价格表或成本金额统计，只继续使用 token / latency。

## Compatibility Rule

实现时必须满足：

- `provider=mock` 行为不变。
- `provider=openai_compatible` 且未配置 DeepSeek 时行为不变。
- 已配置 GLM 时，`glm-*` 仍走 GLM。
- `deepseek-*` 才走 DeepSeek。
- 其他模型仍走默认 endpoint。
- 缺少 DeepSeek API key 时，只有配置了 DeepSeek route 才报错。

---

## Task 1: Add DeepSeek Config Fields

**Files:**

- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Step 1: Write the failing test**

Append this test to `internal/config/config_test.go`:

```go
func TestLoadAppliesDeepSeekEnvOverrides(t *testing.T) {
	t.Setenv("LLM_DEEPSEEK_BASE_URL", "https://api.deepseek.com")
	t.Setenv("LLM_DEEPSEEK_API_KEY", "deepseek-test-key")
	t.Setenv("LLM_DEEPSEEK_MODEL_PREFIX", "deepseek-")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("load config returned error: %v", err)
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
```

**Step 2: Run test to verify it fails**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/config
```

Expected: FAIL because `DeepSeekBaseURL`, `DeepSeekAPIKey`, and `DeepSeekModelPrefix` do not exist.

**Step 3: Write minimal implementation**

Modify `LLMConfig` in `internal/config/config.go`:

```go
type LLMConfig struct {
	Provider            string        `yaml:"provider"`
	BaseURL             string        `yaml:"base_url"`
	APIKey              string        `yaml:"api_key"`
	Model               string        `yaml:"model"`
	Timeout             time.Duration `yaml:"timeout"`
	MockResponse        string        `yaml:"mock_response"`
	GLMBaseURL          string        `yaml:"glm_base_url"`
	GLMAPIKey           string        `yaml:"glm_api_key"`
	GLMModelPrefix      string        `yaml:"glm_model_prefix"`
	DeepSeekBaseURL     string        `yaml:"deepseek_base_url"`
	DeepSeekAPIKey      string        `yaml:"deepseek_api_key"`
	DeepSeekModelPrefix string        `yaml:"deepseek_model_prefix"`
}
```

Modify `applyEnvOverrides`:

```go
cfg.LLM.DeepSeekBaseURL = getEnvString("LLM_DEEPSEEK_BASE_URL", cfg.LLM.DeepSeekBaseURL)
cfg.LLM.DeepSeekAPIKey = getEnvString("LLM_DEEPSEEK_API_KEY", cfg.LLM.DeepSeekAPIKey)
cfg.LLM.DeepSeekModelPrefix = getEnvString("LLM_DEEPSEEK_MODEL_PREFIX", cfg.LLM.DeepSeekModelPrefix)
```

Do not set DeepSeek defaults in `defaultConfig()`. Keeping defaults empty ensures no new route is enabled unless explicitly configured.

**Step 4: Run test to verify it passes**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/config
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: add deepseek llm config"
```

---

## Task 2: Route DeepSeek Models To DeepSeek Endpoint

**Files:**

- Modify: `internal/llm/client.go`
- Modify: `internal/llm/client_test.go`

**Step 1: Write the failing test**

Append this test to `internal/llm/client_test.go`:

```go
func TestOpenAICompatibleClientRoutesDeepSeekModelsByRequestModel(t *testing.T) {
	var defaultCalls int32
	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer deepseek-key" {
			t.Fatalf("unexpected deepseek auth header: %s", got)
		}
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected deepseek path: %s", r.URL.Path)
		}

		var req openAICompatibleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode deepseek request: %v", err)
		}
		if req.Model != "deepseek-chat" {
			t.Fatalf("unexpected deepseek request model: %s", req.Model)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "deepseek-chat",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
			"usage": map[string]any{
				"prompt_tokens":     7,
				"completion_tokens": 3,
			},
		})
	}))
	defer deepseekServer.Close()

	defaultServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&defaultCalls, 1)
		t.Fatal("default endpoint should not be used for deepseek request model")
	}))
	defer defaultServer.Close()

	client, err := NewClient(config.LLMConfig{
		Provider:            ProviderOpenAICompatible,
		BaseURL:             defaultServer.URL,
		APIKey:              "default-key",
		Model:               "gpt-test",
		Timeout:             5 * time.Second,
		DeepSeekBaseURL:     deepseekServer.URL,
		DeepSeekAPIKey:      "deepseek-key",
		DeepSeekModelPrefix: "deepseek-",
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
		Model:  "deepseek-chat",
	})
	if err != nil {
		t.Fatalf("deepseek-routed generate returned error: %v", err)
	}
	if resp.Model != "deepseek-chat" {
		t.Fatalf("expected deepseek model, got %q", resp.Model)
	}
	if resp.InputTokens != 7 || resp.OutputTokens != 3 {
		t.Fatalf("expected deepseek usage, got %+v", resp)
	}
	if atomic.LoadInt32(&defaultCalls) != 0 {
		t.Fatalf("expected default endpoint to remain unused, got %d calls", defaultCalls)
	}
}
```

Add missing import only if needed:

```go
import "sync/atomic"
```

Current file already imports `sync/atomic`, so no change should be needed.

**Step 2: Run test to verify it fails**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/llm
```

Expected: FAIL because `DeepSeekBaseURL`, `DeepSeekAPIKey`, and `DeepSeekModelPrefix` are not used by `NewClient`.

**Step 3: Write minimal implementation**

Modify `NewClient` in `internal/llm/client.go` after the GLM route block:

```go
if strings.TrimSpace(cfg.DeepSeekBaseURL) != "" || strings.TrimSpace(cfg.DeepSeekAPIKey) != "" || strings.TrimSpace(cfg.DeepSeekModelPrefix) != "" {
	if strings.TrimSpace(cfg.DeepSeekAPIKey) == "" {
		return nil, fmt.Errorf("llm deepseek api key is required when deepseek route is configured")
	}
	deepseekEndpoint, err := newOpenAICompatibleEndpoint(defaultString(cfg.DeepSeekBaseURL, "https://api.deepseek.com"), cfg.DeepSeekAPIKey)
	if err != nil {
		return nil, fmt.Errorf("invalid llm deepseek route: %w", err)
	}
	routes = append(routes, modelEndpointRoute{
		modelPrefix: defaultString(cfg.DeepSeekModelPrefix, "deepseek-"),
		endpoint:    deepseekEndpoint,
	})
}
```

Reasoning:

- Route is opt-in.
- If only `DeepSeekAPIKey` is set, use official default base URL.
- If only `DeepSeekModelPrefix` or base URL is set but key is missing, fail fast.
- Existing default endpoint remains unchanged.

**Step 4: Run test to verify it passes**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/llm
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/llm/client.go internal/llm/client_test.go
git commit -m "feat: route deepseek models"
```

---

## Task 3: Add DeepSeek Default-Model Routing And Missing-Key Guard

**Files:**

- Modify: `internal/llm/client_test.go`

**Step 1: Write the failing tests**

Add default-model routing test:

```go
func TestOpenAICompatibleClientRoutesDeepSeekModelsByDefaultModel(t *testing.T) {
	var defaultCalls int32
	deepseekServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer deepseek-key" {
			t.Fatalf("unexpected deepseek auth header: %s", got)
		}

		var req openAICompatibleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode deepseek request: %v", err)
		}
		if req.Model != "deepseek-reasoner" {
			t.Fatalf("unexpected deepseek request model: %s", req.Model)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "deepseek-reasoner",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"role":    "assistant",
						"content": "```cpp\nint main(){return 0;}\n```",
					},
				},
			},
		})
	}))
	defer deepseekServer.Close()

	defaultServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&defaultCalls, 1)
		t.Fatal("default endpoint should not be used for deepseek default model")
	}))
	defer defaultServer.Close()

	client, err := NewClient(config.LLMConfig{
		Provider:            ProviderOpenAICompatible,
		BaseURL:             defaultServer.URL,
		APIKey:              "default-key",
		Model:               "deepseek-reasoner",
		Timeout:             5 * time.Second,
		DeepSeekBaseURL:     deepseekServer.URL,
		DeepSeekAPIKey:      "deepseek-key",
		DeepSeekModelPrefix: "deepseek-",
	}, nil)
	if err != nil {
		t.Fatalf("new openai compatible client returned error: %v", err)
	}

	resp, err := client.Generate(context.Background(), GenerateRequest{
		Prompt: "solve echo",
	})
	if err != nil {
		t.Fatalf("deepseek default-model generate returned error: %v", err)
	}
	if resp.Model != "deepseek-reasoner" {
		t.Fatalf("expected deepseek model, got %q", resp.Model)
	}
	if atomic.LoadInt32(&defaultCalls) != 0 {
		t.Fatalf("expected default endpoint to remain unused, got %d calls", defaultCalls)
	}
}
```

Add missing-key guard test:

```go
func TestNewClientOpenAICompatibleRequiresDeepSeekAPIKeyWhenRouteConfigured(t *testing.T) {
	_, err := NewClient(config.LLMConfig{
		Provider:            ProviderOpenAICompatible,
		BaseURL:             "https://api.example.test/v1",
		APIKey:              "default-key",
		Model:               "gpt-test",
		DeepSeekModelPrefix: "deepseek-",
	}, nil)
	if err == nil {
		t.Fatal("expected missing deepseek api key error")
	}
	if !strings.Contains(err.Error(), "deepseek api key") {
		t.Fatalf("expected deepseek api key error, got %v", err)
	}
}
```

**Step 2: Run test**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/llm
```

Expected: PASS if Task 2 implementation already covers both behaviors. If it fails, fix only the missing behavior in `internal/llm/client.go`.

**Step 3: Commit**

```bash
git add internal/llm/client.go internal/llm/client_test.go
git commit -m "test: cover deepseek route edge cases"
```

If no implementation change was needed, commit only the test file.

---

## Task 4: Document DeepSeek Configuration

**Files:**

- Modify: `configs/config.example.yaml`
- Modify: `README.md`
- Modify: `docs/dev_progress.md`

**Step 1: Update config example**

Modify `configs/config.example.yaml` under `llm:`:

```yaml
  # Optional: route deepseek-* models to DeepSeek while keeping the same request-level model selection.
  # DeepSeek API is OpenAI-compatible. Official base URL: https://api.deepseek.com
  # deepseek_base_url: https://api.deepseek.com
  # deepseek_api_key: your-deepseek-api-key
  # deepseek_model_prefix: deepseek-
```

Keep existing GLM comments unchanged.

**Step 2: Update README**

Add a short subsection near the model-switching / LLM configuration section:

```markdown
### DeepSeek model route

DeepSeek can be used through the existing `openai_compatible` provider. Configure a separate route for `deepseek-*` models:

```yaml
llm:
  provider: openai_compatible
  base_url: https://api.gptsapi.net/v1
  api_key: your-default-api-key
  model: your-default-model
  deepseek_base_url: https://api.deepseek.com
  deepseek_api_key: your-deepseek-api-key
  deepseek_model_prefix: deepseek-
```

Then choose DeepSeek per request:

```bash
curl --noproxy '*' -sS -X POST http://127.0.0.1:8080/api/v1/ai/solve \
  -H 'Content-Type: application/json' \
  -d '{"problem_id":5,"model":"deepseek-chat"}'
```

`deepseek-chat` and `deepseek-reasoner` are routed to DeepSeek. Other model names still use the default OpenAI-compatible endpoint unless another route matches first.
```
```

Use four-backtick fences if nesting this inside another Markdown code block.

**Step 3: Update dev progress**

Append a short dated note:

```markdown
## 2026-04-19 开发补充

新增 DeepSeek 模型路由计划：保持 `openai_compatible` provider 不变，通过 `deepseek-*` 前缀把 `deepseek-chat` / `deepseek-reasoner` 路由到 DeepSeek endpoint，不影响现有默认模型和 GLM 路由。
```

**Step 4: Run docs-adjacent verification**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/config ./internal/llm
```

Expected: PASS

**Step 5: Commit**

```bash
git add configs/config.example.yaml README.md docs/dev_progress.md
git commit -m "docs: document deepseek model route"
```

---

## Task 5: Full Verification

**Files:**

- No file changes expected.

**Step 1: Run full tests**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./...
```

Expected: PASS

**Step 2: Run server build**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go build -o /tmp/ai-for-oj-server ./cmd/server
```

Expected: PASS

**Step 3: Optional manual local request**

Only run this if local config has a valid DeepSeek API key:

```bash
curl --noproxy '*' -sS -X POST http://127.0.0.1:8080/api/v1/ai/solve \
  -H 'Content-Type: application/json' \
  -d '{"problem_id":5,"model":"deepseek-chat"}'
```

Expected:

- `model` in response is `deepseek-chat`.
- AI solve run is created normally.
- Existing default model request still works.

**Step 4: Commit only if any verification-only fix was needed**

```bash
git status --short
```

Expected: clean or only intended changes already committed.

---

## Acceptance Criteria

- Existing `mock` provider behavior is unchanged.
- Existing default `openai_compatible` endpoint behavior is unchanged when no DeepSeek config is present.
- Existing GLM route tests still pass.
- `model=deepseek-chat` routes to the DeepSeek endpoint when DeepSeek route is configured.
- `model=deepseek-reasoner` routes to the DeepSeek endpoint when DeepSeek route is configured.
- Missing DeepSeek API key fails only when DeepSeek route is configured.
- `go test ./...` passes.
- `go build -o /tmp/ai-for-oj-server ./cmd/server` passes.

## Sources Checked

- DeepSeek official API docs: https://api-docs.deepseek.com/
- DeepSeek official chat completion API docs: https://api-docs.deepseek.com/api/create-chat-completion/
