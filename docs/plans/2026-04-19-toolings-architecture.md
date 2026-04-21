# Toolings Architecture Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a first-class tooling layer so AI solve agents can use bounded, replaceable tools during experiments while keeping model / prompt / agent / tooling variables independently comparable.

**Architecture:** Introduce `internal/tooling` as the only package that defines tool config parsing, tool registration, call limits, and concrete tool execution. `internal/agent` receives a tooling runner through `SolveInput` and may call tools, while `internal/service` resolves request config, persists summary fields, and still owns final submission persistence. `experiment`, `compare`, and `repeat` pass tooling config through as an experiment variable without knowing individual tool implementation details.

**Tech Stack:** Go, GORM, Gin DTOs, existing LLM client, existing judge engine, existing repository/service layering, Go testing package.

---

## Context Review

当前项目已经不是普通 OJ 后端，而是一个 AI 算法题实验平台后端。已经完成的主链包括 OJ 题目与测试点、真实 `cpp17` 判题、AI solve、批量 experiment、baseline / candidate compare、repeat 稳定性实验、基础 verdict 分布、成本统计和请求级 `model / prompt / agent` 变量控制。

之前已经做过的重要工作包括：

- `POST /api/v1/ai/solve` 可以走通 `problem -> prompt -> llm -> extract cpp17 -> judge -> AISolveRun`。
- `direct_codegen`、`direct_codegen_repair`、`analyze_then_codegen` 已经作为最小 agent 变量存在。
- `direct_codegen_repair` 目前的 repair loop 仍在 `AISolveService` 中，而不是一个独立 coordinator。
- experiment / compare / repeat 已经能把 `model / prompt_name / agent_name` 传到底层 solve。
- token 和 latency 已经从 `AISolveRun` 汇总到 experiment / compare / repeat。
- `internal/tooling` 目前只有 `doc.go` 占位。
- `model.ExperimentConfig.ToolingConfig` 和 `model.ExperimentRun.ToolCalls` 已经存在，但当前 service 链路没有真正写入或读取。

这个实验平台的核心目标是做可控制变量、可回看结果、可复现实验。toolings 应当成为一个新的实验变量，而不是散落在 agent 或 service 里的 if/else。

## Scope

本计划只做 tooling 第一阶段：

- 添加 `tooling_config` 作为请求级和实验级变量。
- 添加可替换的 `Tool` 接口、`Registry`、`Runner`。
- 默认不开任何工具，保持现有 baseline 行为不变。
- 先实现一个真实工具：`sample_judge`，只在样例测试点上非持久化运行候选代码。
- 新增一个使用工具的 agent：`tooling_codegen_v1`。
- compare 新增 `tooling` 维度识别。
- 输出和落库只保存 tooling 配置快照与工具调用次数，详细逐步 trace 暂不展开。

## Non-Goals

本阶段不做这些内容：

- 不做异步队列。
- 不做通用 plugin marketplace。
- 不允许 agent 任意访问文件系统、网络或 shell。
- 不做外部搜索工具。
- 不把 final judge 变成 tool，final judge 仍由 `AISolveService` 负责并持久化 submission。
- 不重构 adaptive repair plan 中的 attempt-level coordinator，tooling 设计要能和它后续合并，但不依赖它。

## Tooling Boundary

边界必须保持清楚：

- `internal/tooling` 负责 config、registry、runner、call limit、工具实现。
- `internal/agent` 只依赖 `tooling.Runner` 接口，不创建具体工具。
- `internal/service` 负责解析 `tooling_config`，创建 runner，传给 agent，保存 summary。
- `internal/judge` 和 `internal/sandbox` 不感知 AI agent 或 tooling。
- `internal/handler` 只做 DTO 转换，不理解具体 tool。
- `experiment / compare / repeat` 只传递 tooling config 字符串，不调用 tool。

## Public Config Shape

请求中的 `tooling_config` 使用 JSON object。handler DTO 用 `json.RawMessage` 接收，service 和数据库里保存 canonical JSON string。空字符串、缺省、`null` 都表示 disabled。

```json
{
  "enabled": ["sample_judge"],
  "max_calls": 2,
  "per_tool_max_calls": {
    "sample_judge": 2
  }
}
```

Canonical default should be:

```json
{"enabled":[],"max_calls":0,"per_tool_max_calls":{}}
```

## Suggested First Tool

`sample_judge`：

- 输入：problem ID、候选 `cpp17` 代码。
- 行为：读取题目的 `is_sample=true` 测试点，用现有 judge engine 非持久化评测。
- 输出：verdict、passed count、total count、compile stderr、run stdout、run stderr、error message。
- 没有样例时：返回 `status=skipped`，不报错，不消耗 final judge。
- 只用于 agent 内部反馈，不创建 submission。

---

## Task 1: Add Tooling Config Parser

**Files:**

- Create: `internal/tooling/config.go`
- Create: `internal/tooling/config_test.go`
- Modify: `internal/tooling/doc.go`

**Step 1: Write the failing tests**

Create `internal/tooling/config_test.go`:

```go
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
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/tooling
```

Expected: FAIL because `ResolveConfig` does not exist.

**Step 3: Write minimal implementation**

Create `internal/tooling/config.go`:

```go
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
```

Modify `internal/tooling/doc.go`:

```go
// Package tooling contains bounded tool abstractions exposed to solve agents during experiments.
package tooling
```

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/tooling
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/tooling/config.go internal/tooling/config_test.go internal/tooling/doc.go
git commit -m "feat: add tooling config parser"
```

---

## Task 2: Add Tool Registry And Runner

**Files:**

- Create: `internal/tooling/tool.go`
- Create: `internal/tooling/registry.go`
- Create: `internal/tooling/runner.go`
- Create: `internal/tooling/runner_test.go`

**Step 1: Write the failing tests**

Create `internal/tooling/runner_test.go`:

```go
package tooling

import (
	"context"
	"testing"
)

type fakeTool struct {
	name  string
	calls int
}

func (t *fakeTool) Name() string {
	return t.name
}

func (t *fakeTool) Execute(_ context.Context, input CallInput) (CallResult, error) {
	t.calls++
	return CallResult{
		ToolName: t.name,
		Status:   CallStatusOK,
		Summary:  "called with " + input.SourceCode,
	}, nil
}

func TestRunnerRejectsDisabledTool(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&fakeTool{name: "sample_judge"})
	runner := registry.NewRunner(Config{})

	_, err := runner.Call(context.Background(), "sample_judge", CallInput{SourceCode: "code"})
	if err == nil {
		t.Fatal("expected disabled tool call to fail")
	}
}

