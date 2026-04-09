# Claude Code Method Notes

本文档记录本项目在工程方法上借鉴 Claude Code / Claude Coding 的地方，便于后续理解设计取向。

说明：

- 这里只记录**高层方法论借鉴**
- 不记录、不复制任何现成源码实现
- 不模仿具体文件组织、命名风格、提示词模板或内部机制
- 所有实际代码均按 `AI-For-Oj` 当前项目状态原创落地

## 当前已借鉴的方法点

### 1. 先理解现有结构，再做最小改动

每一步都优先检查当前项目已有分层和边界，然后沿现有链路补功能，而不是新开一套平行体系。

当前体现：

- Judge 继续复用既有 `handler -> service -> judge -> sandbox -> repository`
- AI solve 复用既有 `ProblemRepository` 和 `JudgeSubmissionService`
- 新增 `AISolveRun` 时采用薄模型，而不是强行把当前需求塞进 `ExperimentRun`

### 2. 先明确边界：这一步做什么，不做什么

每次实现前先限定当前步骤的目标，避免一次引入过多抽象。

当前体现：

- 最小 AI solve pipeline 只做单次 `problem -> prompt -> llm -> code -> judge`
- 不提前进入 experiment 编排、多轮 agent、tool calling、trace/token 统计
- AI solve run 记录只保存本次最小回看信息，不展开复杂执行轨迹

### 3. 用清晰的执行链路组织功能

新功能优先围绕一条可解释的执行链组织实现，而不是围绕大而全的框架组织。

当前体现：

- AI solve 的链路是：读取题目 -> 构造 prompt -> 调 LLM -> 提取代码 -> 提交评测 -> 返回结果 -> 记录 run
- testcase 级结果链路是：judge 执行 testcase -> 汇总摘要 -> 落库 -> submission detail 回看

### 4. 优先沿现有分层落地

在当前工程边界足够时，不额外引入新的编排层。

当前体现：

- `handler` 只负责 HTTP 请求与响应
- `service` 负责业务编排
- `repository` 负责落库和查询
- `judge` 负责判题语义
- `sandbox` 负责编译运行
- `llm` 只保留极薄 client 适配层

### 5. 重视可回看性、可调试性

不仅关注是否“跑通”，也关注后续是否能解释一次运行发生了什么。

当前体现：

- submission detail 已有 judge 可观测字段
- 已有 testcase 级最小结果记录
- 新增 `AISolveRun` 用于沉淀单次 AI solve 的 prompt 摘要、原始响应、提取代码、submission_id、verdict、status、error_message

### 6. 控制上下文和返回结构

避免把过多上下文或复杂层级一次性塞进接口与模型。

当前体现：

- `prompt_preview` 使用截断预览，避免一次性返回过长 prompt
- `/api/v1/ai/solve` 的返回只增加最小必要字段 `ai_solve_run_id`
- 更完整的回看信息通过 `GET /api/v1/ai/solve-runs/:id` 查询

### 7. 先做最小闭环，再保留扩展点

先把一条可信链路打通，再考虑更高层的实验系统。

当前体现：

- 在 experiment 系统尚未真正开始前，先用 `AISolveRun` 承接单次 solve 记录
- 未来如果需要，可把 `AISolveRun` 平滑接入 experiment/run 体系，但当前不强行合并
- 在开始最小批量实验时，优先复用已有 `Experiment / ExperimentRun`，只做薄补充而不引入新的批处理框架
- 在开始最小单变量对比实验时，优先把“对比”建模成两次已有实验运行的组合，而不是新起通用变量矩阵系统
- 在开始最小重复运行实验时，优先把“repeat”建模成多次已有 experiment 顺序执行的组合，而不是提前做 benchmark worker、异步队列或复杂统计引擎

## 当前刻意没有借鉴的部分

为了保持本项目简洁和原创，当前刻意没有引入：

- 复杂 runtime/orchestrator
- 多轮 agent
- subagent / tool ecosystem
- 大而全的 prompt registry
- 复杂 trace/span 体系
- 批量实验调度

这些都属于未来可选扩展方向，但不应提前侵入当前最小可用闭环。
