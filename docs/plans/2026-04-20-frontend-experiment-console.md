# 前端实验控制台实施计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 做一个模块化前端控制台，用来运行 AI OJ 模型实验。前端要替代手写 curl / 命令行，支持变量控制、单题 solve、批量 experiment、compare、repeat、token / 成本视图、trace 回放和结果下钻。

**架构：** 新增独立的 `web/` 前端项目，使用 React + TypeScript，通过类型化 API 模块调用现有 Gin 后端。每个功能放在 `web/src/features/*` 下的独立目录中，公共 UI、API、格式化工具单独拆分。后端只补前端真正需要的少量支撑接口：实验选项元数据、实验历史列表、experiment run trace。第一版保持同步执行，因为当前后端实验接口本身就是同步返回。

**技术栈：** Go / Gin / GORM 后端，React，TypeScript，Vite，React Router，TanStack Query，Vitest，React Testing Library，Playwright，Recharts 或轻量图表组件，CSS modules 或按组件拆分的普通 CSS。

---

## 当前上下文

后端已经暴露了核心实验链路：

- `GET /health`
- `GET /api/v1/problems`
- `POST /api/v1/problems`
- `GET /api/v1/problems/:id`
- `POST /api/v1/problems/:id/testcases`
- `GET /api/v1/problems/:id/testcases`
- `POST /api/v1/submissions/judge`
- `GET /api/v1/submissions`
- `GET /api/v1/submissions/:id`
- `GET /api/v1/submissions/stats/problems`
- `POST /api/v1/ai/solve`
- `GET /api/v1/ai/solve-runs/:id`
- `POST /api/v1/experiments/run`
- `GET /api/v1/experiments/:id`
- `POST /api/v1/experiments/compare`
- `GET /api/v1/experiments/compare/:id`
- `POST /api/v1/experiments/repeat`
- `GET /api/v1/experiments/repeat/:id`

现有变量和结果字段：

- 请求变量：`model`、`prompt_name`、`agent_name`、`problem_ids`、`repeat_count`。
- 已有 prompt：`default`、`cpp17_minimal`、`strict_cpp17`。
- 已有 agent：`direct_codegen`、`direct_codegen_repair`、`analyze_then_codegen`。
- token 字段：`token_input`、`token_output`、`cost_summary`、`cost_comparison`。
- latency 字段：`llm_latency_ms`、`total_latency_ms`，以及 experiment 层的平均值。
- compare 字段：verdict 分布、delta 分布、成本对比、对比结论、highlighted problems、按题对比结果。
- trace 缺口：现在有 `model.TraceEvent` 模型，但还没有公开查询 trace 的接口，trace 写入链路也还没有完整接上。

## 范围

本计划要做的是一个本地可用的实验控制台：

- 用前端表单和按钮替代 curl 命令。
- 每次运行前都能选择 `model`、`prompt_name`、`agent_name`。
- 可以在浏览器里运行 single solve、batch experiment、compare 和 repeat。
- 展示 verdict、token 用量、latency、成本汇总、raw model output、extracted code、judge 详情。
- 展示 compare 结果，包括 baseline / candidate 差异和按题下钻。
- 基于已有的 `AISolveRun`、`Submission` 和后续 `TraceEvent` 数据做 trace 回放。
- 前端按功能拆目录，方便后续调试和修改。

## 不做的内容

第一版前端不做这些：

- 异步任务队列或实时流式进度。
- 多用户登录和权限。
- 完整 benchmark 套件管理。
- 按价格表换算货币成本。当前只展示 token 和 latency，等后端有价格表后再加金额。
- 在浏览器里编辑后端 LLM provider 密钥。
- 把后端实验逻辑搬到前端。

## 为了前端体验需要补的后端接口

现有 API 已经能运行实验，但前端还需要元数据和历史列表。先补下面几个小接口，前端不用硬编码太多后端常量。

### 新后端接口 1：实验选项

`GET /api/v1/meta/experiment-options`

返回：

