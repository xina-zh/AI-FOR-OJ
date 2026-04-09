# 当前阶段开发总结

本文档用于沉淀当前阶段已经真实落地的工程能力、系统边界、已验证结论和下一步候选方向，方便后续继续迭代时直接接上。

## 当前定位

当前项目定位是 **AI 算法题实验评测后端**，不是面向普通外部用户的完整 OJ 产品。

当前阶段的核心目标是：

- 提供稳定的题目、测试点、提交、判题基础设施
- 提供足够清晰的 submission 级和 testcase 级可观测信息
- 为后续接入最小 AI solve pipeline 和实验配置对比打基础

这意味着当前接口设计更偏向实验调试和结果分析，而不是用户中心、权限体系、复杂产品交互。

## 当前已完成能力

### 1. 题目与测试点管理

已完成最小题库管理闭环：

- `POST /api/v1/problems`
- `GET /api/v1/problems`
- `GET /api/v1/problems/:id`
- `POST /api/v1/problems/:id/testcases`
- `GET /api/v1/problems/:id/testcases`

当前可以通过正式接口录入题目与测试点，Judge 主流程直接消费这些数据。

### 2. 真实判题

已完成最小真实判题闭环：

- `POST /api/v1/submissions/judge`
- 支持 `cpp17`
- 基于第一版 `DockerSandbox`
- 支持单文件源码编译与标准输入输出执行

Judge 当前会：

1. 读取题目和测试点
2. 编译提交代码
3. 顺序执行每个 testcase
4. 比较输出
5. 汇总 verdict
6. 保存 `Submission` 与 `JudgeResult`

### 3. 已真实验证的 verdict

以下 verdict 已完成真实验证：

- `AC`
- `WA`
- `CE`
- `RE`
- `TLE`

这里的“真实验证”指的是已经通过真实 Docker 编译运行链路完成验证，而不是只通过 mock。

### 4. Submission 查询能力

已完成最小 submission 查询闭环：

- `GET /api/v1/submissions`
- `GET /api/v1/submissions/:id`

当前列表接口支持：

- `page`
- `page_size`
- `problem_id`

默认按最新提交倒序返回。

### 5. 按题聚合 verdict 统计

已完成按题聚合的最小统计接口：

- `GET /api/v1/submissions/stats/problems`

当前可直接返回：

- `problem_id`
- `problem_title`
- `total_submissions`
- `ac_count`
- `wa_count`
- `ce_count`
- `re_count`
- `tle_count`
- `latest_submission_at`

### 6. Submission 级可观测字段

当前 submission detail 已结构化返回这类关键信息：

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

这些字段已经能支持最小失败定位，尤其能区分问题发生在编译阶段还是运行阶段。

### 7. Testcase 级最小结果记录

当前 submission detail 已新增 `testcase_results`，用于承载单次提交在 testcase 维度上的最小执行结果摘要。

每条记录当前最小包含：

- `testcase_id`
- `index`
- `verdict`
- `runtime_ms`
- `stdout`
- `stderr`
- `exit_code`
- `timed_out`

对应表：

- `submission_test_case_results`

该表已通过 `AutoMigrate` 自动创建并在当前环境中确认存在。

## 当前已确认的系统语义

### 1. 首个失败 testcase 即停止

当前 Judge 采用顺序执行、遇首个失败即停止的策略。

因此：

- `AC` 场景会记录所有已执行 testcase 的成功结果
- `WA / RE / TLE` 场景会记录失败发生前已经执行过的 testcase，以及首个失败 testcase
- 首个失败之后的 testcase 不会继续执行，也不会产生结果记录

这是当前阶段的有意设计，目标是先保证主流程简单、稳定、容易调试。

### 2. `CE` 不生成 `testcase_results`

当前 `CE` 场景不会生成 `testcase_results`，这是当前阶段的预期行为，不是 bug。

原因是编译失败发生在 testcase 执行之前，现有 testcase 级记录是在逐个运行 testcase 时生成的。

因此 `CE` 场景主要查看：

- `verdict`
- `compile_stderr`
- `error_message`

而不是 `testcase_results`。

## 当前已真实验证内容与未完成内容

### 已真实验证

已确认的真实能力包括：

- DockerSandbox 能真实编译并运行 `cpp17`
- `AC / WA / CE / RE / TLE` 可通过真实链路触发
- `Submission / JudgeResult` 可正确落库
- submission 列表、详情、分页、按题筛选可用
- 按题聚合 verdict 统计可用
- `submission_test_case_results` 已自动建表
- submission detail 已能返回 testcase 级最小结果

### 尚未完成或未重点验证

以下内容暂未纳入当前阶段完成标准：

- `MLE`
- special judge
- 多语言支持
- 异步判题队列
- 分布式评测
- 更强沙箱安全隔离
- AI solve pipeline
- experiment / experiment_run 维度的统计与回放

## 当前系统边界与已知限制

当前已知限制如下：

- 只支持 `cpp17`
- 只支持普通标准输入输出题
- 不支持 special judge
- 当前仍是第一版 `DockerSandbox`
- `memory_kb` 目前不是精确内存统计
- 仍未做更严格的 seccomp / syscall / namespace 级加固
- 提交记录更偏实验调试用途，没有复杂用户产品能力
- 题目在“无 testcase”场景下的行为还需要后续明确和优化

其中最后一点需要特别说明：

- 当前如果题目没有 testcase，Judge 语义仍需后续明确
- 这不是当前阶段的主阻塞，但属于后续要补齐的系统边界

## 当前开发状态判断

到当前阶段，项目已经从“工程骨架 + 数据模型”进化为：

**一个具备真实判题能力、具备 submission 级与 testcase 级最小可观测能力、可作为后续 AI 算法题实验基础设施继续演进的内部评测后端。**

这意味着：

- OJ 基础设施已经不是纯骨架，而是可运行、可验证、可调试
- 后续如果继续完善 OJ 侧观测，不需要推翻当前结构
- 后续如果接最小 AI solve pipeline，也已经有稳定的题目、提交、判题、结果回看基础

## 下一步候选方向

当前最合理的下一步只有两条主线：

### 方向 A：继续补 OJ 侧观测

适合在接入 AI solve pipeline 之前，把基础设施再打磨一层。

可选项包括：

- 继续优化 submission detail 的失败分析能力
- 明确无 testcase 场景语义
- 视需要补更细的 judge 结果展示

### 方向 B：接最小 AI solve pipeline

前提是接受当前 OJ 基础设施已经达到“最小可用、可调试”的标准。

建议只做最小版本：

- 选定一个模型适配入口
- 固定一个 prompt 模板
- 固定一次 solve -> submit -> judge 的最小闭环
- 先不展开复杂 agent / tooling / experiment 抽象

## 建议路线

如果以当前真实进度为基准，我更建议 **先补一小步 OJ 侧边界清理，再接最小 AI solve pipeline**。

最优先的一条小路线是：

1. 明确并处理“题目无 testcase”的系统语义
2. 然后直接接入最小 AI solve -> 提交 -> 判题闭环

这样推进最稳，既不会过早抽象 experiment 层，也不会在 OJ 基础设施尚有明显边界未定义时贸然接 AI 编排。
