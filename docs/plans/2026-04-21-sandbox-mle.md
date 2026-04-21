# OJ Sandbox MLE Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Make the OJ sandbox reliably detect memory limit exceeded executions and report them as `MLE` instead of generic `RE`.

**Architecture:** Keep the current Docker-based sandbox, but make run containers inspectable long enough to read Docker's `State.OOMKilled` flag before cleanup. Propagate that signal through `sandbox.RunResult`, `judge.Result`, persisted judge records, and API responses. This plan intentionally does not implement precise peak RSS accounting; it records `MemoryKB` conservatively and focuses first on correct MLE verdict semantics.

**Tech Stack:** Go 1.22, Docker CLI sandbox, Gin API, GORM AutoMigrate, existing `go test` suite.

---

## Current State

The current sandbox is intentionally simple:

- `internal/sandbox/docker.go` passes Docker `--memory <limit>` during run, so the kernel/container runtime can kill memory-heavy programs.
- `RunResult.MemoryKB` is always `0`.
- `RunResult` has no `MemoryExceeded` or `OOMKilled` signal.
- `internal/judge/verdict.go` has no `MLE` verdict.
- `internal/judge/engine.go` maps any non-zero runtime exit to `RE`, so Docker OOM currently appears as runtime error.
- The run container uses `--rm`, so the code cannot inspect `State.OOMKilled` after exit.

The implementation should fix those points without replacing the whole sandbox.

## Design Decisions

- Add verdict string `MLE`.
- Add `MemoryExceeded bool` to sandbox, judge, service output, DTO, and persisted judge records.
- Remove `--rm` from run containers only, inspect `State.OOMKilled`, then remove the container in `defer`.
- Keep compile containers using `--rm`; compile memory limit is infrastructure protection, not submission MLE.
- On MLE, set `MemoryKB` to at least the configured problem limit in KB when exact memory usage is unavailable.
- Judge priority: TLE first when our context timeout fires, then MLE, then RE, then WA/AC.
- Do not add cgroup polling or Docker stats in this iteration. It is higher complexity and can be added later.

## Acceptance Criteria

- A program killed by Docker OOM is reported as `MLE`.
- The API response includes `verdict: "MLE"` and `memory_exceeded: true`.
- The judge result row persists `verdict = "MLE"` and `memory_exceeded = true`.
- Existing `AC`, `WA`, `TLE`, `RE`, `CE`, `UNJUDGEABLE` behavior remains unchanged.
- `go test ./...` passes.
- Manual Docker/API smoke test confirms a high-memory C++ program gets `MLE`, not `RE`.

---

### Task 1: Add MLE Verdict and Judge-Level Signal

**Files:**
- Modify: `internal/judge/verdict.go`
- Modify: `internal/sandbox/types.go`
- Modify: `internal/judge/types.go`
- Modify: `internal/sandbox/mock.go`
- Modify: `internal/judge/engine.go`
- Test: `internal/judge/engine_test.go`

**Step 1: Write the failing judge test**

Add this test to `internal/judge/engine_test.go`:

```go
func TestJudgeMemoryLimitExceeded(t *testing.T) {
	engine := NewEngine(sandbox.NewMockSandbox())

	result, err := engine.Judge(context.Background(), Request{
		Problem: &model.Problem{
			TimeLimitMS:   1000,
			MemoryLimitMB: 64,
		},
		TestCases: []model.TestCase{
			{Input: "1", ExpectedOutput: "1"},
		},
		Language:   model.LanguageCPP17,
		SourceCode: "MOCK_MLE",
	})
	if err != nil {
		t.Fatalf("judge returned error: %v", err)
	}

	if result.Verdict != VerdictMemoryLimitExceeded {
		t.Fatalf("expected verdict %s, got %s", VerdictMemoryLimitExceeded, result.Verdict)
	}
	if !result.MemoryExceeded || result.ExecStage != "run" {
		t.Fatalf("expected memory exceeded observability fields, got %+v", result)
	}
	if result.MemoryKB < 64*1024 {
		t.Fatalf("expected memory usage to be at least limit, got %dKB", result.MemoryKB)
	}
	if len(result.TestCaseResults) != 1 || result.TestCaseResults[0].Verdict != VerdictMemoryLimitExceeded {
		t.Fatalf("expected testcase MLE summary, got %+v", result.TestCaseResults)
	}
	if !result.TestCaseResults[0].MemoryExceeded {
		t.Fatalf("expected testcase memory exceeded flag, got %+v", result.TestCaseResults[0])
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/judge -run TestJudgeMemoryLimitExceeded -count=1
```

