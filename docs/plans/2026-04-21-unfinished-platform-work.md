# Unfinished Platform Work Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Finish the remaining platform work from the unfinished plans: DeepSeek routing, MLE verdicts, adaptive repair observability, first-class tooling, and the experiment web console.

**Architecture:** Implement backend capabilities first in dependency order, keeping each phase test-driven and independently shippable. DeepSeek and MLE are isolated infrastructure fixes; adaptive repair and tooling extend the AI solve pipeline; the frontend consumes the stabilized APIs and adds operator-facing workflows without moving backend logic into the browser.

**Tech Stack:** Go 1.22, Gin, GORM, Docker sandbox, existing LLM client, React, TypeScript, Vite, TanStack Query, Playwright, Go testing package.

---

## Source Plans Consolidated

This plan supersedes these unfinished plans:

- `docs/plans/2026-04-17-adaptive-repair-agent.md`
- `docs/plans/2026-04-19-deepseek-model-route.md`
- `docs/plans/2026-04-19-toolings-architecture.md`
- `docs/plans/2026-04-20-frontend-experiment-console.md`
- `docs/plans/2026-04-21-sandbox-mle.md`

Do not delete the old plans until this consolidated plan is approved and at least the first execution batch has started. They remain useful as detailed reference material.

## Current Code Snapshot

- `internal/llm/client.go` supports `mock`, default OpenAI-compatible endpoint, and GLM prefix routing.
- `internal/config/config.go` has no DeepSeek fields yet.
- `internal/judge/verdict.go` has `AC`, `WA`, `TLE`, `RE`, `CE`, `UNJUDGEABLE`, but no `MLE`.
- `internal/sandbox/docker.go` still runs containers with `--rm`, so it cannot inspect Docker `State.OOMKilled`.
- `internal/agent/solve.go` has `direct_codegen`, `direct_codegen_repair`, and `analyze_then_codegen`.
- `internal/service/ai_solve.go` still owns the generic repair loop for `direct_codegen_repair`.
- `internal/tooling/doc.go` is only a placeholder package.
- There is no `web/` frontend project.
- Compare/experiment progress persistence is already implemented and is not part of this plan.

## Execution Rules

- Use `@test-driven-development` for every task that changes behavior.
- Use `@systematic-debugging` before changing code in response to any failed verification.
- Use `@senior-frontend` for frontend tasks.
- Run tasks in order unless the human partner explicitly reprioritizes.
- Commit after each task when tests pass.
- Do not mix unrelated dirty worktree files into commits.

Before Task 1, run:

```bash
git branch --show-current
git status --short
go test ./...
```

Expected:

- Branch is not `main` or `master`.
- `go test ./...` passes before starting, or any existing failure is recorded as pre-existing.

---

### Completed Task 1 Summary: Add DeepSeek Config Fields

**Implemented:**

- Added DeepSeek route config fields to `internal/config/config.go`:
  - `DeepSeekBaseURL`
  - `DeepSeekAPIKey`
  - `DeepSeekModelPrefix`
- Added env overrides:
  - `LLM_DEEPSEEK_BASE_URL`
  - `LLM_DEEPSEEK_API_KEY`
  - `LLM_DEEPSEEK_MODEL_PREFIX`
- Added commented DeepSeek example config in `configs/config.example.yaml`.
- Added `TestLoadAppliesDeepSeekEnvOverrides` in `internal/config/config_test.go`.

**Verification:**

- RED: `go test ./internal/config -run TestLoadAppliesDeepSeekEnvOverrides -count=1` failed because DeepSeek fields did not exist.
- GREEN: `go test ./internal/config -count=1` passed.

---

### Task 2: Route DeepSeek Models Through the OpenAI-Compatible Client

**Files:**

- Modify: `internal/llm/client.go`
- Modify: `internal/llm/client_test.go`
- Modify: `README.md`
- Test: `internal/llm/client_test.go`

**Step 1: Write request-model routing test**

Add `TestOpenAICompatibleClientRoutesDeepSeekModelsByRequestModel` to `internal/llm/client_test.go`.

The test should:

- create a DeepSeek `httptest.Server`;
- create a default endpoint server that fails if called;
- configure `config.LLMConfig{Provider: ProviderOpenAICompatible, APIKey: "default-key", BaseURL: defaultServer.URL, DeepSeekBaseURL: deepseekServer.URL, DeepSeekAPIKey: "deepseek-key", DeepSeekModelPrefix: "deepseek-"}`;
- call `Generate` with `Model: "deepseek-chat"`;
- assert request path is `/chat/completions`;
- assert auth header is `Bearer deepseek-key`;
- assert response model and token usage are returned.

**Step 2: Write default-model routing and missing-key tests**

Add tests:

```go
func TestOpenAICompatibleClientRoutesDeepSeekModelsByDefaultModel(t *testing.T) { /* same setup, cfg.Model = "deepseek-reasoner", req.Model empty */ }

func TestNewClientOpenAICompatibleRequiresDeepSeekAPIKeyWhenRouteConfigured(t *testing.T) {
	_, err := NewClient(config.LLMConfig{
		Provider:            ProviderOpenAICompatible,
		APIKey:              "default-key",
		DeepSeekModelPrefix: "deepseek-",
	}, slog.Default())
	if err == nil || !strings.Contains(err.Error(), "deepseek api key") {
		t.Fatalf("expected deepseek api key error, got %v", err)
	}
}
```

**Step 3: Run failing tests**

Run:

```bash
go test ./internal/llm -run 'DeepSeek|RequiresDeepSeek' -count=1
```

