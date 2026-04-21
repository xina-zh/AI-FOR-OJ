# Adaptive Repair Agent Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a stateful `adaptive_repair_v1` solve agent that significantly improves single-model AC rate by handling `WA`, `RE`, and `TLE` through verdict-specific repair flows and attempt-level observability.

**Architecture:** Move solve orchestration out of `AISolveService` and into a new agent coordinator that owns the full attempt loop: initial generation, judge feedback classification, repair planning, and retry. Add attempt-level persistence so every LLM call and judge result is recorded, then expose aggregated attempt metadata through AI solve and experiment APIs for analysis.

**Tech Stack:** Go, GORM, existing repository/service layering, current LLM client, current judge submission flow, Go testing package.

---

## Scope

- Keep existing agents `direct_codegen`, `direct_codegen_repair`, and `analyze_then_codegen` as baselines.
- Add one new primary agent: `adaptive_repair_v1`.
- Optimize for `WA`, `RE`, and `TLE`.
- Do not spend time on `CE` / markdown-format problems beyond keeping current extraction behavior intact.
- Do not add multi-model cascade, sampling/voting, external tools, or problem-type classifiers in this phase.

## Design Summary

The new agent should own a full solve state machine instead of returning a single `SolveOutput`. The coordinator will start with one codegen attempt, submit it to judge, classify the failure, select a repair strategy, and continue until either `AC`, retry budget exhausted, or a terminal error occurs. Every attempt is persisted with the exact prompt, raw response, extracted code, failure classification, and judge details. `AISolveRun` remains the summary row; `AISolveAttempt` becomes the detailed execution log.

Suggested stage names:

- `initial_codegen`
- `wa_analysis_repair`
- `re_safety_repair`
- `tle_complexity_rewrite`
- `fallback_rewrite`

Suggested retry policy:

- Max 4 judged attempts total for `adaptive_repair_v1`
- `WA`: allow up to 2 targeted repairs
- `RE`: allow 1 safety-focused repair, then optional full rewrite
- `TLE`: allow 1 complexity rewrite

## Files To Create Or Modify

**Create:**

- `internal/agent/adaptive_repair.go`
- `internal/agent/coordinator.go`
- `internal/agent/failure_classifier.go`
- `internal/agent/repair_planner.go`
- `internal/agent/executor.go`
- `internal/model/ai_solve_attempt.go`
- `internal/repository/ai_solve_attempt_repository.go`
- `internal/prompt/repair_wa.go`
- `internal/prompt/repair_re.go`
- `internal/prompt/repair_tle.go`
- `internal/agent/adaptive_repair_test.go`
- `internal/service/ai_solve_attempt_test.go` or extend `internal/service/ai_solve_test.go`

**Modify:**

- `internal/agent/solve.go`
- `internal/service/ai_solve.go`
- `internal/model/ai_solve_run.go`
- `internal/model/schema.go`
- `internal/bootstrap/app.go`
- `internal/handler/dto/ai_solve.go`
- `internal/handler/dto/experiment.go`
- `internal/service/experiment.go`
- `internal/repository/experiment_repository.go` if additional preload fields are needed
- `internal/service/experiment_test.go`
- `internal/service/ai_solve_test.go`

## Task 1: Add Attempt-Level Domain Model

**Files:**

- Create: `internal/model/ai_solve_attempt.go`
- Modify: `internal/model/ai_solve_run.go`
- Modify: `internal/model/schema.go`

**Step 1: Write the failing test**

Add a schema test or extend an existing model/schema test to assert that `AllModels()` includes a new `AISolveAttempt` model and that an `AISolveRun` can reference attempts.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/model ./internal/bootstrap`
Expected: FAIL because `AISolveAttempt` does not exist in the schema list yet.

**Step 3: Write minimal implementation**

Add `AISolveAttempt` with fields:

- `AISolveRunID uint`
- `AttemptNo int`
- `Stage string`
- `FailureType string`
- `RepairReason string`
- `StrategyPath string`
- `PromptPreview string`
- `RawResponse string`
- `ExtractedCode string`
- `JudgeVerdict string`
- `JudgePassedCount int`
- `JudgeTotalCount int`
- `TimedOut bool`
- `CompileStderr string`
- `RunStderr string`
- `RunStdout string`
- `ErrorMessage string`
- `TokenInput int64`
- `TokenOutput int64`
- `LLMLatencyMS int`

Extend `AISolveRun` with summary fields:

- `AttemptCount int`
- `FailureType string`
- `StrategyPath string`

Add relation:

- `Attempts []AISolveAttempt`

Register the model in `AllModels()`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/model ./internal/bootstrap`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/model/ai_solve_attempt.go internal/model/ai_solve_run.go internal/model/schema.go
git commit -m "feat: add ai solve attempt model"
```

## Task 2: Add Attempt Repository Support

**Files:**

- Create: `internal/repository/ai_solve_attempt_repository.go`
- Modify: `internal/bootstrap/app.go`

**Step 1: Write the failing test**

Add repository tests or service fake expectations asserting attempts can be created and queried by run ID.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/repository ./internal/service`
Expected: FAIL because no attempt repository is wired.