```json
{
  "default_model": "mock-cpp17",
  "prompts": [
    {"name": "default", "label": "default"},
    {"name": "cpp17_minimal", "label": "cpp17_minimal"},
    {"name": "strict_cpp17", "label": "strict_cpp17"}
  ],
  "agents": [
    {"name": "direct_codegen", "label": "direct_codegen"},
    {"name": "direct_codegen_repair", "label": "direct_codegen_repair"},
    {"name": "analyze_then_codegen", "label": "analyze_then_codegen"}
  ]
}
```

文件：

- Create: `internal/handler/meta.go`
- Create: `internal/handler/dto/meta.go`
- Modify: `internal/runtime/router.go`
- Modify: `internal/bootstrap/app.go`
- Modify: `internal/agent/solve.go`
- Modify: `internal/prompt/solve.go`
- Test: `internal/runtime/router_test.go`

### 新后端接口 2：实验历史列表

增加列表接口，让前端不用提前知道 ID 也能打开旧实验：

- `GET /api/v1/experiments?page=1&page_size=20`
- `GET /api/v1/experiments/compare?page=1&page_size=20`
- `GET /api/v1/experiments/repeat?page=1&page_size=20`

文件：

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
- Modify: `internal/runtime/router.go`
- Test: service、handler、router 的列表响应测试。

### 新后端接口 3：Experiment Run Trace

`GET /api/v1/experiment-runs/:id/trace`

第一版可以先从现有数据合成 trace event：

- `prompt` 来自 `AISolveRun.prompt_preview`
- `llm_response` 来自 `AISolveRun.raw_response`
- `extracted_code` 来自 `AISolveRun.extracted_code`
- `judge_summary` 来自关联 submission 和 judge result
- `testcase_results` 来自关联 submission 的 testcase results
- 后续后端开始写入 `trace_events` 后，再把真实 trace rows 追加进去

文件：

- Create: `internal/handler/trace.go`
- Create: `internal/handler/dto/trace.go`
- Create: `internal/service/trace.go`
- Create: `internal/repository/trace_repository.go`
- Modify: `internal/repository/experiment_repository.go` 或新增一个专门查询 experiment run 的方法。
- Modify: `internal/runtime/router.go`
- Modify: `internal/bootstrap/app.go`
- Test: `internal/service/trace_test.go`、`internal/handler/trace_test.go`、router test。

## 前端文件结构

按功能拆目录。不要把页面逻辑、API 调用、格式化、图表和表单全塞到一个大文件里。

```text
web/
  package.json
  index.html
  vite.config.ts
  tsconfig.json
  src/
    main.tsx
    app/
      App.tsx
      router.tsx
      queryClient.ts
      app.css
    api/
      http.ts
      types.ts
      healthApi.ts
      metaApi.ts
      problemApi.ts
      aiApi.ts
      experimentApi.ts
      submissionApi.ts
      traceApi.ts
    components/
      layout/
        AppShell.tsx
        Sidebar.tsx
        Topbar.tsx
      ui/
        Button.tsx
        Card.tsx
        Field.tsx
        Select.tsx
        TextArea.tsx
        Table.tsx
        Tabs.tsx
        Badge.tsx
        EmptyState.tsx
        ErrorPanel.tsx
        LoadingBlock.tsx
      metrics/
        VerdictBadge.tsx
        TokenSummary.tsx
        LatencySummary.tsx
        VerdictDistribution.tsx
        CostComparison.tsx
      code/
        CodeBlock.tsx
        DiffBlock.tsx
    features/
      dashboard/
        DashboardPage.tsx
        RecentRuns.tsx
        HealthStatus.tsx
      variables/
        ExperimentVariableForm.tsx
        ProblemPicker.tsx
        ModelInput.tsx
      problems/
        ProblemsPage.tsx
        ProblemList.tsx
        ProblemDetail.tsx
        ProblemCreateForm.tsx
        TestCasePanel.tsx
      solve/
        SingleSolvePage.tsx
        SolveResultPanel.tsx
        SolveRunDetail.tsx
      experiments/
        ExperimentRunPage.tsx
        ExperimentDetailPage.tsx
        ExperimentRunForm.tsx
        ExperimentRunTable.tsx
      compare/
        ComparePage.tsx
        CompareForm.tsx
        CompareSummary.tsx
        CompareProblemTable.tsx
      repeat/
        RepeatPage.tsx
        RepeatForm.tsx
        RepeatSummary.tsx
        RepeatStabilityTable.tsx
      tokens/
        TokenAnalyticsPage.tsx
        TokenMetricGrid.tsx
      trace/
        TracePage.tsx
        TraceTimeline.tsx
        TraceEventDetail.tsx
      submissions/
        SubmissionsPage.tsx
        SubmissionDetailPage.tsx
        SubmissionTable.tsx
        TestCaseResultTable.tsx
    lib/
      format.ts
      verdict.ts
      ids.ts
      storage.ts
    test/
      server.ts
      fixtures.ts
```