Expected: FAIL because DeepSeek route is not wired.

**Step 4: Implement DeepSeek route**

In `NewClient`, after GLM route setup, add:

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

Keep GLM route behavior unchanged.

**Step 5: Document usage**

Add a short README section showing:

```yaml
llm:
  provider: openai_compatible
  base_url: https://api.openai.com/v1
  api_key: your-default-key
  deepseek_base_url: https://api.deepseek.com
  deepseek_api_key: your-deepseek-key
  deepseek_model_prefix: deepseek-
```

Explain that request model `deepseek-chat` and `deepseek-reasoner` route to DeepSeek.

**Step 6: Verify and commit**

Run:

```bash
go test ./internal/llm ./internal/config -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/llm/client.go internal/llm/client_test.go internal/config/config.go internal/config/config_test.go configs/config.example.yaml README.md
git commit -m "feat: route deepseek models"
```

---

### Task 3: Add MLE Verdict and Judge-Level Signal

**Files:**

- Modify: `internal/judge/verdict.go`
- Modify: `internal/sandbox/types.go`
- Modify: `internal/sandbox/mock.go`
- Modify: `internal/judge/types.go`
- Modify: `internal/judge/engine.go`
- Test: `internal/judge/engine_test.go`

**Step 1: Write failing judge test**

Add `TestJudgeMemoryLimitExceeded` to `internal/judge/engine_test.go`:

```go
func TestJudgeMemoryLimitExceeded(t *testing.T) {
	engine := NewEngine(sandbox.NewMockSandbox())

	result, err := engine.Judge(context.Background(), Request{
		Problem: &model.Problem{TimeLimitMS: 1000, MemoryLimitMB: 64},
		TestCases: []model.TestCase{{Input: "1", ExpectedOutput: "1"}},
		Language:   model.LanguageCPP17,
		SourceCode: "MOCK_MLE",
	})
	if err != nil {
		t.Fatalf("judge returned error: %v", err)
	}
	if result.Verdict != VerdictMemoryLimitExceeded || !result.MemoryExceeded {
		t.Fatalf("expected MLE result, got %+v", result)
	}
	if result.MemoryKB < 64*1024 {
		t.Fatalf("expected memory usage at least limit, got %d", result.MemoryKB)
	}
	if len(result.TestCaseResults) != 1 || result.TestCaseResults[0].Verdict != VerdictMemoryLimitExceeded || !result.TestCaseResults[0].MemoryExceeded {
		t.Fatalf("expected testcase MLE, got %+v", result.TestCaseResults)
	}
}
```

**Step 2: Run failing test**

Run:

```bash
go test ./internal/judge -run TestJudgeMemoryLimitExceeded -count=1
```

Expected: FAIL because `VerdictMemoryLimitExceeded`, `MemoryExceeded`, and `MOCK_MLE` do not exist.

**Step 3: Add verdict and result fields**

Add in `internal/judge/verdict.go`:

```go
VerdictMemoryLimitExceeded = "MLE"
```

Add to `sandbox.RunResult`:

```go
MemoryExceeded bool
```

Add to `judge.Result` and `judge.TestCaseResult`:

```go
MemoryExceeded bool
```

**Step 4: Teach mock sandbox to simulate MLE**

In `internal/sandbox/mock.go`, add marker handling:

```go
const mockMemoryLimitExceededMarker = "MOCK_MLE"

if strings.Contains(req.SourceCode, mockMemoryLimitExceededMarker) {
	return RunResult{
		ExitCode:        137,
		RuntimeMS:       1,
		MemoryKB:        req.MemoryLimitMB * 1024,
		MemoryExceeded:  true,
		RuntimeError:    true,
		ErrorMessage:    "memory limit exceeded",
	}, nil
}
```

Use the actual mock file shape; the important behavior is `MemoryExceeded: true`.

**Step 5: Map memory exceeded before runtime error**

In `internal/judge/engine.go`, after TLE handling and before generic RE handling:

```go
if runResult.MemoryExceeded {
	result.Verdict = VerdictMemoryLimitExceeded
	result.MemoryExceeded = true
	result.MemoryKB = maxInt(runResult.MemoryKB, req.Problem.MemoryLimitMB*1024)
	result.ErrorMessage = pickErrorMessage("memory limit exceeded", runResult.ErrorMessage, runResult.Stderr)
	testCaseResult.Verdict = VerdictMemoryLimitExceeded
	testCaseResult.MemoryExceeded = true
	testCaseResult.MemoryKB = result.MemoryKB
	return result, nil
}
```

Adapt names to the existing engine loop. Preserve priority: TLE first, then MLE, then RE.

**Step 6: Verify and commit**

Run:

```bash
go test ./internal/judge ./internal/sandbox -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/judge/verdict.go internal/judge/types.go internal/judge/engine.go internal/judge/engine_test.go internal/sandbox/types.go internal/sandbox/mock.go
git commit -m "feat: add MLE verdict to judge"
```

---

### Task 4: Detect Docker OOMKilled in the Real Sandbox

**Files:**

- Modify: `internal/sandbox/docker.go`
- Test: `internal/sandbox/docker_test.go`

**Step 1: Write unit tests for OOM result mapping**

Add tests around a helper function, not Docker itself:

```go
func TestDockerSandboxRunResultForOOM(t *testing.T) {
	result := runResultForOOM("Killed", 137, 64)
	if !result.MemoryExceeded || !result.RuntimeError || result.MemoryKB < 64*1024 {
		t.Fatalf("expected memory exceeded result, got %+v", result)
	}
}
```

**Step 2: Run failing test**

Run:

```bash
go test ./internal/sandbox -run TestDockerSandboxRunResultForOOM -count=1
```

Expected: FAIL because helper does not exist.

**Step 3: Stop using `--rm` for run containers**

In `DockerSandbox.Run`, remove `--rm` only from run containers. Keep compile containers using `--rm`.

Add cleanup:

```go
defer s.removeContainer(context.Background(), containerName)
```

**Step 4: Inspect OOMKilled**

Add helpers:

```go
func (s *DockerSandbox) containerOOMKilled(ctx context.Context, containerName string) bool {
	cmd := exec.CommandContext(ctx, "docker", "inspect", containerName, "--format", "{{.State.OOMKilled}}")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

func (s *DockerSandbox) removeContainer(ctx context.Context, containerName string) {
	_ = exec.CommandContext(ctx, "docker", "rm", "-f", containerName).Run()
}
```

After `runDockerCommand`, inspect OOM if not timed out:

```go
memoryExceeded := false
if !timedOut {
	memoryExceeded = s.containerOOMKilled(context.Background(), containerName)
}
```

If OOM:

```go
return runResultForOOM(stderr, exitCode, req.MemoryLimitMB), nil
```

**Step 5: Verify and commit**

Run:

```bash
go test ./internal/sandbox -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/sandbox/docker.go internal/sandbox/docker_test.go
git commit -m "feat: detect docker sandbox OOM"
```

---

### Task 5: Persist and Expose MemoryExceeded

**Files:**

- Modify: `internal/model/judge_result.go`
- Modify: `internal/model/submission_test_case_result.go`
- Modify: `internal/service/judge_submission.go`
- Modify: `internal/service/submission_query.go`
- Modify: `internal/handler/dto/judge_submission.go`
- Modify: `internal/handler/dto/submission_query.go`
- Test: `internal/service/judge_submission_test.go`
- Test: `internal/model/schema_test.go`

**Step 1: Write failing service persistence test**

In `internal/service/judge_submission_test.go`, add a fake judge result with:

```go
Verdict:        judge.VerdictMemoryLimitExceeded,
MemoryExceeded: true,
MemoryKB:       64 * 1024,
```

Assert persisted `JudgeResult.MemoryExceeded` and testcase `MemoryExceeded` are true, and output DTO fields expose `memory_exceeded`.

**Step 2: Run failing tests**

Run:

```bash
go test ./internal/service -run TestJudgeSubmission.*Memory -count=1
```

Expected: FAIL because persisted/service output fields are missing.

**Step 3: Add model fields**

Add to `model.JudgeResult` and `model.SubmissionTestCaseResult`:

```go
MemoryExceeded bool `gorm:"column:memory_exceeded;not null;default:false" json:"memory_exceeded"`
```

**Step 4: Map fields through service and DTOs**

Add `MemoryExceeded bool` to service outputs and DTO responses:

```go
MemoryExceeded bool `json:"memory_exceeded"`
```

Map the field when saving judge result rows, testcase rows, and when returning submission query results.

**Step 5: Verify and commit**

Run:

```bash
go test ./internal/service ./internal/model ./internal/handler -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/model/judge_result.go internal/model/submission_test_case_result.go internal/service/judge_submission.go internal/service/submission_query.go internal/service/judge_submission_test.go internal/model/schema_test.go internal/handler/dto/judge_submission.go internal/handler/dto/submission_query.go
git commit -m "feat: persist memory exceeded judge results"
```

---

### Task 6: Add Adaptive Attempt Model and Repository

**Files:**

- Create: `internal/model/ai_solve_attempt.go`
- Create: `internal/repository/ai_solve_attempt_repository.go`
- Modify: `internal/model/ai_solve_run.go`
- Modify: `internal/model/schema.go`
- Modify: `internal/bootstrap/app.go`
- Test: `internal/model/schema_test.go`
- Test: `internal/repository/ai_solve_attempt_repository_test.go`

**Step 1: Write failing schema test**

Assert `model.AllModels()` includes `&model.AISolveAttempt{}` and `AISolveRun` has attempt summary fields:

```go
AttemptCount int
FailureType  string
StrategyPath string
```

**Step 2: Run failing model test**

Run:

```bash
go test ./internal/model -run TestSchemaIncludesAISolveAttempt -count=1
```

Expected: FAIL.

**Step 3: Add model**

Create `internal/model/ai_solve_attempt.go`:

```go
package model

type AISolveAttempt struct {
	CreatedModel
	AISolveRunID   uint       `gorm:"column:ai_solve_run_id;not null;index;uniqueIndex:idx_ai_solve_attempt" json:"ai_solve_run_id"`
	AttemptNo      int        `gorm:"column:attempt_no;not null;uniqueIndex:idx_ai_solve_attempt" json:"attempt_no"`
	Stage          string     `gorm:"column:stage;type:varchar(64);not null" json:"stage"`
	Model          string     `gorm:"column:model;type:varchar(128);not null;default:''" json:"model"`
	PromptPreview  string     `gorm:"column:prompt_preview;type:longtext" json:"prompt_preview"`
	RawResponse    string     `gorm:"column:raw_response;type:longtext" json:"raw_response"`
	ExtractedCode  string     `gorm:"column:extracted_code;type:longtext" json:"extracted_code"`
	Verdict        string     `gorm:"column:verdict;type:varchar(32);not null;default:''" json:"verdict"`
	FailureType    string     `gorm:"column:failure_type;type:varchar(64);not null;default:''" json:"failure_type"`
	RepairReason   string     `gorm:"column:repair_reason;type:longtext" json:"repair_reason"`
	TokenInput     int64      `gorm:"column:token_input;not null;default:0" json:"token_input"`
	TokenOutput    int64      `gorm:"column:token_output;not null;default:0" json:"token_output"`
	LLMLatencyMS   int        `gorm:"column:llm_latency_ms;not null;default:0" json:"llm_latency_ms"`
	TotalLatencyMS int        `gorm:"column:total_latency_ms;not null;default:0" json:"total_latency_ms"`
	AISolveRun     AISolveRun `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
}