Expected: FAIL because `VerdictMemoryLimitExceeded`, `MemoryExceeded`, and `MOCK_MLE` do not exist.

**Step 3: Add the minimal types and mock behavior**

In `internal/judge/verdict.go`, add:

```go
VerdictMemoryLimitExceeded = "MLE"
```

In `internal/sandbox/types.go`, extend `RunResult`:

```go
MemoryExceeded bool
```

In `internal/judge/types.go`, extend `Result` and `TestCaseResult`:

```go
MemoryExceeded bool
```

In `internal/sandbox/mock.go`, add marker handling:

```go
const mockMemoryLimitExceededMarker = "MOCK_MLE"
```

Then add a branch before runtime error:

```go
case strings.Contains(sourceCode, mockMemoryLimitExceededMarker):
	return RunResult{
		RuntimeMS:       1,
		MemoryKB:        req.MemoryLimitMB * 1024,
		ExitCode:        137,
		MemoryExceeded: true,
		ErrorMessage:    "memory limit exceeded",
	}, nil
```

**Step 4: Teach judge to map memory exceeded to MLE**

In `internal/judge/engine.go`, after copying common run fields, copy the flag:

```go
result.MemoryExceeded = runResult.MemoryExceeded
```

When building `caseResult`, include:

```go
MemoryExceeded: runResult.MemoryExceeded,
```

Add a switch branch after timeout and before runtime error:

```go
case runResult.MemoryExceeded:
	result.Verdict = VerdictMemoryLimitExceeded
	caseResult.Verdict = VerdictMemoryLimitExceeded
	result.TestCaseResults = append(result.TestCaseResults, caseResult)
	result.ErrorMessage = pickErrorMessage("memory limit exceeded", runResult.ErrorMessage, runResult.Stderr)
	return result, nil
```

