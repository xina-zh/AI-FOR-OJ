# AI-For-Oj

`AI-For-Oj` 当前定位不是普通外部 OJ，而是一个面向后续变量控制与结果分析的 **AI 算法题实验平台后端**。

当前阶段的重点已经不是“只有工程骨架”，而是先把一套可信、可回看、可继续扩展的最小实验闭环搭起来：

- OJ 基础评测链路可真实运行
- AI solve 已能走通单次解题闭环
- experiment / compare / repeat 已具备最小实验运行能力
- 当前已经有一层可用的最小分析视图，便于观察 AC、失败、稳定性和差异

## 当前阶段里程碑

当前这一版可以概括为：

**最小实验闭环、基础成本统计、变量控制与初版 Agent 实验能力已形成。**

也就是说，项目已经具备：

- 题目与测试点录入
- 真实 `cpp17` 判题
- submission 可回看
- AI solve 可落库回放
- model / prompt / agent 变量化
- 批量实验 / 对比实验 / 重复实验
- 最小 verdict 分布、按题稳定性、按题差异分析
- Experiment / Compare / Repeat 成本汇总
- Compare 结构化实验结论

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
- 当前支持请求级变量控制
  - `model`
  - `prompt_name`
  - `agent_name`
- 当前内置 agent
  - `direct_codegen`
  - `direct_codegen_repair`
  - `analyze_then_codegen`
- `direct_codegen_repair` 支持最小失败后修复闭环
  - 首次生成代码并判题
  - 非 `AC` 时基于上一轮代码与判题反馈继续修复
  - 最多总尝试 3 次
- AI 提交已标记
  - `source_type = ai`
- `AISolveRun` 当前可回看
  - `model`
  - `prompt_name`
  - `agent_name`
  - `prompt_preview`
  - `raw_response`
  - `extracted_code`
  - `submission_id`
  - `verdict`
  - `status`
  - `error_message`
  - `token_input`
  - `token_output`
  - `llm_latency_ms`
  - `total_latency_ms`

### 实验运行层

- 最小批量实验
  - `POST /api/v1/experiments/run`
  - `GET /api/v1/experiments/:id`
- 最小单变量 compare
  - `POST /api/v1/experiments/compare`
  - `GET /api/v1/experiments/compare/:id`
  - 当前可用于 `model / prompt / agent` 等单变量对比
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
- experiment `cost_summary`
- compare `cost_comparison`
- repeat `cost_summary`
- compare `comparison_summary`
  - `candidate_better_ac / worse_ac / same_ac`
  - `candidate_more_expensive / cheaper / same_cost`
  - `candidate_slower / faster / same_latency`
  - `tradeoff_type`

### 本地批量导题工具

- 当前已提供最小本地批量导题脚本
  - `scripts/import_problems.py`
- 目标是便于本地批量导入实验题目，不依赖手工逐题录入
- 固定目录格式示例

```text
problems/
  shortest-path/
    statement.txt
    1.in
    1.out
  tree-dp/
    statement.txt
    1.in
    1.out
```

- `dry-run`

```bash
python3 scripts/import_problems.py --dir ./problems/ready --dry-run
```

- 真正导入

```bash
python3 scripts/import_problems.py --dir ./problems/ready
```

## 当前阶段已验证内容

当前已经完成真实验证的内容包括：

- `DockerSandbox` 能真实编译并运行 `cpp17`
- `AC / WA / CE / RE / TLE` 已经过真实链路验证
- `UNJUDGEABLE` 语义已接入并用于“无 testcase 不可评测”
- submission / judge result / testcase result 能正确落库与查询
- 最小 AI solve 闭环已验收通过
- experiment / compare / repeat 的最小 service 链路已通过测试与构建验证
- 请求级 `model / prompt_name / agent_name` 已打通到底层 LLM 调用
- 真实 OpenAI-compatible 网关下，已验证可在同一容器内切换不同模型
- 中文题目、中文 prompt preview、中文 error message 已可稳定写库

## 当前阶段系统语义

- Judge 采用“首个失败 testcase 即停止”
- `CE` 不生成 `testcase_results`
- 无 testcase 不再误判为 `AC`，而是返回 `UNJUDGEABLE`
- `direct_codegen` 当前为单次生成，不自动修复
- `direct_codegen_repair` 当前为单次生成 + 最多 2 次修复重试
- `analyze_then_codegen` 当前为“分析后生成代码”的两步 agent，不自动修复
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
- tooling / verifier / critic / planner
- 多变量实验矩阵
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
- 每次 AI solve / experiment / compare / repeat 都会优先使用请求指定的 `model`
- 未显式传入时，再 fallback 到默认 `LLM_MODEL`
- 每次 AI solve 都会记录实际 `model`
- `AISolveRun` 会记录最小 `token_input / token_output / llm_latency_ms / total_latency_ms`

## Prompt 与 Agent 变量

当前 `prompt_name` 支持：

- `default`
- `cpp17_minimal`
- `strict_cpp17`

当前 `agent_name` 支持：

- `direct_codegen`
- `direct_codegen_repair`
- `analyze_then_codegen`

说明：

- `prompt_name` 控制提示词模板
- `agent_name` 控制解题编排策略
- compare 可用于同模型下比较不同 prompt 或不同 agent

## 当前路线说明

当前阶段：

- 最小实验闭环、基础成本统计、变量控制与初版 Agent 实验能力

下一阶段建议方向：

- 继续提升单次 AI solve 稳定完成率
- 增强 compare 的实验结论表达能力
- 在保持最小实现的前提下继续补强变量实验能力

但在进入下一阶段之前，当前这一版已经适合作为一个清晰的阶段性里程碑提交到 GitHub。