func (AISolveAttempt) TableName() string {
	return "ai_solve_attempts"
}
```

Add to `AISolveRun`:

```go
AttemptCount int              `gorm:"column:attempt_count;not null;default:0" json:"attempt_count"`
FailureType  string           `gorm:"column:failure_type;type:varchar(64);not null;default:''" json:"failure_type"`
StrategyPath string           `gorm:"column:strategy_path;type:longtext" json:"strategy_path"`
Attempts     []AISolveAttempt `gorm:"foreignKey:AISolveRunID" json:"attempts,omitempty"`
```

**Step 4: Add repository**

Create `internal/repository/ai_solve_attempt_repository.go`:

```go
type AISolveAttemptRepository interface {
	Create(ctx context.Context, attempt *model.AISolveAttempt) error
	ListByRunID(ctx context.Context, runID uint) ([]model.AISolveAttempt, error)
}
```

Implement GORM `Create` and `ListByRunID` ordered by `attempt_no ASC, id ASC`.

Wire repository in `internal/bootstrap/app.go`.

**Step 5: Verify and commit**

Run:

```bash
go test ./internal/model ./internal/repository ./internal/bootstrap -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/model/ai_solve_attempt.go internal/model/ai_solve_run.go internal/model/schema.go internal/model/schema_test.go internal/repository/ai_solve_attempt_repository.go internal/repository/ai_solve_attempt_repository_test.go internal/bootstrap/app.go
git commit -m "feat: add ai solve attempt persistence"
```

---

### Task 7: Add Adaptive Agent Registration, Classifier, Planner, and Repair Prompts

**Files:**

- Create: `internal/agent/failure_classifier.go`
- Create: `internal/agent/repair_planner.go`
- Create: `internal/prompt/repair_wa.go`
- Create: `internal/prompt/repair_re.go`
- Create: `internal/prompt/repair_tle.go`
- Modify: `internal/agent/solve.go`
- Modify: `internal/prompt/solve.go`
- Test: `internal/agent/adaptive_repair_test.go`
- Test: `internal/prompt/solve_test.go`

**Step 1: Write failing agent registration test**

Assert:

```go
resolved, err := ResolveSolveAgentName("adaptive_repair_v1")
if err != nil || resolved != AdaptiveRepairV1AgentName { t.Fatalf(...) }
if !slices.Contains(ListSolveAgents(), AdaptiveRepairV1AgentName) { t.Fatalf(...) }
```

**Step 2: Write failing classifier and planner tests**

Add cases:

- verdict `WA` -> failure type `wrong_answer` and next stage `wa_analysis_repair`;
- verdict `RE` -> `runtime_error` and `re_safety_repair`;
- verdict `TLE` -> `time_limit` and `tle_complexity_rewrite`;
- exhausted attempts -> no next repair.

**Step 3: Run failing tests**

Run:

```bash
go test ./internal/agent ./internal/prompt -run 'Adaptive|Repair|Failure|Planner' -count=1
```

Expected: FAIL.

**Step 4: Implement constants and helpers**

Add:

```go
const AdaptiveRepairV1AgentName = "adaptive_repair_v1"
```

Add classifier and planner types with explicit stage constants:

```go
const (
	StageInitialCodegen       = "initial_codegen"
	StageWAAnalysisRepair     = "wa_analysis_repair"
	StageRESafetyRepair       = "re_safety_repair"
	StageTLEComplexityRewrite = "tle_complexity_rewrite"
	StageFallbackRewrite      = "fallback_rewrite"
)
```

**Step 5: Implement repair prompts**

Each repair prompt builder must include:

- original problem statement;
- previous source code;
- judge verdict and feedback;
- C++17-only instruction;
- return-only-code-block instruction.

**Step 6: Verify and commit**

Run:

```bash
go test ./internal/agent ./internal/prompt -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/agent/solve.go internal/agent/failure_classifier.go internal/agent/repair_planner.go internal/agent/adaptive_repair_test.go internal/prompt/repair_wa.go internal/prompt/repair_re.go internal/prompt/repair_tle.go internal/prompt/solve.go internal/prompt/solve_test.go
git commit -m "feat: add adaptive repair planning"
```

---

### Task 8: Move Adaptive Solve Loop Into a Coordinator and Persist Attempts

**Files:**

- Create: `internal/agent/coordinator.go`
- Create: `internal/agent/executor.go`
- Modify: `internal/service/ai_solve.go`
- Modify: `internal/service/ai_solve_test.go`
- Modify: `internal/repository/ai_solve_run_repository.go`
- Modify: `internal/handler/dto/ai_solve.go`
- Modify: `internal/handler/ai.go`
- Modify: `internal/service/experiment.go`
- Modify: `internal/handler/dto/experiment.go`
- Test: `internal/agent/adaptive_repair_test.go`
- Test: `internal/service/ai_solve_test.go`
- Test: `internal/service/experiment_test.go`

**Step 1: Write failing adaptive service tests**

Add tests proving:

- `adaptive_repair_v1` creates multiple persisted attempts on `WA`;
- `RE` and `TLE` select different repair stages;
- terminal output contains `attempt_count`, `failure_type`, and `strategy_path`;
- existing `direct_codegen` still stops after one failed attempt;
- existing `direct_codegen_repair` behavior remains compatible.

**Step 2: Run failing tests**

Run:

```bash
go test ./internal/service -run 'AdaptiveRepair|DirectCodegen' -count=1
```

Expected: FAIL because attempt persistence and coordinator are missing.

**Step 3: Create coordinator contract**

In `internal/agent/coordinator.go`, define:

```go
type JudgeSubmitter interface {
	Submit(ctx context.Context, sourceCode string) (*JudgeResult, error)
}