## 前端功能要求

### Dashboard

- 展示后端健康状态。
- 展示题目总数和 submission 统计。
- 等历史列表接口可用后，展示最近的 experiment、compare、repeat。
- 提供快捷入口：single solve、run experiment、compare、repeat。

### 共享变量控制

- 做一个可复用的表单区，统一控制 `model`、`prompt_name`、`agent_name`。
- `model` 使用自由输入，并把最近输入过的模型名存在浏览器 local storage。
- `prompt` 和 `agent` 使用 metadata 接口返回的选项。
- 题目选择器支持按 problem ID / title 多选。
- 执行前要能看见当前选中的配置。

### Problems

- 列出全部题目。
- 展示题目详情、限制、难度、标签、样例。
- 创建题目。
- 添加 testcase。
- 列出 testcase。
- 支持在实验页面选择题目。

### Single Solve

- 输入：一个 problem、model、prompt、agent。
- 调用 `POST /api/v1/ai/solve`。
- 展示 verdict、submission ID、token input / output、total tokens、LLM latency、total latency。
- 展示 prompt preview、raw response、extracted C++ code。
- 链接到 AI solve run detail 和 submission detail。
- 支持用同一组变量再次运行。

### Experiment Run

- 输入：name、problem set、model、prompt、agent。
- 调用 `POST /api/v1/experiments/run`。
- 展示 total / success / AC / failed count。
- 展示 verdict distribution 和 cost summary。
- 展示 run table，包含 problem ID、AI solve run ID、submission ID、verdict、status、error。
- 每一行都能跳到 trace、AI run detail、submission detail。

### Compare

- 输入：name、problem set、baseline model / prompt / agent、candidate model / prompt / agent。
- 调用 `POST /api/v1/experiments/compare`。
- 左右展示 baseline 和 candidate summary。
- 展示 delta AC count、delta failed count、improved / regressed / changed non-AC count。
- 展示 verdict distributions 和 delta distribution。
- 展示 `cost_comparison`，包括 baseline / candidate 的 total tokens、average tokens、total latency、average latency 和 delta。
- 展示 `comparison_summary.tradeoff_type`。
- 展示 highlighted problems 和完整 per-problem table。
- baseline / candidate submission ID 存在时，链接到 submission detail。

### Repeat

- 输入：name、problem set、model、prompt、agent、repeat count，repeat count 范围是 1 到 10。
- 调用 `POST /api/v1/experiments/repeat`。
- 展示 overall AC rate、total run count、best / worst round AC count。
- 展示 round summaries。
- 展示 per-problem stability table。
- 展示 most unstable problems。
- 展示跨轮次 token 和 latency 汇总。

### Token / Cost Analytics

- 从 single solve、experiment、compare、repeat detail 里展示 token summary。
- 统一使用这些字段：
  - input tokens
  - output tokens
  - total tokens
  - average tokens
  - LLM latency
  - total latency
- 不虚构货币价格。
- compare 页面要让 token delta 一眼能看出来。

### Trace

- 第一版从 `GET /api/v1/experiment-runs/:id/trace` 获取 timeline。
- timeline event 类型：
  - prompt
  - llm_response
  - extracted_code
  - judge_summary
  - testcase_results
  - backend_trace_event
- 每个 event 可以打开详情面板。
- 代码和长文本放在可滚动的 code / text block 中。
- 没有 trace 时展示清楚的 empty state，并链接到 AI run / submission detail。