**Step 5: Run judge tests**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/judge -count=1
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/judge/verdict.go internal/sandbox/types.go internal/judge/types.go internal/sandbox/mock.go internal/judge/engine.go internal/judge/engine_test.go
git commit -m "feat: add MLE verdict to judge"
```

---

### Task 2: Detect Docker OOMKilled in the Real Sandbox

**Files:**
- Modify: `internal/sandbox/docker.go`
- Test: `internal/sandbox/docker_test.go`

**Step 1: Write failing helper tests**

Add tests to `internal/sandbox/docker_test.go`:

```go
func TestParseDockerOOMKilled(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		want bool
	}{
		{name: "true", raw: "true\n", want: true},
		{name: "false", raw: "false\n", want: false},
		{name: "empty", raw: "", want: false},
		{name: "garbage", raw: "not-json", want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseDockerBool(tc.raw); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestDockerSandboxRunMemoryKBForOOMFallsBackToLimit(t *testing.T) {
	result := runResultForOOM("stderr", 137, 64)
	if !result.MemoryExceeded {
		t.Fatalf("expected memory exceeded result, got %+v", result)
	}
	if result.MemoryKB != 64*1024 {
		t.Fatalf("expected memory fallback to limit, got %d", result.MemoryKB)
	}
	if result.RuntimeError {
		t.Fatalf("MLE should not also be marked runtime error: %+v", result)
	}
}
```

**Step 2: Run test to verify it fails**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/sandbox -run 'TestParseDockerOOMKilled|TestDockerSandboxRunMemoryKBForOOMFallsBackToLimit' -count=1
```

Expected: FAIL because helpers do not exist.

**Step 3: Add inspect helper and result helper**

In `internal/sandbox/docker.go`, add:

```go
func parseDockerBool(raw string) bool {
	return strings.TrimSpace(raw) == "true"
}

func runResultForOOM(stderr string, exitCode int, memoryLimitMB int) RunResult {
	if memoryLimitMB <= 0 {
		memoryLimitMB = 256
	}
	return RunResult{
		Stderr:          stderr,
		ExitCode:        exitCode,
		MemoryKB:        memoryLimitMB * 1024,
		MemoryExceeded: true,
		ErrorMessage:    "memory limit exceeded",
	}
}

func (s *DockerSandbox) containerOOMKilled(ctx context.Context, containerName string) bool {
	inspectCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	cmd := exec.CommandContext(
		inspectCtx,
		"docker",
		"inspect",
		"--format",
		"{{.State.OOMKilled}}",
		containerName,
	)
	output, err := cmd.Output()
	if err != nil {
		if s.logger != nil {
			s.logger.Warn("inspect run container oom state failed", "container_name", containerName, "error", err)
		}
		return false
	}
	return parseDockerBool(string(output))
}
```

**Step 4: Change run container lifecycle**

In `DockerSandbox.Run`, remove `--rm` from the run command. Add cleanup immediately after `containerName` is created:

```go
defer s.forceRemoveContainer(containerName)
```

Keep compile containers unchanged.

After `runDockerCommand` returns and before the timed-out branch, inspect OOM:

```go
memoryExceeded := false
if !timedOut {
	memoryExceeded = s.containerOOMKilled(context.Background(), containerName)
}
```

Then add this branch before the Docker infra exit and generic runtime result:

```go
if memoryExceeded {
	result := runResultForOOM(stderr, exitCode, req.MemoryLimitMB)
	result.Stdout = stdout
	result.RuntimeMS = runtimeMS
	return result, nil
}
```

Important: keep `timedOut` behavior first. If our timeout kills the container, report `TLE`, not `MLE`.

**Step 5: Run sandbox tests**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/sandbox -count=1
```

Expected: PASS.

**Step 6: Commit**

```bash
git add internal/sandbox/docker.go internal/sandbox/docker_test.go
git commit -m "feat: detect docker oom in sandbox"
```

---

### Task 3: Persist and Expose MemoryExceeded

**Files:**
- Modify: `internal/model/judge_result.go`
- Modify: `internal/model/submission_test_case_result.go`
- Modify: `internal/service/judge_submission.go`
- Modify: `internal/handler/dto/judge_submission.go`
- Modify: `internal/handler/dto/submission_query.go`
- Modify: `internal/service/submission_query.go`
- Test: `internal/service/judge_submission_test.go`
- Test: `internal/service/submission_query_test.go`

**Step 1: Write failing service persistence test**

In `internal/service/judge_submission_test.go`, add or extend the MLE path using a fake engine result:

```go
func TestJudgeSubmissionPersistsMemoryExceeded(t *testing.T) {
	problemRepo := fakeProblemRepository{
		problem: &model.Problem{
			BaseModel:     model.BaseModel{ID: 1},
			TimeLimitMS:   1000,
			MemoryLimitMB: 64,
			TestCases:     []model.TestCase{{BaseModel: model.BaseModel{ID: 10}, Input: "1", ExpectedOutput: "1"}},
		},
	}
	submissionRepo := &fakeSubmissionRepository{}
	engine := fakeJudgeEngine{
		result: judge.Result{
			Verdict:         judge.VerdictMemoryLimitExceeded,
			RuntimeMS:       1,
			MemoryKB:        64 * 1024,
			MemoryExceeded:  true,
			TotalCount:      1,
			ExecStage:       "run",
			ErrorMessage:    "memory limit exceeded",
			TestCaseResults: []judge.TestCaseResult{{TestCaseID: 10, CaseIndex: 1, Verdict: judge.VerdictMemoryLimitExceeded, MemoryExceeded: true}},
		},
	}
	service := NewJudgeSubmissionService(problemRepo, submissionRepo, engine)

	output, err := service.Submit(context.Background(), JudgeSubmissionInput{
		ProblemID:  1,
		SourceCode: "MOCK_MLE",
		Language:   model.LanguageCPP17,
	})
	if err != nil {
		t.Fatalf("submit returned error: %v", err)
	}

	if output.Verdict != judge.VerdictMemoryLimitExceeded || !output.MemoryExceeded {
		t.Fatalf("expected MLE output, got %+v", output)
	}
	if submissionRepo.judgeResult == nil || !submissionRepo.judgeResult.MemoryExceeded {
		t.Fatalf("expected persisted judge result memory flag, got %+v", submissionRepo.judgeResult)
	}
	if len(submissionRepo.testCaseResults) != 1 || !submissionRepo.testCaseResults[0].MemoryExceeded {
		t.Fatalf("expected persisted testcase memory flag, got %+v", submissionRepo.testCaseResults)
	}
}
```

Adapt fake names to the existing test helpers if they differ.

**Step 2: Run test to verify it fails**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/service -run TestJudgeSubmissionPersistsMemoryExceeded -count=1
```

Expected: FAIL because model/output fields do not exist.

**Step 3: Add model fields**

In `internal/model/judge_result.go`, add:

```go
MemoryExceeded bool `gorm:"column:memory_exceeded;not null;default:false" json:"memory_exceeded"`
```

In `internal/model/submission_test_case_result.go`, add:

```go
MemoryExceeded bool `gorm:"column:memory_exceeded;not null;default:false" json:"memory_exceeded"`
```

GORM AutoMigrate should add these columns on startup.

**Step 4: Add service output fields and persistence mapping**

In `internal/service/judge_submission.go`, add `MemoryExceeded bool` to:

- `JudgeSubmissionOutput`
- `JudgeSubmissionCaseFeedback`

When creating `model.JudgeResult`, set:

```go
MemoryExceeded: judgeResult.MemoryExceeded,
```

When creating `model.SubmissionTestCaseResult`, set:

```go
MemoryExceeded: item.MemoryExceeded,
```

When creating `JudgeSubmissionOutput` and `JudgeSubmissionCaseFeedback`, copy the same flag.

**Step 5: Add DTO and query output fields**

In `internal/handler/dto/judge_submission.go`, add:

```go
MemoryExceeded bool `json:"memory_exceeded"`
```

to the submission response and case feedback DTOs.

In `internal/handler/dto/submission_query.go` and `internal/service/submission_query.go`, add the same field to detail/list result shapes where `TimedOut` already exists.

**Step 6: Run affected tests**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/service ./internal/handler ./internal/model -count=1
```

Expected: PASS.

**Step 7: Commit**

```bash
git add internal/model/judge_result.go internal/model/submission_test_case_result.go internal/service/judge_submission.go internal/handler/dto/judge_submission.go internal/handler/dto/submission_query.go internal/service/submission_query.go internal/service/judge_submission_test.go internal/service/submission_query_test.go
git commit -m "feat: persist memory exceeded judge results"
```

---

### Task 4: Add an End-to-End Docker Sandbox Test Hook

**Files:**
- Create: `internal/sandbox/docker_integration_test.go`

**Step 1: Write skipped-by-default integration test**

Create `internal/sandbox/docker_integration_test.go`:

```go
package sandbox

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"ai-for-oj/internal/config"
	"ai-for-oj/internal/model"
)

func TestDockerSandboxDetectsOOMKilledIntegration(t *testing.T) {
	if os.Getenv("AI_FOR_OJ_DOCKER_INTEGRATION") != "1" {
		t.Skip("set AI_FOR_OJ_DOCKER_INTEGRATION=1 to run docker integration test")
	}

	s, err := NewDockerSandbox(config.SandboxConfig{
		WorkDir:          t.TempDir(),
		DockerImage:      "gcc:13",
		CompileTimeout:   10 * time.Second,
		RunTimeoutBuffer: 500 * time.Millisecond,
		CompileMemoryMB:  512,
	}, slog.Default())
	if err != nil {
		t.Fatalf("new docker sandbox: %v", err)
	}

	source := `#include <vector>
int main() {
  std::vector<char> data;
  while (true) data.resize(data.size() + 1024 * 1024, 1);
}`

	compileResult, err := s.Compile(context.Background(), CompileRequest{
		Language:   model.LanguageCPP17,
		SourceCode: source,
	})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if !compileResult.Success {
		t.Fatalf("compile failed: %+v", compileResult)
	}
	defer func() {
		_ = s.Cleanup(context.Background(), compileResult.ArtifactID)
	}()

	runResult, err := s.Run(context.Background(), RunRequest{
		Language:      model.LanguageCPP17,
		ArtifactID:    compileResult.ArtifactID,
		TimeLimitMS:   5000,
		MemoryLimitMB: 32,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !runResult.MemoryExceeded {
		t.Fatalf("expected memory exceeded result, got %+v", runResult)
	}
}
```

**Step 2: Run normal tests and confirm it skips**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/sandbox -count=1
```

Expected: PASS with integration test skipped.

**Step 3: Run integration test manually**

Run:

```bash
AI_FOR_OJ_DOCKER_INTEGRATION=1 GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/sandbox -run TestDockerSandboxDetectsOOMKilledIntegration -count=1 -v
```

Expected: PASS if local Docker can run `gcc:13`.

If it fails because `gcc:13` is missing, run:

```bash
docker pull gcc:13
```

Then rerun the integration test.

**Step 4: Commit**

```bash
git add internal/sandbox/docker_integration_test.go
git commit -m "test: cover docker sandbox oom detection"
```

---

### Task 5: API Smoke Test Against the Running App

**Files:**
- No code changes expected.

**Step 1: Rebuild the app container**

Run:

```bash
docker compose up -d --build app
```

Expected: app image builds and `ai-for-oj-app` starts.

**Step 2: Create a temporary low-memory problem**

Run:

```bash
curl --noproxy '*' -sS -X POST http://127.0.0.1:8080/api/v1/problems \
  -H 'Content-Type: application/json' \
  --data '{
    "title":"mle smoke test",
    "description":"temporary problem for memory limit verification",
    "input_spec":"none",
    "output_spec":"none",
    "samples":"[]",
    "time_limit_ms":5000,
    "memory_limit_mb":32,
    "difficulty":"easy",
    "tags":"smoke"
  }'
```

Expected: JSON response with an `id`. Save it as `PROBLEM_ID`.

**Step 3: Add one testcase**

Run:

```bash
curl --noproxy '*' -sS -X POST http://127.0.0.1:8080/api/v1/problems/PROBLEM_ID/testcases \
  -H 'Content-Type: application/json' \
  --data '{"input":"","expected_output":"","is_sample":false}'
```

Expected: `201 Created` JSON response with testcase id.

**Step 4: Submit memory-heavy C++**

Run:

```bash
curl --noproxy '*' -sS -X POST http://127.0.0.1:8080/api/v1/submissions/judge \
  -H 'Content-Type: application/json' \
  --data '{
    "problem_id": PROBLEM_ID,
    "language": "cpp17",
    "source_type": "human",
    "source_code": "#include <vector>\nint main(){std::vector<char> v; while(true) v.resize(v.size()+1024*1024, 1);}"
  }'
```

Expected response contains:

```json
{
  "verdict": "MLE",
  "memory_exceeded": true
}
```

It must not return `RE`.

**Step 5: Delete the temporary problem**

Run:

```bash
curl --noproxy '*' -sS -i -X DELETE http://127.0.0.1:8080/api/v1/problems/PROBLEM_ID
```

Expected: `204 No Content`.

---

### Task 6: Full Verification and Documentation

**Files:**
- Modify: `README.md` or `docs/` only if the project already documents verdicts/sandbox behavior.

**Step 1: Run full backend test suite**

Run:

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./...
```

Expected: PASS.

**Step 2: Check for sandbox container leaks**

Run:

```bash
docker ps -a --filter 'name=ai-for-oj-run-' --format '{{.Names}} {{.Status}}'
```

Expected: no stale run containers after normal tests and manual smoke.

If stale containers exist, fix cleanup before continuing.

**Step 3: Document behavior**

If README has a verdict/sandbox section, add a concise note:

```markdown
- `MLE`: the Docker sandbox detected that the run container was OOM-killed by the configured problem memory limit.
- Memory usage is currently reported as the problem limit when Docker OOMs; precise peak RSS accounting is not implemented yet.
```

**Step 4: Commit**

```bash
git add README.md docs
git commit -m "docs: describe sandbox memory limit behavior"
```

Skip this commit if no docs were changed.

---

## Risks and Follow-Ups

- Docker `OOMKilled` is the reliable signal for cgroup memory kills, but it requires keeping the container around briefly for inspect. Cleanup must be tested carefully.
- Exact peak memory is not solved here. Docker does not provide a simple post-exit peak RSS through `docker inspect`.
- If a program catches allocation failure and exits normally, Docker will not mark OOMKilled. That is not MLE; it is normal program behavior.
- If context timeout and OOM happen at nearly the same time, this plan reports TLE when our timeout path killed the container first.
- Future enhancement: add cgroup-based or polling-based peak memory measurement and expose real `MemoryKB` for all submissions.

## Final Review Checklist

- [ ] `MLE` appears in judge verdict constants.
- [ ] Mock sandbox can simulate MLE.
- [ ] Docker sandbox inspects `State.OOMKilled` before removing run containers.
- [ ] Run containers are still cleaned up after success, RE, MLE, and TLE.
- [ ] Judge maps memory exceeded to `MLE`, not `RE`.
- [ ] Service persists `memory_exceeded`.
- [ ] API responses expose `memory_exceeded`.
- [ ] `go test ./...` passes.
- [ ] Manual Docker/API MLE smoke test passes.