type AttemptRecorder interface {
	RecordAttempt(ctx context.Context, attempt AttemptRecord) error
}
```

Keep the exact adapter shapes in service to avoid importing service types into `agent`.

**Step 4: Refactor `AISolveService`**

For `adaptive_repair_v1`, call the coordinator. For legacy agents, keep current behavior unless a helper can be extracted without changing output.

Persist every attempt through `AISolveAttemptRepository`. Update `AISolveRun.AttemptCount`, `FailureType`, and `StrategyPath` before terminal persistence.

**Step 5: Expose attempts in API outputs**

Add response DTO:

```go
type AISolveAttemptResponse struct {
	ID             uint   `json:"id"`
	AttemptNo      int    `json:"attempt_no"`
	Stage          string `json:"stage"`
	Model          string `json:"model"`
	Verdict        string `json:"verdict"`
	FailureType    string `json:"failure_type"`
	RepairReason   string `json:"repair_reason"`
	TokenInput     int64  `json:"token_input"`
	TokenOutput    int64  `json:"token_output"`
	LLMLatencyMS   int    `json:"llm_latency_ms"`
	TotalLatencyMS int    `json:"total_latency_ms"`
}
```

Add `attempt_count`, `failure_type`, `strategy_path`, and `attempts` to AI solve and experiment outputs.

**Step 6: Verify and commit**

Run:

```bash
go test ./internal/agent ./internal/service ./internal/handler ./internal/repository -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/agent/coordinator.go internal/agent/executor.go internal/agent/adaptive_repair_test.go internal/service/ai_solve.go internal/service/ai_solve_test.go internal/service/experiment.go internal/service/experiment_test.go internal/repository/ai_solve_run_repository.go internal/handler/ai.go internal/handler/dto/ai_solve.go internal/handler/dto/experiment.go
git commit -m "feat: implement adaptive repair coordinator"
```

---

### Task 9: Add Tooling Config Parser, Registry, and Runner

**Files:**

- Create: `internal/tooling/config.go`
- Create: `internal/tooling/registry.go`
- Create: `internal/tooling/runner.go`
- Test: `internal/tooling/config_test.go`
- Test: `internal/tooling/runner_test.go`

**Step 1: Write failing config parser tests**

Cover:

- empty string, missing, and `null` -> disabled config;
- malformed JSON returns error;
- unknown `enabled` tool returns validation error when registry is provided;
- canonical JSON output is stable.

**Step 2: Write failing runner limit tests**

Cover:

- total `max_calls` limit;
- per-tool limit;
- disabled config rejects calls.

**Step 3: Run failing tests**

Run:

```bash
go test ./internal/tooling -count=1
```

Expected: FAIL because tooling implementation does not exist.

**Step 4: Implement config and interfaces**

Add:

```go
type Config struct {
	Enabled         []string       `json:"enabled"`
	MaxCalls        int            `json:"max_calls"`
	PerToolMaxCalls map[string]int `json:"per_tool_max_calls"`
}

type Tool interface {
	Name() string
	Run(ctx context.Context, input CallInput) (CallOutput, error)
}

type Registry struct { /* register and lookup by name */ }