### Submissions

- 分页列出 submissions，支持可选 problem filter。
- 展示 submission detail，包括 source code、verdict、runtime、memory、compile stderr、stdout、stderr、exit code、timeout、testcase results。
- experiment 和 compare 表格可以跳转到这里。

## 实施任务

### Task 1: 创建隔离 Worktree

**Files:**

- 不在主工作区修改源代码文件。

**Steps:**

1. 运行 `git status --short`，记录已有用户改动。
2. 创建 worktree，例如：

```bash
git worktree add .worktrees/frontend-console -b frontend-console
```

3. 在 `.worktrees/frontend-console` 中工作。
4. 校验：

```bash
git status --short
```

Expected: clean；如果分支状态带入了未跟踪 docs，则只出现预期的未跟踪文档。

### Task 2: 添加后端 Metadata 接口

**Files:**

- Create: `internal/handler/meta.go`
- Create: `internal/handler/dto/meta.go`
- Modify: `internal/runtime/router.go`
- Modify: `internal/bootstrap/app.go`
- Modify: `internal/agent/solve.go`
- Modify: `internal/prompt/solve.go`
- Test: `internal/runtime/router_test.go`

**Steps:**

1. 在 `agent` 和 `prompt` 中添加列表函数，例如 `ListSolveAgents()` 和 `ListSolvePrompts()`。
2. 添加 option response DTO。
3. 添加 handler，返回默认 model 和选项列表。
4. 注册 `GET /api/v1/meta/experiment-options`。
5. 添加 router test。
6. 运行：

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/agent ./internal/prompt ./internal/runtime
```

Expected: PASS。

### Task 3: 添加后端历史列表接口

**Files:**

- 修改 experiment、compare、repeat 对应的 repository、service、handler、DTO、router 文件。

**Steps:**

1. 添加分页 repository list 方法，按 `created_at desc` 排序。
2. 添加 service list output。
3. 添加 DTO list response。
4. 添加 handlers：
   - `GET /api/v1/experiments`
   - `GET /api/v1/experiments/compare`
   - `GET /api/v1/experiments/repeat`
5. 保持现有详情路由可用。
6. 运行：

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/repository ./internal/service ./internal/handler ./internal/runtime
```

Expected: PASS。

### Task 4: 添加后端 Trace 接口

**Files:**

- Create: `internal/repository/trace_repository.go`
- Create: `internal/service/trace.go`
- Create: `internal/handler/trace.go`
- Create: `internal/handler/dto/trace.go`
- Modify: `internal/bootstrap/app.go`
- Modify: `internal/runtime/router.go`
- Test: `internal/service/trace_test.go`
- Test: `internal/handler/trace_test.go`

**Steps:**

1. 定义 `TraceEventResponse`，包含 `sequence_no`、`step_type`、`title`、`content`、`metadata`、`created_at`。
2. 根据 experiment run ID 加载关联的 AI solve run 和 submission 数据。
3. 加载已持久化的 `trace_events`，如果有的话。
4. 当持久化 trace 不存在或不完整时，用现有 run 数据合成事件。
5. 注册 `GET /api/v1/experiment-runs/:id/trace`。
6. 运行：

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./internal/service ./internal/handler ./internal/runtime
```

Expected: PASS。

### Task 5: 搭建前端脚手架

**Files:**

- 在 `web/` 下创建所有前端文件。

**Steps:**

1. 在 `web/` 下创建 Vite React TypeScript 项目。
2. 添加 React Router、TanStack Query、测试库和 Playwright。
3. 配置 Vite dev proxy，把 `/api` 和 `/health` 转发到 `http://127.0.0.1:8080`。
4. 添加 root app shell 和空路由。
5. 运行：

```bash
cd web
npm install
npm run build
npm test -- --run
```

Expected: build 和 tests 通过。

### Task 6: 添加类型化 API 层

**Files:**

- Create: `web/src/api/http.ts`
- Create: `web/src/api/types.ts`
- Create: `web/src/api/` 下所有 `*Api.ts` 文件。
- Test: `web/src/api/*.test.ts`