**Step 3: Write minimal implementation**

Create interface:

- `Create(ctx, attempt *model.AISolveAttempt) error`
- `ListByRunID(ctx, runID uint) ([]model.AISolveAttempt, error)`

Implement a GORM repository and wire it in `internal/bootstrap/app.go` so services can persist attempts.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/repository ./internal/service`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/repository/ai_solve_attempt_repository.go internal/bootstrap/app.go
git commit -m "feat: add ai solve attempt repository"
```

## Task 3: Introduce Adaptive Agent Entry Point

**Files:**

- Modify: `internal/agent/solve.go`
- Create: `internal/agent/adaptive_repair.go`

**Step 1: Write the failing test**

Add tests asserting:

- `ResolveSolveAgentName("adaptive_repair_v1")` succeeds
- `ResolveSolveStrategy("adaptive_repair_v1")` returns the new strategy
- `SupportsSelfRepair` no longer drives adaptive repair behavior

**Step 2: Run test to verify it fails**

Run: `go test ./internal/agent`
Expected: FAIL because the new agent name is unknown.

**Step 3: Write minimal implementation**

Add constant `AdaptiveRepairV1AgentName = "adaptive_repair_v1"` and wire it into resolver functions. Keep legacy agents intact.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/agent`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/solve.go internal/agent/adaptive_repair.go
git commit -m "feat: register adaptive repair agent"
```

## Task 4: Extract LLM Execution Utilities

**Files:**

- Create: `internal/agent/executor.go`
- Modify: `internal/agent/solve.go`
- Test: `internal/agent/adaptive_repair_test.go`

**Step 1: Write the failing test**

Add tests covering a shared helper that runs one LLM prompt and returns prompt preview, raw response, token counts, latency, and effective model selection.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/agent`
Expected: FAIL because no reusable executor exists.

**Step 3: Write minimal implementation**

Move `generateOnce`, `elapsedMS`, and effective-model logic into `executor.go`, and expose a small helper used by both legacy agents and the new coordinator.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/agent`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/executor.go internal/agent/solve.go internal/agent/adaptive_repair_test.go
git commit -m "refactor: share llm execution helpers"
```

## Task 5: Implement Failure Classification

**Files:**

- Create: `internal/agent/failure_classifier.go`
- Test: `internal/agent/adaptive_repair_test.go`

**Step 1: Write the failing test**

Add table-driven tests for judge outputs:

- `WA` -> `wrong_answer`
- `RE` -> `runtime_error`
- `TLE` or `TimedOut` -> `time_limit`
- empty/other -> `unknown`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/agent -run TestClassifyFailure`
Expected: FAIL because classifier does not exist.

**Step 3: Write minimal implementation**

Create a classifier that reads verdict, timeout flag, stderr, passed counts, and execution stage. Keep it deterministic and minimal; no speculative NLP.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/agent -run TestClassifyFailure`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/failure_classifier.go internal/agent/adaptive_repair_test.go
git commit -m "feat: classify ai solve failures"
```

## Task 6: Add Verdict-Specific Repair Prompts

**Files:**

- Create: `internal/prompt/repair_wa.go`
- Create: `internal/prompt/repair_re.go`
- Create: `internal/prompt/repair_tle.go`
- Modify: `internal/prompt/solve.go`
- Test: `internal/prompt/solve_test.go`

**Step 1: Write the failing test**

Add prompt tests asserting each repair builder includes the right constraints:

- `WA`: requires edge cases and algorithm correction
- `RE`: requires safety checks and runtime robustness
- `TLE`: requires complexity comparison and algorithm rewrite

**Step 2: Run test to verify it fails**

Run: `go test ./internal/prompt`
Expected: FAIL because the repair prompt builders are missing.

**Step 3: Write minimal implementation**

Implement three prompt builders:

- `BuildWARepairPrompt(...)`
- `BuildRERepairPrompt(...)`
- `BuildTLERepairPrompt(...)`

Prompt requirements:

- `WA`: ask for mistake diagnosis, at least 3 edge cases, corrected algorithm, then full code
- `RE`: ask for root cause in implementation safety, then robust full code
- `TLE`: ask for old vs new complexity and a more efficient rewrite

**Step 4: Run test to verify it passes**

Run: `go test ./internal/prompt`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/prompt/repair_wa.go internal/prompt/repair_re.go internal/prompt/repair_tle.go internal/prompt/solve.go internal/prompt/solve_test.go
git commit -m "feat: add verdict specific repair prompts"
```

## Task 7: Implement Repair Planner

**Files:**

- Create: `internal/agent/repair_planner.go`
- Test: `internal/agent/adaptive_repair_test.go`

**Step 1: Write the failing test**

Add tests for planner decisions:

- `wrong_answer` after initial attempt -> `wa_analysis_repair`
- `runtime_error` -> `re_safety_repair`
- `time_limit` -> `tle_complexity_rewrite`
- repeated same failure beyond budget -> `fallback_rewrite` or stop

**Step 2: Run test to verify it fails**

Run: `go test ./internal/agent -run TestRepairPlanner`
Expected: FAIL because no planner exists.

**Step 3: Write minimal implementation**

Implement a small planner using:

- current attempt count
- last failure type
- previous stages
- max budget

Keep rules explicit; avoid heuristic sprawl.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/agent -run TestRepairPlanner`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/repair_planner.go internal/agent/adaptive_repair_test.go
git commit -m "feat: add adaptive repair planner"
```

## Task 8: Build Agent Coordinator

**Files:**

- Create: `internal/agent/coordinator.go`
- Modify: `internal/agent/adaptive_repair.go`
- Test: `internal/agent/adaptive_repair_test.go`

**Step 1: Write the failing test**

Add coordinator tests using fake LLM responses and fake judge outcomes to verify:

- initial AC stops after one attempt
- `WA` triggers `wa_analysis_repair`
- `RE` triggers `re_safety_repair`
- `TLE` triggers `tle_complexity_rewrite`
- attempt count and strategy path accumulate correctly

**Step 2: Run test to verify it fails**

Run: `go test ./internal/agent`
Expected: FAIL because the coordinator loop does not exist.

**Step 3: Write minimal implementation**

Implement a coordinator that:

1. Builds initial solve prompt
2. Calls LLM
3. Extracts code
4. Returns attempt output metadata to caller
5. Accepts judge feedback for classification and next-step planning

Design note: if full judge integration inside agent is too invasive for the existing interfaces, define a coordinator contract that the service drives but the coordinator owns all repair decisions and stage transitions. The key requirement is that the repair loop logic is no longer open-coded in `AISolveService`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/agent`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/coordinator.go internal/agent/adaptive_repair.go internal/agent/adaptive_repair_test.go
git commit -m "feat: implement adaptive repair coordinator"
```

## Task 9: Refactor AISolve Service To Use Coordinator

**Files:**

- Modify: `internal/service/ai_solve.go`
- Modify: `internal/service/ai_solve_test.go`

**Step 1: Write the failing test**

Extend service tests to assert:

- adaptive agent creates multiple attempts on `WA/RE/TLE`
- final run stores `AttemptCount`, `FailureType`, and `StrategyPath`
- legacy agents still work unchanged

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestAISolve`
Expected: FAIL because service still contains the old repair loop and no attempt persistence.

**Step 3: Write minimal implementation**

Refactor `AISolveService` so it:

- loads problem and creates `AISolveRun`
- invokes the strategy/coordinator
- persists each attempt through the new repository
- submits judged code through existing `JudgeSubmitter`
- updates only final run summary fields

Remove `SupportsSelfRepair` branching from service for the new adaptive flow.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run TestAISolve`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/ai_solve.go internal/service/ai_solve_test.go
git commit -m "refactor: move adaptive solve loop into agent flow"
```

## Task 10: Persist And Expose Attempt Details In Run Queries

**Files:**

- Modify: `internal/repository/ai_solve_run_repository.go`
- Modify: `internal/handler/dto/ai_solve.go`
- Modify: `internal/handler/ai.go`
- Modify: `internal/service/ai_solve.go`

**Step 1: Write the failing test**

Add API/service tests asserting `GET /api/v1/ai/solve-runs/:id` includes:

- summary attempt metadata
- attempt list with stage, verdict, failure type, repair reason

**Step 2: Run test to verify it fails**

Run: `go test ./internal/handler ./internal/service`
Expected: FAIL because attempts are not preloaded or serialized.

**Step 3: Write minimal implementation**