type Runner interface {
	Call(ctx context.Context, name string, input CallInput) (CallOutput, error)
	CallCount() int
}
```

**Step 5: Verify and commit**

Run:

```bash
go test ./internal/tooling -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/tooling/config.go internal/tooling/config_test.go internal/tooling/registry.go internal/tooling/runner.go internal/tooling/runner_test.go
git commit -m "feat: add tooling runner"
```

---

### Task 10: Implement Sample Judge Tool and Shared Code Extraction

**Files:**

- Create: `internal/tooling/sample_judge.go`
- Modify: `internal/agent/solve.go`
- Modify: `internal/service/ai_solve.go`
- Test: `internal/tooling/sample_judge_test.go`
- Test: `internal/agent/solve_test.go`
- Test: `internal/service/ai_solve_test.go`

**Step 1: Write failing sample judge tests**

Use a fake sandbox/judge adapter to assert `sample_judge`:

- compiles and runs only sample/public testcase input;
- does not persist a final submission;
- returns verdict, stdout/stderr, runtime, and error text.

**Step 2: Move C++ extraction into agent package**

Write a failing test for `agent.ExtractCPPCode`, then move current `extractCPPCode` logic from `internal/service/ai_solve.go` to `internal/agent`.

**Step 3: Implement sample judge tool**

The tool input must include:

```go
type SampleJudgeInput struct {
	Problem    *model.Problem
	SourceCode string
}
```

Use existing judge/sandbox abstractions. The tool must not create `Submission` rows.

**Step 4: Verify and commit**

Run:

```bash
go test ./internal/tooling ./internal/agent ./internal/service -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/tooling/sample_judge.go internal/tooling/sample_judge_test.go internal/agent/solve.go internal/agent/solve_test.go internal/service/ai_solve.go internal/service/ai_solve_test.go
git commit -m "feat: add sample judge tool"
```

---

### Task 11: Pass Tooling Through AI Solve, Experiment, Compare, and Repeat

**Files:**

- Modify: `internal/model/ai_solve_run.go`
- Modify: `internal/model/experiment.go`
- Modify: `internal/model/experiment_compare.go`
- Modify: `internal/model/experiment_repeat.go`
- Modify: `internal/service/ai_solve.go`
- Modify: `internal/service/experiment.go`
- Modify: `internal/service/experiment_compare.go`
- Modify: `internal/service/experiment_repeat.go`
- Modify: `internal/handler/experiment.go`
- Modify: `internal/handler/dto/ai_solve.go`
- Modify: `internal/handler/dto/experiment.go`
- Modify: `internal/handler/dto/experiment_compare.go`
- Modify: `internal/handler/dto/experiment_repeat.go`
- Test: `internal/service/ai_solve_test.go`
- Test: `internal/service/experiment_test.go`
- Test: `internal/service/experiment_compare_test.go`
- Test: `internal/service/experiment_repeat_test.go`

**Step 1: Write failing pass-through tests**

Assert:

- AI solve accepts `tooling_config` and returns canonical config plus `tool_call_count`.
- Experiment passes `tooling_config` into every solve.
- Compare accepts baseline/candidate tooling configs and detects `compare_dimension=tooling` when only tooling differs.
- Repeat passes tooling config to each round.
- Requests without `tooling_config` behave as disabled.

**Step 2: Run failing tests**

Run:

```bash
go test ./internal/service -run 'Tooling|Experiment|Compare|Repeat' -count=1
```

Expected: FAIL.

**Step 3: Add persistence fields**

Add canonical string fields:

```go
ToolingConfig string `gorm:"column:tooling_config;type:longtext;not null;default:'{}'" json:"tooling_config"`
ToolCallCount int    `gorm:"column:tool_call_count;not null;default:0" json:"tool_call_count"`
```

For compare, use:

```go
BaselineToolingConfig  string
CandidateToolingConfig string
```

**Step 4: Map service and DTO fields**

Use `json.RawMessage` in request DTOs where useful, but store and return canonical JSON strings in service/model outputs.

**Step 5: Verify and commit**

Run:

```bash
go test ./internal/service ./internal/handler ./internal/model -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/model/ai_solve_run.go internal/model/experiment.go internal/model/experiment_compare.go internal/model/experiment_repeat.go internal/service/ai_solve.go internal/service/experiment.go internal/service/experiment_compare.go internal/service/experiment_repeat.go internal/handler/experiment.go internal/handler/dto/ai_solve.go internal/handler/dto/experiment.go internal/handler/dto/experiment_compare.go internal/handler/dto/experiment_repeat.go internal/service/*_test.go
git commit -m "feat: pass tooling through experiments"
```

---

### Task 12: Add Tooling Codegen Agent and Bootstrap Registration

**Files:**

- Modify: `internal/agent/solve.go`
- Modify: `internal/bootstrap/app.go`
- Modify: `internal/service/ai_solve.go`
- Test: `internal/agent/solve_test.go`
- Test: `internal/service/ai_solve_test.go`

**Step 1: Write failing tooling agent test**

Assert `tooling_codegen_v1`:

- is listed in `ListSolveAgents`;
- can call `sample_judge` when enabled;
- continues with generated code if the tool call fails;
- increments `tool_call_count`.

**Step 2: Run failing tests**

Run:

```bash
go test ./internal/agent ./internal/service -run 'ToolingCodegen|Tooling' -count=1
```

Expected: FAIL.

**Step 3: Implement agent**

Add:

```go
const ToolingCodegenV1AgentName = "tooling_codegen_v1"
```

Extend `agent.SolveInput` with a `tooling.Runner` dependency. `tooling_codegen_v1` should:

1. generate code once;
2. extract C++ code;
3. call `sample_judge` if enabled;
4. optionally include sample feedback in one repair prompt;
5. return final code response and tool call count.

Do not let tool failure fail the whole solve in phase 1.

**Step 4: Register sample judge in bootstrap**

Create a registry in `internal/bootstrap/app.go` and pass a runner factory to `AISolveService`.

**Step 5: Verify and commit**

Run:

```bash
go test ./internal/agent ./internal/tooling ./internal/service ./internal/bootstrap -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/agent/solve.go internal/agent/solve_test.go internal/bootstrap/app.go internal/service/ai_solve.go internal/service/ai_solve_test.go
git commit -m "feat: add tooling codegen agent"
```

---

### Task 13: Add Backend APIs Needed by the Frontend Console

**Files:**

- Create: `internal/handler/meta.go`
- Create: `internal/handler/dto/meta.go`
- Modify: `internal/runtime/router.go`
- Modify: `internal/runtime/router_test.go`
- Modify: `internal/repository/experiment_repository.go`
- Modify: `internal/repository/experiment_compare_repository.go`
- Modify: `internal/repository/experiment_repeat_repository.go`
- Modify: `internal/service/experiment.go`
- Modify: `internal/service/experiment_compare.go`
- Modify: `internal/service/experiment_repeat.go`
- Modify: `internal/handler/experiment.go`
- Modify: `internal/handler/dto/experiment.go`
- Modify: `internal/handler/dto/experiment_compare.go`
- Modify: `internal/handler/dto/experiment_repeat.go`

**Step 1: Write failing metadata route test**

Add router test for:

```text
GET /api/v1/meta/experiment-options
```

Expected response includes models, prompts, agents, and tooling options.

**Step 2: Write failing history route tests**

Add route tests for:

```text
GET /api/v1/experiments?page=1&page_size=20
GET /api/v1/experiments/compare?page=1&page_size=20
GET /api/v1/experiments/repeat?page=1&page_size=20
```

**Step 3: Write failing trace route test**

Add route test for:

```text
GET /api/v1/experiment-runs/:id/trace
```

Return a timeline built from `TraceEvent` plus linked AI solve/submission data if present.

**Step 4: Implement routes and list methods**

Use existing repository/service patterns. Keep pagination shape:

```json
{
  "items": [],
  "page": 1,
  "page_size": 20,
  "total": 0,
  "total_pages": 0
}
```

**Step 5: Verify and commit**

Run:

```bash
go test ./internal/runtime ./internal/handler ./internal/service ./internal/repository -count=1
```

Expected: PASS.

Commit:

```bash
git add internal/handler/meta.go internal/handler/dto/meta.go internal/runtime/router.go internal/runtime/router_test.go internal/repository/experiment_repository.go internal/repository/experiment_compare_repository.go internal/repository/experiment_repeat_repository.go internal/service/experiment.go internal/service/experiment_compare.go internal/service/experiment_repeat.go internal/handler/experiment.go internal/handler/dto/experiment.go internal/handler/dto/experiment_compare.go internal/handler/dto/experiment_repeat.go
git commit -m "feat: add experiment console backend APIs"
```

---

### Task 14: Scaffold the React Frontend

**Files:**

- Create: `web/package.json`
- Create: `web/vite.config.ts`
- Create: `web/tsconfig.json`
- Create: `web/index.html`
- Create: `web/src/main.tsx`
- Create: `web/src/App.tsx`
- Create: `web/src/styles/base.css`
- Modify: `README.md`

**Step 1: Create Vite app structure**

Use React + TypeScript. Required dependencies:

- `@vitejs/plugin-react`
- `vite`
- `typescript`
- `react`
- `react-dom`
- `react-router-dom`
- `@tanstack/react-query`
- `vitest`
- `@testing-library/react`
- `@testing-library/jest-dom`
- `@playwright/test`

**Step 2: Add scripts**

`web/package.json` must include:

```json
{
  "scripts": {
    "dev": "vite --host 0.0.0.0",
    "build": "tsc -b && vite build",
    "test": "vitest run",
    "e2e": "playwright test"
  }
}
```

**Step 3: Build minimal shell**

Add routes for:

- `/`
- `/problems`
- `/solve`
- `/experiments`
- `/experiments/:id`
- `/compare`
- `/compare/:id`
- `/repeat`
- `/repeat/:id`
- `/submissions`
- `/trace/experiment-runs/:id`

**Step 4: Verify and commit**

Run:

```bash
cd web && npm install
cd web && npm run build
```

Expected: PASS.

Commit:

```bash
git add web README.md
git commit -m "feat: scaffold experiment console"
```

---

### Task 15: Add Frontend API Layer and Shared UI

**Files:**

- Create: `web/src/api/client.ts`
- Create: `web/src/api/types.ts`
- Create: `web/src/api/experimentApi.ts`
- Create: `web/src/components/Layout.tsx`
- Create: `web/src/components/DataTable.tsx`
- Create: `web/src/components/StatusBadge.tsx`
- Create: `web/src/components/MetricStrip.tsx`
- Create: `web/src/features/options/ExperimentVariableForm.tsx`
- Test: `web/src/api/experimentApi.test.ts`
- Test: `web/src/components/Layout.test.tsx`

**Step 1: Define API types**

Include types for:

- problem;
- submission;
- AI solve run;
- experiment;
- compare;
- repeat;
- trace event;
- options metadata;
- tooling config.

**Step 2: Implement typed fetch client**

The client should use `VITE_API_BASE_URL`, defaulting to same-origin `/api/v1`.

**Step 3: Build shared controls**

Use domain-appropriate dense controls:

- selects for `model`, `prompt_name`, `agent_name`;
- textarea/editor for JSON `tooling_config`;
- status badges for verdict/run status;
- metric strips for token/latency counts.

**Step 4: Verify and commit**

Run:

```bash
cd web && npm run test
cd web && npm run build
```

Expected: PASS.

Commit:

```bash
git add web/src/api web/src/components web/src/features/options
git commit -m "feat: add console API layer"
```

---

### Task 16: Build Frontend Workflows

**Files:**

- Create: `web/src/features/dashboard/*`
- Create: `web/src/features/problems/*`
- Create: `web/src/features/solve/*`
- Create: `web/src/features/experiments/*`
- Create: `web/src/features/compare/*`
- Create: `web/src/features/repeat/*`
- Create: `web/src/features/submissions/*`
- Create: `web/src/features/trace/*`
- Create: `web/src/features/analytics/*`
- Modify: `web/src/App.tsx`
- Test: `web/src/features/**/*.test.tsx`

**Step 1: Build Problems workflow**

List problems, show detail, and create problem/testcases if backend endpoints exist.

**Step 2: Build Single Solve workflow**

Form fields:

- `problem_id`;
- `model`;
- `prompt_name`;
- `agent_name`;
- `tooling_config`.

Show raw response, extracted code, verdict, token usage, latency, attempt history, and linked submission.

**Step 3: Build Experiment/Compare/Repeat workflows**

Each workflow must support:

- run form;
- history table;
- detail page;
- verdict distribution;
- cost summary;
- per-problem rows.

Compare detail must show baseline/candidate side by side and highlight improved/regressed/changed non-AC problems.

**Step 4: Build Trace and Analytics views**

Trace view loads `GET /api/v1/experiment-runs/:id/trace`.

Analytics view aggregates token and latency summaries from solve/experiment/compare/repeat responses.

**Step 5: Verify and commit**

Run:

```bash
cd web && npm run test
cd web && npm run build
```

Expected: PASS.

Commit:

```bash
git add web/src/features web/src/App.tsx
git commit -m "feat: build experiment console workflows"
```

---

### Task 17: Add E2E Smoke Tests and Documentation

**Files:**

- Create: `web/e2e/console.spec.ts`
- Modify: `README.md`
- Modify: `docs/ai_solve_options.md`
- Modify: `configs/config.example.yaml`

**Step 1: Add Playwright smoke tests**

Cover:

- app loads dashboard;
- options metadata renders;
- problem list renders from mocked or dev API;
- solve form can submit with mocked API;
- compare detail renders baseline/candidate summaries.

**Step 2: Document workflows**

Document:

- DeepSeek route config;
- MLE verdict meaning;
- `adaptive_repair_v1`;
- `tooling_codegen_v1`;
- `tooling_config` examples;
- frontend dev commands.

**Step 3: Run full verification**

Run:

```bash
go test ./...
cd web && npm run test
cd web && npm run build
cd web && npm run e2e
```

Expected: PASS.

**Step 4: Commit**

```bash
git add web/e2e README.md docs/ai_solve_options.md configs/config.example.yaml
git commit -m "docs: document experiment platform workflows"
```

---

### Task 18: Manual Runtime Smoke Tests

**Files:**

- No code changes expected.

**Step 1: Start the stack**

Run:

```bash
docker compose up -d --build
```

Expected:

- Backend container is healthy.
- Database is reachable.
- Frontend dev server can be started with `cd web && npm run dev`.

**Step 2: Smoke test MLE**

Create a low-memory problem and submit memory-heavy C++.

Expected response:

```json
{
  "verdict": "MLE",
  "memory_exceeded": true
}
```

**Step 3: Smoke test model routes**

Run one solve with:

- default model;
- `glm-*`;
- `deepseek-chat`.

Expected:

- default endpoint used for default model;
- GLM route used for GLM;
- DeepSeek route used for DeepSeek.

**Step 4: Smoke test adaptive/tooling/frontend**

Run:

- `adaptive_repair_v1` on a problem requiring repair;
- `tooling_codegen_v1` with `{"enabled":["sample_judge"],"max_calls":1}`;
- one compare from the web UI.

Expected:

- attempt history appears;
- tool call count appears;
- compare detail shows baseline/candidate data and progress child IDs.

**Step 5: Final cleanup and old-plan decision**

After this task passes, ask the human partner whether to delete the five superseded plan files.

---

## Cross-Phase Acceptance Criteria

- DeepSeek models route through DeepSeek without affecting mock, default OpenAI-compatible, or GLM behavior.
- Docker OOM-killed submissions return and persist `MLE` with `memory_exceeded=true`.
- `adaptive_repair_v1` is selectable and persists attempt-level observability.
- Adaptive repairs use different stages/prompts for `WA`, `RE`, and `TLE`.
- Tooling config is a first-class experiment variable across solve, experiment, compare, and repeat.
- `tooling_codegen_v1` can call `sample_judge` when enabled and remains disabled by default.
- Backend exposes metadata, history, and trace APIs required by the console.
- `web/` provides usable workflows for problems, solve, experiment, compare, repeat, submissions, trace, and token/latency analytics.
- `go test ./...`, frontend unit tests, frontend build, and Playwright smoke tests pass.

## Risks and Guardrails

- Adaptive repair and tooling both touch `AISolveService`; do adaptive attempt persistence first, then tooling, to avoid two simultaneous rewrites.
- MLE detection depends on Docker inspect after run container exit; never use `--rm` for run containers after implementing OOM inspection.
- Do not make tools able to access arbitrary shell/network/filesystem.
- Do not make frontend the source of truth for experiment logic.
- Keep default behavior disabled for DeepSeek route, adaptive repair, and tooling unless explicitly selected.

## 这个计划实现了什么

这个计划把当前所有没完成的工作合并为一条可执行路线。完成后，项目会具备以下能力：

1. **模型接入更完整**：`deepseek-chat` / `deepseek-reasoner` 可以通过 OpenAI-compatible client 自动路由到 DeepSeek，同时保留默认模型和 GLM 路由。
2. **判题语义更准确**：Docker OOM 会被识别为 `MLE`，并通过数据库和 API 暴露 `memory_exceeded`。
3. **AI 解题过程可观察**：`adaptive_repair_v1` 会记录每次生成、判题、失败分类、修复阶段和 token/latency，让一次 solve 不再只是一个最终状态。
4. **实验变量更丰富**：tooling 成为和 model/prompt/agent 同级的实验变量，第一阶段提供 `sample_judge` 和 `tooling_codegen_v1`。
5. **前端可操作**：新增 `web/` 控制台，用表单、历史列表、详情页、trace 和图表替代手写 curl，支持 solve、experiment、compare、repeat 的日常实验工作流。
6. **旧计划被整合**：原来五个分散计划的剩余事项被统一排序、拆批、加验收标准，后续可以按这个计划执行并在完成后删除旧计划文件。

## Execution Handoff

Plan complete and saved to `docs/plans/2026-04-21-unfinished-platform-work.md`. Two execution options:

1. **Subagent-Driven (this session)** - Dispatch a fresh subagent per task, review between tasks, fast iteration.
2. **Parallel Session (separate)** - Open a new session with `executing-plans`, batch execution with checkpoints.

Which approach?