**Steps:**

1. 实现一个小的 `request<T>()` wrapper。
2. 统一解析后端 error response。
3. 添加 health、meta、problems、submissions、AI solve、experiment、compare、repeat、trace 的类型化函数。
4. 使用 mocked fetch 添加 API 单测。
5. 运行：

```bash
cd web
npm test -- --run src/api
```

Expected: PASS。

### Task 7: 构建共享布局和 UI 组件

**Files:**

- Create: `web/src/app/*`
- Create: `web/src/components/layout/*`
- Create: `web/src/components/ui/*`
- Create: `web/src/components/metrics/*`
- Create: `web/src/components/code/*`
- Create: `web/src/lib/*`

**Steps:**

1. 构建带侧边栏导航的 app shell。
2. 添加复用组件：buttons、fields、selects、tabs、tables、badges、loading、error、empty states。
3. 添加 verdict badge 颜色映射：`AC`、`WA`、`CE`、`RE`、`TLE`、`UNJUDGEABLE`。
4. 添加 token 和 latency formatter。
5. 添加 code display component。
6. 运行：

```bash
cd web
npm run build
```

Expected: PASS。

### Task 8: 构建 Problems 功能

**Files:**

- Create: `web/src/features/problems/*`
- Modify: router，添加 `/problems` 和 `/problems/:id`。

**Steps:**

1. 构建 problem list。
2. 构建 problem detail。
3. 构建 create problem form。
4. 构建 testcase panel。
5. 添加 loading / error / empty states。
6. 运行前端测试和 build。

### Task 9: 构建共享变量控制

**Files:**

- Create: `web/src/features/variables/*`
- Test: variable form tests。

**Steps:**

1. 从后端加载 meta options。
2. 构建可复用的 model / prompt / agent controls。
3. 构建 problem multi-picker。
4. 把最近使用的 model name 存入 local storage。
5. 在 solve、experiment、compare、repeat 页面复用这个模块。

### Task 10: 构建 Single Solve 功能

**Files:**

- Create: `web/src/features/solve/*`
- Modify: router，添加 `/solve` 和 `/ai-runs/:id`。

**Steps:**

1. 构建 solve form。
2. 调用 `POST /api/v1/ai/solve`。
3. 展示 solve result panel。
4. 添加 AI solve run detail view。
5. 链接到 submission detail。
6. 添加成功结果和后端错误响应的测试。

### Task 11: 构建 Experiment Run 功能

**Files:**

- Create: `web/src/features/experiments/*`
- Modify: router，添加 `/experiments` 和 `/experiments/:id`。

**Steps:**

1. 构建 experiment run form。
2. 调用 `POST /api/v1/experiments/run`。
3. 构建 experiment detail view。
4. 展示 run table，并链接到 trace、AI run、submission。
5. 展示 verdict distribution 和 cost summary。
6. 添加测试。

### Task 12: 构建 Compare 功能

**Files:**

- Create: `web/src/features/compare/*`
- Modify: router，添加 `/compare` 和 `/compare/:id`。

**Steps:**

1. 使用共享变量控制构建 baseline / candidate form。
2. 调用 `POST /api/v1/experiments/compare`。
3. 构建 compare summary cards。
4. 构建 verdict distribution 和 cost comparison sections。
5. 构建 highlighted problem 和完整 problem summary tables。
6. 链接 submissions。
7. 添加 improved / regressed / cost tradeoff 展示测试。

### Task 13: 构建 Repeat 功能

**Files:**

- Create: `web/src/features/repeat/*`
- Modify: router，添加 `/repeat` 和 `/repeat/:id`。

**Steps:**

1. 构建 repeat form。
2. 在 UI 中限制 repeat count 为 1 到 10。
3. 调用 `POST /api/v1/experiments/repeat`。
4. 构建 repeat summary。
5. 构建 round summary 和 stability tables。
6. 展示 most unstable problems。
7. 添加测试。

### Task 14: 构建 Token Analytics 功能

**Files:**

- Create: `web/src/features/tokens/*`
- Modify: router，添加 `/tokens`。