func TestRunnerExecutesEnabledToolAndTracksCalls(t *testing.T) {
	tool := &fakeTool{name: "sample_judge"}
	registry := NewRegistry()
	registry.Register(tool)
	runner := registry.NewRunner(Config{
		EnabledTools: []string{"sample_judge"},
		MaxCalls:     2,
	})

	result, err := runner.Call(context.Background(), "sample_judge", CallInput{SourceCode: "code"})
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if result.Status != CallStatusOK {
		t.Fatalf("unexpected result: %+v", result)
	}
	if runner.CallCount() != 1 || tool.calls != 1 {
		t.Fatalf("expected one call, runner=%d tool=%d", runner.CallCount(), tool.calls)
	}
}

func TestRunnerEnforcesGlobalCallLimit(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&fakeTool{name: "sample_judge"})
	runner := registry.NewRunner(Config{
		EnabledTools: []string{"sample_judge"},
		MaxCalls:     1,
	})

	if _, err := runner.Call(context.Background(), "sample_judge", CallInput{}); err != nil {
		t.Fatalf("first call returned error: %v", err)
	}
	if _, err := runner.Call(context.Background(), "sample_judge", CallInput{}); err == nil {
		t.Fatal("expected second call to fail")
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/tooling
```

Expected: FAIL because `Tool`, `Registry`, and `Runner` do not exist.

**Step 3: Write minimal implementation**

Create `internal/tooling/tool.go`:

```go
package tooling

import (
	"context"

	"ai-for-oj/internal/model"
)

const (
	CallStatusOK      = "ok"
	CallStatusSkipped = "skipped"
	CallStatusFailed  = "failed"
)

type Tool interface {
	Name() string
	Execute(ctx context.Context, input CallInput) (CallResult, error)
}

type CallInput struct {
	Problem    *model.Problem
	ProblemID  uint
	SourceCode string
}

type CallResult struct {
	ToolName string `json:"tool_name"`
	Status   string `json:"status"`
	Summary  string `json:"summary,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}
```

Create `internal/tooling/registry.go`:

```go
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
```

Create `internal/tooling/runner.go`:

```go
package tooling

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrToolDisabled     = errors.New("tool disabled")
	ErrToolNotFound     = errors.New("tool not found")
	ErrToolCallLimitHit = errors.New("tool call limit reached")
)

type Runner struct {
	config       Config
	tools        map[string]Tool
	callCount    int
	perToolCalls map[string]int
	results      []CallResult
}

func (r *Runner) Call(ctx context.Context, name string, input CallInput) (CallResult, error) {
	name = strings.TrimSpace(name)
	if name == "" || !r.config.Enabled(name) {
		return CallResult{}, ErrToolDisabled
	}
	if r.config.MaxCalls <= 0 || r.callCount >= r.config.MaxCalls {
		return CallResult{}, ErrToolCallLimitHit
	}
	if limit := r.config.LimitFor(name); limit > 0 && r.perToolCalls[name] >= limit {
		return CallResult{}, ErrToolCallLimitHit
	}
	tool, ok := r.tools[name]
	if !ok {
		return CallResult{}, fmt.Errorf("%w: %s", ErrToolNotFound, name)
	}

	result, err := tool.Execute(ctx, input)
	if result.ToolName == "" {
		result.ToolName = name
	}
	r.callCount++
	r.perToolCalls[name]++
	r.results = append(r.results, result)
	return result, err
}

func (r *Runner) CallCount() int {
	if r == nil {
		return 0
	}
	return r.callCount
}

func (r *Runner) Results() []CallResult {
	if r == nil {
		return nil
	}
	return append([]CallResult(nil), r.results...)
}
```

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/tooling
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/tooling/tool.go internal/tooling/registry.go internal/tooling/runner.go internal/tooling/runner_test.go
git commit -m "feat: add tooling runner registry"
```

---

## Task 3: Implement Sample Judge Tool

**Files:**

- Create: `internal/tooling/sample_judge.go`
- Create: `internal/tooling/sample_judge_test.go`

**Step 1: Write the failing tests**

Create `internal/tooling/sample_judge_test.go`:

```go
package tooling

import (
	"context"
	"strings"
	"testing"

	"ai-for-oj/internal/judge"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

type fakeSampleProblemRepository struct {
	problem *model.Problem
	err     error
}

func (r fakeSampleProblemRepository) Create(context.Context, *model.Problem) error {
	return nil
}

func (r fakeSampleProblemRepository) List(context.Context) ([]model.Problem, error) {
	return nil, nil
}

func (r fakeSampleProblemRepository) GetByID(context.Context, uint) (*model.Problem, error) {
	return r.problem, r.err
}

func (r fakeSampleProblemRepository) GetByIDWithTestCases(context.Context, uint) (*model.Problem, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.problem, nil
}

type fakeSampleJudgeEngine struct {
	req judge.Request
	res judge.Result
	err error
}

func (e *fakeSampleJudgeEngine) Judge(_ context.Context, req judge.Request) (judge.Result, error) {
	e.req = req
	return e.res, e.err
}

func TestSampleJudgeToolRunsOnlySampleCases(t *testing.T) {
	engine := &fakeSampleJudgeEngine{
		res: judge.Result{Verdict: "AC", PassedCount: 1, TotalCount: 1},
	}
	tool := NewSampleJudgeTool(fakeSampleProblemRepository{
		problem: &model.Problem{
			BaseModel: model.BaseModel{ID: 7},
			TestCases: []model.TestCase{
				{BaseModel: model.BaseModel{ID: 1}, Input: "1\n", ExpectedOutput: "1\n", IsSample: true},
				{BaseModel: model.BaseModel{ID: 2}, Input: "2\n", ExpectedOutput: "2\n", IsSample: false},
			},
		},
	}, engine)

	result, err := tool.Execute(context.Background(), CallInput{ProblemID: 7, SourceCode: "int main(){}"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Status != CallStatusOK {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(engine.req.TestCases) != 1 || !engine.req.TestCases[0].IsSample {
		t.Fatalf("expected only sample cases, got %+v", engine.req.TestCases)
	}
	if !strings.Contains(result.Summary, "verdict=AC") {
		t.Fatalf("expected verdict summary, got %q", result.Summary)
	}
}

func TestSampleJudgeToolSkipsWhenNoSamples(t *testing.T) {
	tool := NewSampleJudgeTool(fakeSampleProblemRepository{
		problem: &model.Problem{BaseModel: model.BaseModel{ID: 7}},
	}, &fakeSampleJudgeEngine{})

	result, err := tool.Execute(context.Background(), CallInput{ProblemID: 7, SourceCode: "int main(){}"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Status != CallStatusSkipped {
		t.Fatalf("expected skipped result, got %+v", result)
	}
}

func TestSampleJudgeToolReturnsRepositoryError(t *testing.T) {
	tool := NewSampleJudgeTool(fakeSampleProblemRepository{err: repository.ErrProblemNotFound}, &fakeSampleJudgeEngine{})
	_, err := tool.Execute(context.Background(), CallInput{ProblemID: 7, SourceCode: "int main(){}"})
	if err == nil {
		t.Fatal("expected repository error")
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/tooling
```

Expected: FAIL because `NewSampleJudgeTool` does not exist.

**Step 3: Write minimal implementation**

Create `internal/tooling/sample_judge.go`:

```go
package tooling

import (
	"context"
	"encoding/json"
	"fmt"

	"ai-for-oj/internal/judge"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/repository"
)

const SampleJudgeToolName = "sample_judge"

type SampleJudgeTool struct {
	problems repository.ProblemRepository
	engine   judge.Engine
}

type sampleJudgeMetadata struct {
	Verdict       string `json:"verdict,omitempty"`
	PassedCount   int    `json:"passed_count"`
	TotalCount    int    `json:"total_count"`
	CompileStderr string `json:"compile_stderr,omitempty"`
	RunStdout     string `json:"run_stdout,omitempty"`
	RunStderr     string `json:"run_stderr,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

func NewSampleJudgeTool(problems repository.ProblemRepository, engine judge.Engine) *SampleJudgeTool {
	return &SampleJudgeTool{problems: problems, engine: engine}
}

func (t *SampleJudgeTool) Name() string {
	return SampleJudgeToolName
}

func (t *SampleJudgeTool) Execute(ctx context.Context, input CallInput) (CallResult, error) {
	problem := input.Problem
	if problem == nil {
		var err error
		problem, err = t.problems.GetByIDWithTestCases(ctx, input.ProblemID)
		if err != nil {
			return CallResult{}, err
		}
	}

	samples := sampleCases(problem.TestCases)
	if len(samples) == 0 {
		return CallResult{
			ToolName: SampleJudgeToolName,
			Status:   CallStatusSkipped,
			Summary:  "sample_judge skipped: problem has no sample test cases",
		}, nil
	}

	result, err := t.engine.Judge(ctx, judge.Request{
		Problem:    problem,
		TestCases:  samples,
		Language:   model.LanguageCPP17,
		SourceCode: input.SourceCode,
	})
	if err != nil {
		return CallResult{}, err
	}

	metadata := sampleJudgeMetadata{
		Verdict:       result.Verdict,
		PassedCount:   result.PassedCount,
		TotalCount:    result.TotalCount,
		CompileStderr: result.CompileStderr,
		RunStdout:     result.RunStdout,
		RunStderr:     result.RunStderr,
		ErrorMessage:  result.ErrorMessage,
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		return CallResult{}, fmt.Errorf("marshal sample judge metadata: %w", err)
	}

	return CallResult{
		ToolName: SampleJudgeToolName,
		Status:   CallStatusOK,
		Summary:  fmt.Sprintf("sample_judge verdict=%s passed=%d/%d", result.Verdict, result.PassedCount, result.TotalCount),
		Metadata: string(data),
	}, nil
}

func sampleCases(testCases []model.TestCase) []model.TestCase {
	samples := make([]model.TestCase, 0, len(testCases))
	for _, item := range testCases {
		if item.IsSample {
			samples = append(samples, item)
		}
	}
	return samples
}
```

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/tooling
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/tooling/sample_judge.go internal/tooling/sample_judge_test.go
git commit -m "feat: add sample judge tooling"
```

---

## Task 4: Move C++ Code Extraction Into Agent Package

**Files:**

- Create: `internal/agent/code_extract.go`
- Create: `internal/agent/code_extract_test.go`
- Modify: `internal/service/ai_solve.go`

**Step 1: Write the failing tests**

Create `internal/agent/code_extract_test.go`:

```go
package agent

import "testing"

func TestExtractCPPCodePrefersCPPFence(t *testing.T) {
	raw := "text\n```python\nprint(1)\n```\n```cpp\nint main(){return 0;}\n```"
	got := ExtractCPPCode(raw)
	if got != "int main(){return 0;}" {
		t.Fatalf("unexpected code: %q", got)
	}
}

func TestExtractCPPCodeFallsBackToGenericFence(t *testing.T) {
	got := ExtractCPPCode("```\nint main(){return 0;}\n```")
	if got != "int main(){return 0;}" {
		t.Fatalf("unexpected code: %q", got)
	}
}

func TestExtractCPPCodeReturnsEmptyWithoutFence(t *testing.T) {
	if got := ExtractCPPCode("int main(){return 0;}"); got != "" {
		t.Fatalf("expected empty code, got %q", got)
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/agent
```

Expected: FAIL because `ExtractCPPCode` does not exist.

**Step 3: Write minimal implementation**

Create `internal/agent/code_extract.go`:

```go
package agent

import (
	"regexp"
	"strings"
)

var (
	cppFencePattern     = regexp.MustCompile("(?is)```(?:cpp|c\\+\\+|cc|cxx)\\s*(.*?)```")
	genericFencePattern = regexp.MustCompile("(?is)```(?:[a-z0-9_+-]+)?\\s*(.*?)```")
)

func ExtractCPPCode(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if matches := cppFencePattern.FindStringSubmatch(raw); len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	if matches := genericFencePattern.FindStringSubmatch(raw); len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}
```

Modify `internal/service/ai_solve.go`:

```go
code := agent.ExtractCPPCode(agentOutput.RawResponse)
```

Remove the old local `cppFencePattern`, `genericFencePattern`, and `extractCPPCode` definitions from `internal/service/ai_solve.go`.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/agent ./internal/service
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/code_extract.go internal/agent/code_extract_test.go internal/service/ai_solve.go
git commit -m "refactor: share cpp code extraction"
```

---

## Task 5: Add Tooling Persistence And API Fields

**Files:**

- Modify: `internal/model/ai_solve_run.go`
- Modify: `internal/model/experiment.go`
- Modify: `internal/model/experiment_repeat.go`
- Modify: `internal/model/experiment_compare.go`
- Modify: `internal/handler/dto/ai_solve.go`
- Modify: `internal/handler/dto/experiment.go`
- Modify: `internal/handler/dto/experiment_compare.go`
- Modify: `internal/handler/dto/experiment_repeat.go`

**Step 1: Write the failing tests**

Add assertions to existing service and handler-facing tests after the service fields are wired in later. For this task, add a small model compile-time field test.

Create `internal/model/schema_test.go` if it does not already exist:

```go
package model

import "testing"

func TestToolingFieldsExistOnRuntimeModels(t *testing.T) {
	run := AISolveRun{
		ToolingConfig: "{}",
		ToolCallCount: 1,
	}
	experiment := Experiment{ToolingConfig: "{}"}
	repeat := ExperimentRepeat{ToolingConfig: "{}"}
	compare := ExperimentCompare{
		BaselineToolingConfig:  "{}",
		CandidateToolingConfig: `{"enabled":["sample_judge"],"max_calls":1}`,
	}

	if run.ToolingConfig == "" || run.ToolCallCount != 1 {
		t.Fatalf("unexpected run tooling fields: %+v", run)
	}
	if experiment.ToolingConfig == "" || repeat.ToolingConfig == "" {
		t.Fatalf("unexpected experiment tooling fields: %+v %+v", experiment, repeat)
	}
	if compare.BaselineToolingConfig == "" || compare.CandidateToolingConfig == "" {
		t.Fatalf("unexpected compare tooling fields: %+v", compare)
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/model
```

Expected: FAIL because the tooling fields do not exist on runtime models yet.

**Step 3: Add minimal fields**

Modify `internal/model/ai_solve_run.go`:

```go
ToolingConfig string `gorm:"column:tooling_config;type:longtext;not null" json:"tooling_config"`
ToolCallCount int    `gorm:"column:tool_call_count;not null;default:0" json:"tool_call_count"`
```

Modify `internal/model/experiment.go`:

```go
ToolingConfig string `gorm:"column:tooling_config;type:longtext;not null" json:"tooling_config"`
```

Modify `internal/model/experiment_repeat.go`:

```go
ToolingConfig string `gorm:"column:tooling_config;type:longtext;not null" json:"tooling_config"`
```

Modify `internal/model/experiment_compare.go`:

```go
BaselineToolingConfig  string `gorm:"column:baseline_tooling_config;type:longtext;not null" json:"baseline_tooling_config"`
CandidateToolingConfig string `gorm:"column:candidate_tooling_config;type:longtext;not null" json:"candidate_tooling_config"`
```

Modify DTOs so requests and responses carry the same variable.

In `internal/handler/dto/ai_solve.go`, request DTOs should use `json.RawMessage` and response DTOs should use canonical string fields:

```go
import "encoding/json"

type AISolveRequest struct {
	ToolingConfig json.RawMessage `json:"tooling_config"`
}

type AISolveResponse struct {
	ToolingConfig string `json:"tooling_config"`
	ToolCallCount int    `json:"tool_call_count"`
}
```

In experiment request DTOs add `ToolingConfig json.RawMessage`, and in compare request DTOs add `BaselineToolingConfig json.RawMessage` and `CandidateToolingConfig json.RawMessage`. Response DTOs should return canonical strings.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/model ./internal/handler
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/model/ai_solve_run.go internal/model/experiment.go internal/model/experiment_repeat.go internal/model/experiment_compare.go internal/model/schema_test.go internal/handler/dto/ai_solve.go internal/handler/dto/experiment.go internal/handler/dto/experiment_compare.go internal/handler/dto/experiment_repeat.go
git commit -m "feat: add tooling api fields"
```

---

## Task 6: Pass Tooling Through AI Solve Service

**Files:**

- Modify: `internal/agent/solve.go`
- Modify: `internal/service/ai_solve.go`
- Modify: `internal/service/ai_solve_test.go`
- Modify: `internal/handler/ai.go`
- Modify: `internal/bootstrap/app.go`

**Step 1: Write the failing service test**

Extend `TestAISolveServiceSolve` in `internal/service/ai_solve_test.go`:

```go
output, err := service.Solve(context.Background(), AISolveInput{
	ProblemID:      1,
	ToolingConfig: `{"enabled":["sample_judge"],"max_calls":1}`,
})
if err != nil {
	t.Fatalf("solve returned error: %v", err)
}
if output.ToolingConfig != `{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}` {
	t.Fatalf("unexpected tooling config: %s", output.ToolingConfig)
}
if len(runRepo.created) != 1 || runRepo.created[0].ToolingConfig != output.ToolingConfig {
	t.Fatalf("expected tooling config to be persisted on create, got %+v", runRepo.created)
}
```

Add a new rejection test:

```go
func TestAISolveServiceSolveRejectsInvalidToolingConfig(t *testing.T) {
	service := NewAISolveService(
		fakeProblemRepository{},
		&fakeAISolveRunRepository{},
		&fakeLLMClient{},
		&fakeJudgeSubmitter{},
		"default-model",
		nil,
	)

	output, err := service.Solve(context.Background(), AISolveInput{
		ProblemID:      1,
		ToolingConfig: `{"enabled":`,
	})
	if err == nil {
		t.Fatal("expected invalid tooling config to fail")
	}
	if output != nil {
		t.Fatalf("expected nil output before run creation, got %+v", output)
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/service
```

Expected: FAIL because `ToolingConfig` is not in `AISolveInput`, constructor signature has not changed, and service does not parse config.

**Step 3: Write minimal implementation**

Modify `internal/agent/solve.go`:

```go
import "ai-for-oj/internal/tooling"

type SolveInput struct {
	Problem       *model.Problem
	Model         string
	PromptName    string
	ToolingRunner *tooling.Runner
}

type SolveOutput struct {
	AgentName       string
	Model           string
	PromptPreview   string
	RawResponse     string
	ExtractedCode   string
	TokenInput      int64
	TokenOutput     int64
	LLMLatencyMS    int
	AnalysisPreview string
	ToolCallCount   int
	ToolResults     []tooling.CallResult
}
```

Modify `internal/service/ai_solve.go`:

```go
import "ai-for-oj/internal/tooling"

type AISolveInput struct {
	ProblemID      uint
	Model          string
	PromptName     string
	AgentName      string
	ToolingConfig string
}

type AISolveOutput struct {
	ToolingConfig string `json:"tooling_config"`
	ToolCallCount int    `json:"tool_call_count"`
}

type AISolveService struct {
	problems        repository.ProblemRepository
	runs            repository.AISolveRunRepository
	llmClient       llm.Client
	submissions     JudgeSubmitter
	defaultModel    string
	toolingRegistry *tooling.Registry
}
```

Update constructor:

```go
func NewAISolveService(
	problems repository.ProblemRepository,
	runs repository.AISolveRunRepository,
	llmClient llm.Client,
	submissions JudgeSubmitter,
	defaultModel string,
	toolingRegistry *tooling.Registry,
) *AISolveService {
	if toolingRegistry == nil {
		toolingRegistry = tooling.NewRegistry()
	}
	return &AISolveService{
		problems:        problems,
		runs:            runs,
		llmClient:       llmClient,
		submissions:     submissions,
		defaultModel:    defaultModel,
		toolingRegistry: toolingRegistry,
	}
}
```

At the start of `Solve` after resolving model / prompt / agent:

```go
toolingConfig, canonicalToolingConfig, err := tooling.ResolveConfig(input.ToolingConfig)
if err != nil {
	return nil, err
}
toolingRunner := s.toolingRegistry.NewRunner(toolingConfig)
```

Set fields on create:

```go
run := &model.AISolveRun{
	ProblemID:      input.ProblemID,
	Model:          resolvedModel,
	PromptName:     resolvedPromptName,
	AgentName:      resolvedAgentName,
	ToolingConfig: canonicalToolingConfig,
	Status:         model.AISolveRunStatusRunning,
}
```

Pass runner into agent:

```go
agentOutput, err := strategy.Execute(solveCtx, s.llmClient, agent.SolveInput{
	Problem:       problem,
	Model:         resolvedModel,
	PromptName:    resolvedPromptName,
	ToolingRunner: toolingRunner,
})
```

In `applyAttemptLLMOutput`:

```go
run.ToolCallCount += attempt.ToolCallCount
output.ToolCallCount = run.ToolCallCount
```

When extracting code:

```go
code := firstNonEmpty(agentOutput.ExtractedCode, agent.ExtractCPPCode(agentOutput.RawResponse))
```

Update `syncAISolveOutputFromRun`:

```go
output.ToolingConfig = run.ToolingConfig
output.ToolCallCount = run.ToolCallCount
```

Modify `internal/handler/ai.go` to pass `string(req.ToolingConfig)` into service and return response `ToolingConfig` / `ToolCallCount`.

Modify `internal/bootstrap/app.go`:

```go
toolingRegistry := tooling.NewRegistry()
aiSolveService := service.NewAISolveService(problemRepository, aiSolveRunRepository, llmClient, judgeSubmissionService, cfg.LLM.Model, toolingRegistry)
```

Do not register `sample_judge` in this task. That keeps the service test independent.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/service ./internal/handler ./internal/bootstrap
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/solve.go internal/service/ai_solve.go internal/service/ai_solve_test.go internal/handler/ai.go internal/bootstrap/app.go
git commit -m "feat: pass tooling config through ai solve"
```

---

## Task 7: Add Tooling Codegen Agent

**Files:**

- Modify: `internal/agent/solve.go`
- Create: `internal/agent/tooling_codegen.go`
- Create: `internal/agent/tooling_codegen_test.go`
- Modify: `internal/prompt/solve.go`

**Step 1: Write the failing tests**

Create `internal/agent/tooling_codegen_test.go`:

```go
package agent

import (
	"context"
	"strings"
	"testing"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/model"
	"ai-for-oj/internal/tooling"
)

type fakeToolingLLM struct {
	requests  []llm.GenerateRequest
	responses []llm.GenerateResponse
}

func (c *fakeToolingLLM) Generate(_ context.Context, req llm.GenerateRequest) (llm.GenerateResponse, error) {
	c.requests = append(c.requests, req)
	index := len(c.requests) - 1
	if index < len(c.responses) {
		return c.responses[index], nil
	}
	return llm.GenerateResponse{}, nil
}

type staticTool struct {
	result tooling.CallResult
}

func (t staticTool) Name() string {
	return tooling.SampleJudgeToolName
}

func (t staticTool) Execute(context.Context, tooling.CallInput) (tooling.CallResult, error) {
	return t.result, nil
}

func TestResolveToolingCodegenAgent(t *testing.T) {
	name, err := ResolveSolveAgentName(ToolingCodegenV1AgentName)
	if err != nil {
		t.Fatalf("ResolveSolveAgentName returned error: %v", err)
	}
	if name != ToolingCodegenV1AgentName {
		t.Fatalf("unexpected agent name: %s", name)
	}
}

func TestToolingCodegenRepairsAfterFailedSampleJudge(t *testing.T) {
	client := &fakeToolingLLM{
		responses: []llm.GenerateResponse{
			{Model: "mock", Content: "```cpp\nint main(){return 1;}\n```", InputTokens: 10, OutputTokens: 5},
			{Model: "mock", Content: "```cpp\nint main(){return 0;}\n```", InputTokens: 12, OutputTokens: 6},
		},
	}
	registry := tooling.NewRegistry()
	registry.Register(staticTool{result: tooling.CallResult{
		ToolName: tooling.SampleJudgeToolName,
		Status:   tooling.CallStatusOK,
		Summary:  "sample_judge verdict=WA passed=0/1",
	}})
	runner := registry.NewRunner(tooling.Config{
		EnabledTools: []string{tooling.SampleJudgeToolName},
		MaxCalls:     1,
	})

	strategy, err := ResolveSolveStrategy(ToolingCodegenV1AgentName)
	if err != nil {
		t.Fatalf("ResolveSolveStrategy returned error: %v", err)
	}
	output, err := strategy.Execute(context.Background(), client, SolveInput{
		Problem:       &model.Problem{Title: "A+B"},
		Model:         "mock",
		PromptName:    "default",
		ToolingRunner: runner,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if output.ToolCallCount != 1 {
		t.Fatalf("expected one tool call, got %d", output.ToolCallCount)
	}
	if output.ExtractedCode != "int main(){return 0;}" {
		t.Fatalf("expected repaired code, got %q", output.ExtractedCode)
	}
	if len(client.requests) != 2 {
		t.Fatalf("expected two llm requests, got %d", len(client.requests))
	}
	if !strings.Contains(client.requests[1].Prompt, "sample_judge verdict=WA") {
		t.Fatalf("expected repair prompt to include tool feedback, got %q", client.requests[1].Prompt)
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/agent
```

Expected: FAIL because `ToolingCodegenV1AgentName` and strategy do not exist.

**Step 3: Write minimal implementation**

Modify `internal/agent/solve.go`:

```go
const (
	DirectCodegenAgentName       = "direct_codegen"
	DirectCodegenRepairAgentName = "direct_codegen_repair"
	AnalyzeThenCodegenAgentName  = "analyze_then_codegen"
	ToolingCodegenV1AgentName    = "tooling_codegen_v1"
)
```

Add to resolver switches:

```go
case ToolingCodegenV1AgentName:
	return ToolingCodegenV1AgentName, nil
```

```go
case ToolingCodegenV1AgentName:
	return toolingCodegenV1Strategy{}, nil
```

Modify `internal/prompt/solve.go`:

```go
func BuildToolingRepairPrompt(problem *model.Problem, promptName, previousCode, toolFeedback string) string {
	base := BuildSolvePrompt(problem, promptName)
	return strings.TrimSpace(fmt.Sprintf(`
%s

Your previous candidate solution was checked by an internal tool before final submission.
Use the tool feedback below to repair the code.
Return a complete submit-ready C++17 program as exactly one markdown cpp code block.

Previous Code (cpp):
%s

Tool Feedback:
%s
`, base, strings.TrimSpace(previousCode), strings.TrimSpace(toolFeedback)))
}
```

Create `internal/agent/tooling_codegen.go`:

```go
package agent

import (
	"context"
	"strings"

	"ai-for-oj/internal/llm"
	"ai-for-oj/internal/prompt"
	"ai-for-oj/internal/tooling"
)

type toolingCodegenV1Strategy struct{}

func (toolingCodegenV1Strategy) Name() string {
	return ToolingCodegenV1AgentName
}

func (toolingCodegenV1Strategy) Execute(ctx context.Context, client llm.Client, input SolveInput) (SolveOutput, error) {
	finalPrompt := prompt.BuildSolvePrompt(input.Problem, input.PromptName)
	resp, latencyMS, err := generateOnce(ctx, client, input.Model, finalPrompt)
	if err != nil {
		return SolveOutput{
			AgentName:     ToolingCodegenV1AgentName,
			Model:         input.Model,
			PromptPreview: finalPrompt,
			LLMLatencyMS:  latencyMS,
		}, err
	}

	output := SolveOutput{
		AgentName:     ToolingCodegenV1AgentName,
		Model:         effectiveModel(resp.Model, input.Model),
		PromptPreview: finalPrompt,
		RawResponse:   resp.Content,
		ExtractedCode: ExtractCPPCode(resp.Content),
		TokenInput:    resp.InputTokens,
		TokenOutput:   resp.OutputTokens,
		LLMLatencyMS:  latencyMS,
	}
	if input.ToolingRunner == nil || strings.TrimSpace(output.ExtractedCode) == "" {
		return output, nil
	}

	toolResult, toolErr := input.ToolingRunner.Call(ctx, tooling.SampleJudgeToolName, tooling.CallInput{
		Problem:    input.Problem,
		ProblemID:  input.Problem.ID,
		SourceCode: output.ExtractedCode,
	})
	if toolErr != nil {
		return output, nil
	}
	output.ToolCallCount = input.ToolingRunner.CallCount()
	output.ToolResults = input.ToolingRunner.Results()
	if toolResult.Status != tooling.CallStatusOK || !strings.Contains(toolResult.Summary, "verdict=AC") {
		return repairAfterToolFeedback(ctx, client, input, output, toolResult.Summary)
	}
	return output, nil
}

func repairAfterToolFeedback(ctx context.Context, client llm.Client, input SolveInput, previous SolveOutput, feedback string) (SolveOutput, error) {
	repairPrompt := prompt.BuildToolingRepairPrompt(input.Problem, input.PromptName, previous.ExtractedCode, feedback)
	resp, latencyMS, err := generateOnce(ctx, client, input.Model, repairPrompt)
	previous.PromptPreview = repairPrompt
	previous.LLMLatencyMS += latencyMS
	if err != nil {
		return previous, err
	}
	previous.Model = effectiveModel(resp.Model, previous.Model, input.Model)
	previous.RawResponse = resp.Content
	previous.ExtractedCode = ExtractCPPCode(resp.Content)
	previous.TokenInput += resp.InputTokens
	previous.TokenOutput += resp.OutputTokens
	return previous, nil
}
```

This implementation intentionally swallows tool execution errors in phase 1, because tooling should not make baseline solve less reliable. Later we can make this configurable.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/agent ./internal/prompt
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/solve.go internal/agent/tooling_codegen.go internal/agent/tooling_codegen_test.go internal/prompt/solve.go
git commit -m "feat: add tooling codegen agent"
```

---

## Task 8: Register Sample Judge Tool In Bootstrap

**Files:**

- Modify: `internal/bootstrap/app.go`
- Test: `internal/bootstrap/app.go` through package build

**Step 1: Write the failing integration expectation**

No new unit test is required here because `bootstrap.Build` depends on database and Docker wiring. The expected compile-time failure is enough after adding imports incorrectly.

**Step 2: Run package build before edit**

Run:

```bash
go test ./internal/bootstrap
```

Expected: PASS before edit.

**Step 3: Register tool**

Modify `internal/bootstrap/app.go` imports:

```go
import "ai-for-oj/internal/tooling"
```

After `judgeEngine := judge.NewEngine(sandboxExecutor)`:

```go
toolingRegistry := tooling.NewRegistry()
toolingRegistry.Register(tooling.NewSampleJudgeTool(problemRepository, judgeEngine))
```

Pass registry into AI solve service:

```go
aiSolveService := service.NewAISolveService(problemRepository, aiSolveRunRepository, llmClient, judgeSubmissionService, cfg.LLM.Model, toolingRegistry)
```

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/bootstrap ./internal/tooling
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/bootstrap/app.go
git commit -m "feat: register sample judge tooling"
```

---

## Task 9: Pass Tooling Through Experiment Run

**Files:**

- Modify: `internal/service/experiment.go`
- Modify: `internal/service/experiment_test.go`
- Modify: `internal/handler/experiment.go`
- Modify: `internal/handler/dto/experiment.go`

**Step 1: Write the failing experiment test**

Extend `TestExperimentServiceRunPassesModelPromptAndAgentToAISolve` in `internal/service/experiment_test.go`:

```go
input := RunExperimentInput{
	Name:          "tooling-exp",
	ProblemIDs:    []uint{1, 2},
	Model:         "mock-a",
	PromptName:    prompt.StrictCPP17SolvePromptName,
	AgentName:     agent.ToolingCodegenV1AgentName,
	ToolingConfig: `{"enabled":["sample_judge"],"max_calls":1}`,
}
```

After run:

```go
if output.ToolingConfig != `{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}` {
	t.Fatalf("expected canonical tooling config in output, got %q", output.ToolingConfig)
}
if aiSolver.inputs[0].ToolingConfig != output.ToolingConfig || aiSolver.inputs[1].ToolingConfig != output.ToolingConfig {
	t.Fatalf("expected tooling config to be passed to every solve, got %+v", aiSolver.inputs)
}
```

Update fake AI solver output in tests:

```go
ToolCallCount: 1,
```

Assert experiment run output:

```go
if output.Runs[0].ToolCallCount != 1 {
	t.Fatalf("expected tool call count on run output, got %+v", output.Runs[0])
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/service
```

Expected: FAIL because experiment input and output do not include tooling fields.

**Step 3: Write minimal implementation**

Modify `internal/service/experiment.go`:

```go
type RunExperimentInput struct {
	Name          string
	ProblemIDs    []uint
	Model         string
	PromptName    string
	AgentName     string
	ToolingConfig string
}
```

Resolve config at the start of `Run`:

```go
_, canonicalToolingConfig, err := tooling.ResolveConfig(input.ToolingConfig)
if err != nil {
	return nil, err
}
```

Set experiment field:

```go
ToolingConfig: canonicalToolingConfig,
```

Pass to solve:

```go
ToolingConfig: experiment.ToolingConfig,
```

Set run summary:

```go
run.ToolCalls = aiOutput.ToolCallCount
```

Extend `ExperimentRunOutput`:

```go
ToolCallCount int `json:"tool_call_count"`
```

Extend `ExperimentOutput`:

```go
ToolingConfig string `json:"tooling_config"`
```

Map fields in `toExperimentOutput`.

Modify `internal/handler/experiment.go` to pass `string(req.ToolingConfig)` into `RunExperimentInput` and map response `ToolingConfig` / `ToolCallCount`.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/service ./internal/handler
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/experiment.go internal/service/experiment_test.go internal/handler/experiment.go internal/handler/dto/experiment.go
git commit -m "feat: pass tooling config through experiments"
```

---

## Task 10: Pass Tooling Through Compare And Add Tooling Dimension

**Files:**

- Modify: `internal/service/experiment_compare.go`
- Modify: `internal/service/experiment_compare_test.go`
- Modify: `internal/handler/experiment.go`
- Modify: `internal/handler/dto/experiment_compare.go`

**Step 1: Write the failing compare test**

Add a test to `internal/service/experiment_compare_test.go`:

```go
func TestExperimentCompareServiceDetectsToolingDimension(t *testing.T) {
	runner := &fakeExperimentRunner{
		outputs: []*ExperimentOutput{
			{ID: 10, Model: "mock-a", PromptName: prompt.DefaultSolvePromptName, AgentName: agent.ToolingCodegenV1AgentName, ToolingConfig: `{"enabled":[],"max_calls":0,"per_tool_max_calls":{}}`},
			{ID: 20, Model: "mock-a", PromptName: prompt.DefaultSolvePromptName, AgentName: agent.ToolingCodegenV1AgentName, ToolingConfig: `{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}`},
		},
	}
	compares := &fakeExperimentCompareRepository{}
	service := NewExperimentCompareService(compares, runner, "mock-a")

	output, err := service.Compare(context.Background(), CompareExperimentInput{
		Name:                   "tooling-compare",
		ProblemIDs:             []uint{1},
		BaselineModel:          "mock-a",
		CandidateModel:         "mock-a",
		BaselinePromptName:     prompt.DefaultSolvePromptName,
		CandidatePromptName:    prompt.DefaultSolvePromptName,
		BaselineAgentName:      agent.ToolingCodegenV1AgentName,
		CandidateAgentName:     agent.ToolingCodegenV1AgentName,
		BaselineToolingConfig:  `{}`,
		CandidateToolingConfig: `{"enabled":["sample_judge"],"max_calls":1}`,
	})
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}
	if output.CompareDimension != ExperimentCompareDimensionTooling {
		t.Fatalf("expected tooling dimension, got %q", output.CompareDimension)
	}
	if runner.runInputs[0].ToolingConfig != `{"enabled":[],"max_calls":0,"per_tool_max_calls":{}}` {
		t.Fatalf("unexpected baseline tooling config: %q", runner.runInputs[0].ToolingConfig)
	}
	if runner.runInputs[1].ToolingConfig != `{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}` {
		t.Fatalf("unexpected candidate tooling config: %q", runner.runInputs[1].ToolingConfig)
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/service
```

Expected: FAIL because compare input has no tooling config and dimension resolver has no tooling dimension.

**Step 3: Write minimal implementation**

Modify constants:

```go
const (
	ExperimentCompareDimensionModel   = "model"
	ExperimentCompareDimensionPrompt  = "prompt"
	ExperimentCompareDimensionAgent   = "agent"
	ExperimentCompareDimensionTooling = "tooling"
	ExperimentCompareDimensionMixed   = "mixed"
)
```

Modify `CompareExperimentInput`:

```go
BaselineToolingConfig  string
CandidateToolingConfig string
```

Resolve canonical configs in `Compare`:

```go
_, baselineToolingConfig, err := tooling.ResolveConfig(input.BaselineToolingConfig)
if err != nil {
	return nil, err
}
_, candidateToolingConfig, err := tooling.ResolveConfig(input.CandidateToolingConfig)
if err != nil {
	return nil, err
}
```

Persist on `model.ExperimentCompare`:

```go
BaselineToolingConfig:  baselineToolingConfig,
CandidateToolingConfig: candidateToolingConfig,
```

Pass into baseline / candidate `RunExperimentInput`.

Update dimension resolver signature:

```go
func resolveCompareDimensionAndValues(
	baselineModel, candidateModel string,
	baselinePromptName, candidatePromptName string,
	baselineAgentName, candidateAgentName string,
	baselineToolingConfig, candidateToolingConfig string,
) (string, string, string)
```

Implement simple resolver order:

```go
if baselineModel != candidateModel && baselinePromptName == candidatePromptName && baselineAgentName == candidateAgentName && baselineToolingConfig == candidateToolingConfig {
	return ExperimentCompareDimensionModel, baselineModel, candidateModel
}
if baselineModel == candidateModel && baselinePromptName != candidatePromptName && baselineAgentName == candidateAgentName && baselineToolingConfig == candidateToolingConfig {
	return ExperimentCompareDimensionPrompt, baselinePromptName, candidatePromptName
}
if baselineModel == candidateModel && baselinePromptName == candidatePromptName && baselineAgentName != candidateAgentName && baselineToolingConfig == candidateToolingConfig {
	return ExperimentCompareDimensionAgent, baselineAgentName, candidateAgentName
}
if baselineModel == candidateModel && baselinePromptName == candidatePromptName && baselineAgentName == candidateAgentName && baselineToolingConfig != candidateToolingConfig {
	return ExperimentCompareDimensionTooling, baselineToolingConfig, candidateToolingConfig
}
return ExperimentCompareDimensionMixed, "baseline", "candidate"
```

Extend `ExperimentCompareOutput` and DTO response with baseline / candidate tooling config fields.

Modify `internal/handler/experiment.go` to pass compare request tooling configs through with `string(req.BaselineToolingConfig)` and `string(req.CandidateToolingConfig)`.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/service ./internal/handler
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/experiment_compare.go internal/service/experiment_compare_test.go internal/handler/experiment.go internal/handler/dto/experiment_compare.go
git commit -m "feat: compare tooling experiment variable"
```

---

## Task 11: Pass Tooling Through Repeat

**Files:**

- Modify: `internal/service/experiment_repeat.go`
- Modify: `internal/service/experiment_repeat_test.go`
- Modify: `internal/handler/experiment.go`
- Modify: `internal/handler/dto/experiment_repeat.go`

**Step 1: Write the failing repeat test**

Extend `TestExperimentRepeatServiceRepeatPassesModelPromptAndAgentToRounds` in `internal/service/experiment_repeat_test.go`:

```go
input := RepeatExperimentInput{
	Name:          "repeat-with-tools",
	ProblemIDs:    []uint{1, 2},
	Model:         "mock-a",
	PromptName:    prompt.StrictCPP17SolvePromptName,
	AgentName:     agent.ToolingCodegenV1AgentName,
	ToolingConfig: `{"enabled":["sample_judge"],"max_calls":1}`,
	RepeatCount:   2,
}
```

Assert:

```go
if output.ToolingConfig != `{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}` {
	t.Fatalf("unexpected repeat tooling config: %q", output.ToolingConfig)
}
if runner.runInputs[0].ToolingConfig != output.ToolingConfig || runner.runInputs[1].ToolingConfig != output.ToolingConfig {
	t.Fatalf("expected tooling config in every round, got %+v", runner.runInputs)
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
go test ./internal/service
```

Expected: FAIL because repeat input and output do not include tooling config.

**Step 3: Write minimal implementation**

Modify `RepeatExperimentInput`:

```go
ToolingConfig string
```

Resolve canonical config in `Repeat`:

```go
_, canonicalToolingConfig, err := tooling.ResolveConfig(input.ToolingConfig)
if err != nil {
	return nil, err
}
```

Persist:

```go
ToolingConfig: canonicalToolingConfig,
```

Pass to each round:

```go
ToolingConfig: repeat.ToolingConfig,
```

Extend `ExperimentRepeatOutput`:

```go
ToolingConfig string `json:"tooling_config"`
```

Map response in `buildExperimentRepeatOutput`.

Modify handler and DTO to pass `string(req.ToolingConfig)` into service and return the canonical response field.

**Step 4: Run test to verify it passes**

Run:

```bash
go test ./internal/service ./internal/handler
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/experiment_repeat.go internal/service/experiment_repeat_test.go internal/handler/experiment.go internal/handler/dto/experiment_repeat.go
git commit -m "feat: pass tooling config through repeats"
```

---

## Task 12: Document Usage And Verify Full Build

**Files:**

- Modify: `README.md`
- Modify: `docs/dev_progress.md`
- Modify: `configs/config.example.yaml`

**Step 1: Update README with API examples**

Add a short section under AI Solve:

````markdown
### Tooling-enabled solve

`tooling_config` is disabled by default. To let an agent use sample-only pre-judge feedback, use `tooling_codegen_v1` with `sample_judge`:

```bash
curl --noproxy '*' -sS -X POST http://127.0.0.1:8080/api/v1/ai/solve \
  -H 'Content-Type: application/json' \
  -d '{
    "problem_id": 5,
    "model": "mock-cpp17",
    "prompt_name": "strict_cpp17",
    "agent_name": "tooling_codegen_v1",
    "tooling_config": {
      "enabled": ["sample_judge"],
      "max_calls": 1
    }
  }'
```

The response includes `tooling_config` and `tool_call_count`. Final submissions are still created only by the AI solve service after the agent returns its final code.
````

If DTO keeps `tooling_config` as a string rather than object, use this request shape instead:

```json
{
  "problem_id": 5,
  "agent_name": "tooling_codegen_v1",
  "tooling_config": "{\"enabled\":[\"sample_judge\"],\"max_calls\":1}"
}
```

Prefer object-shaped DTO if implementation time allows. It is easier for API users.

**Step 2: Update dev progress**

Append a dated note:

```markdown
## 2026-04-19 开发补充

新增 tooling 第一阶段设计：tooling config、registry、runner、sample_judge、tooling_codegen_v1，并将 tooling 作为 experiment / compare / repeat 的可控变量。默认不开启工具，现有 direct / repair / analyze agent 行为保持不变。
```

**Step 3: Update config example**

Add a comment only. Do not add global tooling defaults unless the implementation actually reads them.

```yaml
# Tooling is request-level in this phase.
# Example API tooling_config:
# {"enabled":["sample_judge"],"max_calls":1}
```

**Step 4: Run full verification**

Run:

```bash
go test ./...
```

Expected: PASS

Run:

```bash
go build -o /tmp/ai-for-oj-server ./cmd/server
```

Expected: PASS

**Step 5: Commit**

```bash
git add README.md docs/dev_progress.md configs/config.example.yaml
git commit -m "docs: document tooling experiments"
```

---

## Final Manual Smoke Test

Use local Docker setup if available:

```bash
docker compose up -d --build
```

Expected: app and mysql containers start successfully.

Create a problem with at least one sample testcase and one hidden testcase, then run:

```bash
curl --noproxy '*' -sS -X POST http://127.0.0.1:8080/api/v1/ai/solve \
  -H 'Content-Type: application/json' \
  -d '{"problem_id":5,"agent_name":"tooling_codegen_v1","tooling_config":"{\"enabled\":[\"sample_judge\"],\"max_calls\":1}"}'
```

Expected:

- Response has `agent_name=tooling_codegen_v1`.
- Response has canonical `tooling_config`.
- Response has `tool_call_count=1` when sample cases exist.
- `GET /api/v1/ai/solve-runs/:id` returns the same `tooling_config` and `tool_call_count`.
- A final submission is still created once by `AISolveService`.

Run a compare:

```bash
curl --noproxy '*' -sS -X POST http://127.0.0.1:8080/api/v1/experiments/compare \
  -H 'Content-Type: application/json' \
  -d '{
    "name":"tooling-ab",
    "problem_ids":[5],
    "baseline_agent_name":"tooling_codegen_v1",
    "candidate_agent_name":"tooling_codegen_v1",
    "baseline_tooling_config":"{}",
    "candidate_tooling_config":"{\"enabled\":[\"sample_judge\"],\"max_calls\":1}"
  }'
```

Expected:

- `compare_dimension=tooling`
- baseline and candidate summaries both exist
- candidate summary includes canonical sample judge tooling config

## Implementation Notes

- Keep default disabled. Existing experiments must not change behavior unless `tooling_config` is provided and a tooling-aware agent is selected.
- If a tool call fails, `tooling_codegen_v1` should continue with the original generated code in phase 1. Tool failure should be observable through logs later, but not terminal by default.
- `sample_judge` must not write submission rows. It uses `judge.Engine` directly with sample test cases only.
- Do not make `direct_codegen` call tools. Additive agent behavior keeps baselines clean.
- Do not wire `ExperimentConfig` yet unless you are ready to migrate the whole experiment config path. The active runtime path currently stores variables directly on `Experiment`, `ExperimentCompare`, and `ExperimentRepeat`.
- Later, when adaptive repair coordinator is implemented, it should consume the same `tooling.Runner` interface instead of defining a second tool abstraction.

## Acceptance Criteria

- `go test ./...` passes.
- `go build -o /tmp/ai-for-oj-server ./cmd/server` passes.
- Existing AI solve requests without `tooling_config` still behave exactly like before.
- Existing experiment / compare / repeat requests without `tooling_config` still behave exactly like before.
- `tooling_codegen_v1` can call `sample_judge` when enabled.
- `sample_judge` only runs sample test cases and creates no submission.
- AI solve run detail returns `tooling_config` and `tool_call_count`.
- Experiment run output returns `tool_call_count`.
- Compare can identify `tooling` as the changed dimension.
