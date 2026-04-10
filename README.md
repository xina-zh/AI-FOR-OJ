# AI-For-Oj

`AI-For-Oj` 当前定位不是普通外部 OJ，而是一个面向后续变量控制与结果分析的 **AI 算法题实验平台后端**。

当前阶段的重点已经不是“只有工程骨架”，而是先把一套可信、可回看、可继续扩展的最小实验闭环搭起来：

- OJ 基础评测链路可真实运行
- AI solve 已能走通单次解题闭环
- experiment / compare / repeat 已具备最小实验运行能力
- 当前已经有一层可用的最小分析视图，便于观察 AC、失败、稳定性和差异

## 当前阶段里程碑

当前这一版可以概括为：

**最小实验闭环与基础分析能力已形成。**

也就是说，项目已经具备：

- 题目与测试点录入
- 真实 `cpp17` 判题
- submission 可回看
- AI solve 可落库回放
- 批量实验 / 对比实验 / 重复实验
- 最小 verdict 分布、按题稳定性、按题差异分析

对应的阶段总结文档见：

- [docs/dev_progress.md](/home/xina/projects/AI-For-Oj/docs/dev_progress.md)

## 已完成能力

### OJ 基础评测层

- 题目管理
  - `POST /api/v1/problems`
  - `GET /api/v1/problems`
  - `GET /api/v1/problems/:id`
- 测试点管理
  - `POST /api/v1/problems/:id/testcases`
  - `GET /api/v1/problems/:id/testcases`
- 提交评测
  - `POST /api/v1/submissions/judge`
- 真实沙箱执行
  - 第一版 `DockerSandbox`
  - 当前只支持 `cpp17`
  - 普通标准输入输出题
- 当前支持的 verdict
  - `AC`
  - `WA`
  - `CE`
  - `RE`
  - `TLE`
  - `UNJUDGEABLE`

### Submission 与判题可观测层

- 提交列表与详情
  - `GET /api/v1/submissions`
  - `GET /api/v1/submissions/:id`
- 提交列表支持
  - `page`
  - `page_size`
  - `problem_id`
- submission detail 当前可回看
  - `verdict`
  - `runtime_ms`
  - `memory_kb`
  - `passed_count`
  - `total_count`
  - `compile_stderr`
  - `run_stdout`
  - `run_stderr`
  - `exit_code`
  - `timed_out`
  - `exec_stage`
  - `error_message`
  - `testcase_results`
- 按题聚合 submission 统计
  - `GET /api/v1/submissions/stats/problems`

### AI Solve 层

- 最小 AI solve 闭环
  - `POST /api/v1/ai/solve`
  - `problem -> prompt -> llm -> 提取 cpp17 代码 -> judge`
- 单次 AI solve 记录
  - `GET /api/v1/ai/solve-runs/:id`
- 当前默认支持本地闭环
  - `mock` LLM provider
- AI 提交已标记
  - `source_type = ai`

### 实验运行层

- 最小批量实验
  - `POST /api/v1/experiments/run`
  - `GET /api/v1/experiments/:id`
- 最小单变量 compare
  - `POST /api/v1/experiments/compare`
  - `GET /api/v1/experiments/compare/:id`
  - 当前只支持单变量 `model` 对比
- 最小 repeat 实验
  - `POST /api/v1/experiments/repeat`
  - `GET /api/v1/experiments/repeat/:id`

### 最小分析层

- experiment verdict 分布
- compare baseline / candidate verdict 分布
- repeat 按题稳定性统计
- repeat 最不稳定题列表
- compare 按题差异统计
- compare highlighted problems

## 当前阶段已验证内容

当前已经完成真实验证的内容包括：

- `DockerSandbox` 能真实编译并运行 `cpp17`
- `AC / WA / CE / RE / TLE` 已经过真实链路验证
- `UNJUDGEABLE` 语义已接入并用于“无 testcase 不可评测”
- submission / judge result / testcase result 能正确落库与查询
- 最小 AI solve 闭环已验收通过
- experiment / compare / repeat 的最小 service 链路已通过测试与构建验证

## 当前阶段系统语义

- Judge 采用“首个失败 testcase 即停止”
- `CE` 不生成 `testcase_results`
- 无 testcase 不再误判为 `AC`，而是返回 `UNJUDGEABLE`
- compare 中按题变化当前采用最小规则
  - `regressed`
  - `improved`
  - `changed_non_ac`
  - `same`
- repeat 中“最不稳定题”当前采用最小规则
  - `instability_score = min(ac_count, failed_count)`

## 当前未做内容

当前阶段明确还没有展开的内容包括：

- 异步执行 / 后台任务 / 队列系统
- token / latency 统计
- prompt / tooling / agent 变量矩阵
- 多语言支持
- stronger sandbox / 更严格隔离
- special judge
- 通用 benchmark 分析平台
- 前端页面与可视化展示

## 当前数据库迁移策略

当前阶段仍使用 GORM `AutoMigrate`。

原因：

- 项目仍处于快速迭代期
- experiment 相关模型还在持续收敛
- 当前更优先保证闭环与演进速度

后续当表结构稳定、真实数据开始积累、或多人协作改表变频繁时，再平滑升级到版本化 migration。

关键位置：

- 启动入口：[cmd/server/main.go](/home/xina/projects/AI-For-Oj/cmd/server/main.go)
- 启动装配：[internal/bootstrap/app.go](/home/xina/projects/AI-For-Oj/internal/bootstrap/app.go)
- 自动迁移：[internal/bootstrap/migrate.go](/home/xina/projects/AI-For-Oj/internal/bootstrap/migrate.go)
- 模型注册：[internal/model/schema.go](/home/xina/projects/AI-For-Oj/internal/model/schema.go)

## 本地启动

1. 启动依赖：

```bash
docker compose up -d mysql
```

2. 启动服务：

```bash
go run ./cmd/server
```

3. 健康检查：

```bash
curl --noproxy '*' -sS http://127.0.0.1:8080/health
```

如果要走真实判题链路，需要提前准备：

```bash
docker pull gcc:13
```

## LLM Provider 切换

当前默认：

- `provider=mock`

这适合本地无 key 的开发和测试。

如果要切到真实 OpenAI-compatible 网关，例如 gptsapi，可使用：

```bash
export LLM_PROVIDER=openai_compatible
export LLM_BASE_URL=https://api.gptsapi.net/v1
export LLM_API_KEY=your-api-key
export LLM_MODEL=your-actual-model-name
```

说明：

- 当前只接 `OpenAI-compatible /v1/chat/completions`
- 模型名以中转站实际支持的名称为准
- 每次 AI solve 都会记录实际 `model`
- `AISolveRun` 也会记录最小 `token_input / token_output / llm_latency_ms / total_latency_ms`

## 当前路线说明

当前阶段：

- 最小实验闭环与基础分析能力

下一阶段建议方向：

- 更强实验指标
- 更明确的变量控制
- 更稳定的实验可复现能力

但在进入下一阶段之前，当前这一版已经适合作为一个清晰的阶段性里程碑提交到 GitHub。
