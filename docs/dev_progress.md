# 阶段性里程碑：最小实验闭环与基础分析能力

本文档用于给当前阶段做一次明确收束，说明项目现在已经做到哪里、哪些能力已经真实落地、哪些边界仍然存在，以及下一阶段最适合往哪里推进。

当前这份总结不是产品宣传，而是面向后续协作与继续开发的工程状态说明。

## 当前阶段目标

当前阶段的核心目标不是直接做完整实验平台，而是先完成一套 **可信、可运行、可回看、可继续扩展** 的最小实验闭环。

对应到工程上，主要分为四层：

1. OJ 基础评测层
2. AI solve 层
3. 实验运行层
4. 最小分析层

当前阶段的重点，是先让这四层形成一条可验证的主链，而不是提前进入复杂的 benchmark / orchestration / agent 抽象。

## 本阶段已完成内容

### 1. OJ 基础评测层

已经完成：

- 题目管理
- 测试点管理
- 提交评测
- `cpp17` 真实编译与执行
- 第一版 `DockerSandbox`
- `AC / WA / CE / RE / TLE / UNJUDGEABLE`

已经形成的最小可信链路：

- 录题
- 录测试点
- 提交代码
- 真实判题
- 保存 submission / judge result
- 详情回看

### 2. 判题可观测层

已经完成：

- submission 列表
- submission 详情
- 按题聚合 submission 统计
- submission 级 judge 可观测字段
- testcase 级最小结果记录

当前已经能回答这些问题：

- 这次提交最终 verdict 是什么
- 编译失败还是运行失败
- stdout / stderr 是什么
- 是否超时
- 哪个 testcase 首先失败

### 3. AI Solve 层

已经完成：

- 最小 AI solve 闭环
- `POST /api/v1/ai/solve`
- `AISolveRun` 单次运行记录
- `GET /api/v1/ai/solve-runs/:id`
- mock provider 本地闭环

当前已经能走通：

- problem
- prompt
- llm
- 提取 `cpp17` 代码
- 提交 judge
- 返回 verdict
- 沉淀单次运行记录

### 4. 实验运行层

已经完成：

- 最小批量实验 `Experiment + ExperimentRun`
- 最小单变量 model 对比实验 `ExperimentCompare`
- 最小重复运行实验 `ExperimentRepeat`

当前已经具备：

- 固定配置跑一批题
- 同一批题做 baseline / candidate 对比
- 同一批题、同一配置连续跑多轮

### 5. 最小分析层

已经完成：

- experiment verdict 分布
- compare verdict 分布
- repeat 按题稳定性统计
- repeat 最不稳定题列表
- compare 按题差异统计
- compare highlighted problems

当前已经能直接观察：

- 哪种配置 AC 更多
- 哪些题更稳定或更不稳定
- 哪些题在 baseline / candidate 间发生了关键变化

## 本阶段已验证内容

当前已确认的真实落地能力包括：

- `DockerSandbox` 真实运行 `cpp17`
- `AC / WA / CE / RE / TLE` 已完成真实验证
- “无 testcase 不再误判 AC” 已修复，并引入 `UNJUDGEABLE`
- submission / judge result / testcase result 可正确落库与查询
- 最小 AI solve 闭环已验收通过
- experiment / compare / repeat 的 service 链路已通过测试与构建验证

代码层面的基础验证已覆盖：

- `go test ./...`
- `go build ./cmd/server`

## 当前阶段系统语义

当前已确认的关键语义包括：

- Judge 采用“首个失败 testcase 即停止”
- `CE` 不生成 `testcase_results`
- 无 testcase 时返回 `UNJUDGEABLE`
- compare 的按题变化规则当前是：
  - `regressed`
  - `improved`
  - `changed_non_ac`
  - `same`
- repeat 的“最不稳定题”当前使用最小排序规则：
  - `instability_score = min(ac_count, failed_count)`

这些规则当前都偏向“简单、可解释、便于分析”，而不是复杂评分系统。

## 当前限制

这一阶段虽然已经形成明确成果，但边界也很清楚。

当前还没有做：

- 异步执行
- token / latency 指标
- prompt / tooling / agent 变量矩阵
- 多语言
- stronger sandbox
- special judge
- 更完整的 trace / span / replay
- 更复杂的 benchmark 统计与分析
- 前端页面与可视化

另外，当前分析能力仍然是“最小实验分析”，还不是通用实验平台或完整 benchmark 产品。

## 当前阶段状态判断

现在这个项目已经不再是“工程骨架”或“只会判题的最小 OJ”，而是：