Preload `Attempts` when fetching runs. Add DTOs for attempt output. Expose summary fields on `AISolveRunResponse`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/handler ./internal/service`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/repository/ai_solve_run_repository.go internal/handler/dto/ai_solve.go internal/handler/ai.go internal/service/ai_solve.go
git commit -m "feat: expose ai solve attempt details"
```

## Task 11: Extend Experiment Output For Agent Analysis

**Files:**

- Modify: `internal/service/experiment.go`
- Modify: `internal/handler/dto/experiment.go`
- Modify: `internal/service/experiment_test.go`

**Step 1: Write the failing test**

Add experiment tests asserting output includes:

- `attempt_count`
- `strategy_path`
- optionally final `failure_type`

and that cost/AC summaries still work with the new fields present.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestExperiment`
Expected: FAIL because the new summary fields are absent.

**Step 3: Write minimal implementation**

Populate experiment runs and summary DTOs with attempt-level aggregate metadata from linked `AISolveRun`.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run TestExperiment`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/service/experiment.go internal/handler/dto/experiment.go internal/service/experiment_test.go
git commit -m "feat: add adaptive agent metadata to experiments"
```

## Task 12: Build Regression Tests For WA / RE / TLE Paths

**Files:**

- Modify: `internal/agent/adaptive_repair_test.go`
- Modify: `internal/service/ai_solve_test.go`

**Step 1: Write the failing test**

Add end-to-end-ish tests with fake LLM and fake judge:

- initial wrong logic -> `WA` -> analysis repair -> `AC`
- runtime crash -> `RE` -> safety repair -> `AC`
- quadratic solution -> `TLE` -> complexity rewrite -> `AC`

**Step 2: Run test to verify it fails**

Run: `go test ./internal/agent ./internal/service`
Expected: FAIL because at least one target path is not handled correctly yet.

**Step 3: Write minimal implementation**

Fill missing coordinator or prompt details until the test scenarios pass. Do not add unrelated heuristics.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/agent ./internal/service`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/agent/adaptive_repair_test.go internal/service/ai_solve_test.go
git commit -m "test: cover adaptive repair verdict flows"
```

## Task 13: Run Full Verification

**Files:**

- No code changes required unless a failure is found

**Step 1: Run focused tests**

Run:

```bash
go test ./internal/agent ./internal/prompt ./internal/service ./internal/repository ./internal/handler
```

Expected: PASS

**Step 2: Run broader regression**

Run:

```bash
go test ./...
```

Expected: PASS

**Step 3: Smoke-check API behavior if a local server workflow already exists**

Run existing local AI solve / experiment flows against a fixed problem set and compare:

- old baseline agent
- `adaptive_repair_v1`

Metrics to compare:

- AC rate
- average attempts per solved problem
- `WA -> AC` conversion count
- `RE -> AC` conversion count
- `TLE -> AC` conversion count

**Step 4: Commit final integration fixes**

```bash
git add .
git commit -m "feat: add adaptive repair solve agent"
```

## Acceptance Criteria

- `adaptive_repair_v1` is selectable anywhere agent names are currently accepted.
- `AISolveService` no longer hardcodes the old generic repair loop for the adaptive path.
- Every adaptive solve records attempt-level rows with stage, failure type, verdict, code, and token/latency metadata.
- `WA`, `RE`, and `TLE` each trigger different repair prompts and different planner stages.
- AI solve run queries expose attempt history and run-level summary metadata.
- Experiment output exposes enough metadata to compare strategy paths and attempt counts.
- Existing baseline agents still behave as before.
- All relevant tests pass.

## Non-Goals

- No multi-model fallback
- No tool calling
- No candidate voting
- No problem taxonomy engine
- No prompt-template explosion beyond the 3 targeted repair prompts

## Risks And Mitigations

- Risk: The agent/service split becomes awkward because judge submission currently lives in service.
  Mitigation: Keep judge submission in service if needed, but move all decision-making and planner state into the coordinator contract.

- Risk: Attempt payloads become too large in API responses.
  Mitigation: Truncate previews for prompt/raw-response fields, matching current run preview patterns.

- Risk: New prompts increase latency significantly without improving AC.
  Mitigation: Attempt-level observability is mandatory so low-yield stages can be pruned after the first benchmark pass.

- Risk: `unknown` failures degrade into useless retries.
  Mitigation: Allow at most one `fallback_rewrite` stage, then stop.

## Execution Notes

- Prefer TDD for every task.
- Keep commits small and scoped to one task.
- Do not refactor unrelated repository or DTO code while touching the solve path.
- Preserve backward compatibility for existing API fields.

Plan complete and saved to `docs/plans/2026-04-17-adaptive-repair-agent.md`. Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