**Steps:**

1. 构建 token input / output / total 的 metric cards。
2. 构建 latency cards。
3. 在 experiment、compare、repeat、single solve detail 中复用。
4. 添加 compare 专用 token delta 展示。
5. 添加格式化和 delta 符号测试。

### Task 15: 构建 Trace 功能

**Files:**

- Create: `web/src/features/trace/*`
- Modify: router，添加 `/trace/experiment-runs/:id`。

**Steps:**

1. 调用 `GET /api/v1/experiment-runs/:id/trace`。
2. 构建 timeline。
3. 构建 event detail drawer / panel。
4. 安全渲染长 prompt、raw response、code、testcase output。
5. 添加 missing trace 的 empty state。
6. 添加测试。

### Task 16: 构建 Submission Explorer

**Files:**

- Create: `web/src/features/submissions/*`
- Modify: router，添加 `/submissions` 和 `/submissions/:id`。

**Steps:**

1. 构建分页 submission list。
2. 添加 problem filter。
3. 构建 submission detail page。
4. 展示 source code、judge result、compile stderr、stdout / stderr、testcase results。
5. 添加测试。

### Task 17: 构建 Dashboard

**Files:**

- Create: `web/src/features/dashboard/*`
- Modify: 默认路由到 dashboard。

**Steps:**

1. 展示 health status。
2. 展示 submission problem stats。
3. 展示最近 experiment / compare / repeat 列表。
4. 添加 quick action links。
5. 添加测试。

### Task 18: 添加前端集成和开发文档

**Files:**

- Modify: `README.md`
- Modify: `docker-compose.yml`，仅当需要新增可选 frontend service 时修改。
- Create: `web/README.md`

**Steps:**

1. 记录后端启动方式：

```bash
docker compose up -d mysql
go run ./cmd/server
```

2. 记录前端启动方式：

```bash
cd web
npm install
npm run dev
```

3. 记录本地 URL：
   - backend: `http://127.0.0.1:8080`
   - frontend: Vite 默认地址，通常是 `http://127.0.0.1:5173`
4. 记录常见工作流：single solve、compare、trace。

### Task 19: 添加端到端冒烟测试

**Files:**

- Create: `web/e2e/*.spec.ts`
- Modify: `web/playwright.config.ts`

**Steps:**

1. 添加 app loading 和 health panel 冒烟测试。
2. 添加 single solve 页面的 mocked API 冒烟测试。
3. 添加 compare 页面的 mocked API 冒烟测试。
4. 添加 trace 页面的 mocked API 测试。
5. 运行：

```bash
cd web
npm run build
npm test -- --run
npm run e2e
```

Expected: PASS。

### Task 20: 全量验证

运行后端：

```bash
GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go test ./...
```

运行前端：

```bash
cd web
npm run build
npm test -- --run
```

可选浏览器验证：

```bash
cd web
npm run dev
```

打开 Vite URL，手动确认：

- Dashboard 能加载。
- Problems list 能加载。
- Single solve 可以用 mock provider 跑通。
- Experiment run 返回结果。
- Compare 返回 baseline / candidate 结果。
- Repeat 返回稳定性结果。
- token 字段可见。
- 从 experiment run 行可以打开 trace 页面。

## 建议实施顺序

1. 后端 metadata 接口。
2. 后端历史列表接口。
3. 后端 trace 接口。
4. 前端脚手架。
5. API 层。
6. 共享 UI 和变量控制。
7. Problems 和 submissions。
8. Single solve。
9. Experiment run。
10. Compare。
11. Repeat。
12. Token analytics。
13. Trace。
14. Dashboard 打磨和文档。

这个顺序可以避免前端硬编码后端常量，也能让后续每个页面更容易测试。

## 实施前需要确认的问题

1. 第一版前端使用 React + Vite，还是你更想用 Vue？
2. 前端只用 Vite 开发服务器运行，还是希望 Go server 也能托管 build 后的静态文件？
3. 第一版 trace 页面先从已有 solve / submission 数据合成 trace，还是先实现后端真实持久化 `trace_events` 写入？