**一个具备真实判题能力、最小 AI solve 能力、最小实验运行能力、以及基础分析能力的 AI 算法题实验后端。**

这意味着：

- OJ 基础层已经可信
- AI solve 已经打通
- experiment / compare / repeat 已经能运行
- 最小分析视图已经能支撑下一阶段继续扩展指标与变量控制

## 下一阶段建议方向

当前最适合的下一阶段方向，不是继续零散补小点，而是围绕“更明确的实验控制与更强指标”展开。

建议优先级如下：

1. 增强实验指标
   - token
   - latency
   - attempt 成本

2. 增强变量控制
   - model
   - prompt
   - tooling
   - agent

3. 增强实验可复现能力
   - 更清晰的配置快照
   - 更稳定的 run 记录结构

4. 视需要再补更强沙箱能力
   - 更严格隔离
   - 更多运行约束

## 阶段结论

当前这一版已经可以视为一个明确的阶段性里程碑：

**最小实验闭环已完成，基础分析能力已具备，适合先收束并提交 GitHub。**

后续再往前推进时，可以基于这一版继续扩展，而不需要推翻现有目录结构和主链设计。

## 2026-04-13 开发补充

### 1. 成本比较主线补齐

今天把 token / latency 的汇总主线从单次 `AISolveRun` 继续向上补齐到了实验层：

- `Experiment` 层增加了 `cost_summary`
- `Compare` 层增加了 `cost_comparison`
- `Repeat` 层增加了 `cost_summary`

当前这些汇总主要覆盖：

- token 输入总量
- token 输出总量
- total tokens
- llm latency 总量
- total latency 总量
- 对应平均值

其中：

- `Compare` 的成本对比直接复用了 baseline / candidate 两侧 `Experiment` 的 `cost_summary`
- `Repeat` 的成本汇总直接复用了每一轮 `Experiment` 的 `cost_summary`

这样避免了在 compare / repeat 层重复扫描更底层的统计对象，也保证了口径一致。

### 2. 请求级 model 变量化完成

今天把 model 从“主要依赖全局默认配置”进一步收敛为“请求级 / 实验级可指定”的最小闭环。

当前已打通：

- `POST /api/v1/ai/solve`
- `POST /api/v1/experiments/run`
- `POST /api/v1/experiments/compare`
- `POST /api/v1/experiments/repeat`

当前语义是：

- 请求里传了 `model` 时，优先使用请求值
- 没传时，才回退到默认配置

另外：

- compare 的 `baseline_model` / `candidate_model` 已能分别独立传到底层
- repeat 顶层 `model` 已能传到每一轮 `experiment`，再继续传到底层 solve

### 3. 真实模型切换方式已验证

真实模型接入主线仍固定为 `openai_compatible` 中转站方案。

在固定 `provider / base_url / api_key`、不依赖手动重启容器的前提下，已经实际验证可通过请求里的 `model` 直接切换不同模型，包括：

- GPT
- Gemini
- Claude

这一步验证的重点不是扩 provider，而是确认后续做模型对比实验时，不再主要依赖手动改全局 `LLM_MODEL`。

### 4. 修复一个历史遗留问题

今天还修复了 `experiments.problem_id` 历史字段仍为 `NOT NULL`，导致多题 batch experiment 创建失败的问题。

当前该字段已允许为空，和现在的多题 `experiment` 语义保持一致。修复后：

- `experiment`
- `compare`
- `repeat`

这三条依赖 experiment 创建的实验链路都恢复正常。

### 5. 当前阶段补充结论

经过今天这一轮，当前项目已经从“能比较结果”进一步推进到“能比较结果 + 能比较成本”，并且模型切换已经具备了请求级控制能力，为后续做更自然的模型对比实验打下了基础。

### 6. 下一步建议

下一步更适合优先增强 compare 的实验结论表达能力，而不是继续扩基础成本字段。

### Adaptive Repair Agent

- 新增 `adaptive_repair_v1`，按判题结果选择修复路径。
- 在 `ai_solve_attempts` 中记录每次 attempt 的 prompt/code/judge/cost 元数据。
- 将 `WA`、`RE`、`TLE` 分别路由到不同修复提示词。
- 实验运行输出增加 `attempt_count`、`strategy_path` 和 `failure_type`。
- 本地 smoke 已验证 `adaptive_repair_v1`：`problem_id=1`，`ai_solve_run_id=2`，可通过 `POST /api/v1/ai/solve` 和 `GET /api/v1/ai/solve-runs/:id` 取回运行详情。
