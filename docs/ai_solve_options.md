# AI Solve 变量选项说明

本文档说明 AI-FOR-OJ 中 `model`、`prompt_name`、`agent_name` 三类变量的可选项和作用。它们用于单题 AI Solve、批量 Experiment、Repeat 稳定性实验和 Compare 对比实验。

## 总览

| 变量 | 前端字段 | 后端字段 | 作用 |
| --- | --- | --- | --- |
| 模型 | Model | `model` | 决定本次调用哪个 LLM 模型。 |
| 提示词 | Prompt | `prompt_name` | 决定题面如何被组织成发给模型的 prompt。 |
| 智能体 | Agent | `agent_name` | 决定解题流程，是直接生成、先分析再生成，还是失败后修复。 |

前端会通过 `GET /api/v1/meta/experiment-options` 获取默认模型、Prompt 选项和 Agent 选项。当前 Prompt 和 Agent 是固定枚举，Model 是可输入字段，不是固定枚举。

## Model 选择

`model` 表示本次 LLM 请求使用的模型名。当前系统不在前端固定列出所有模型，而是允许用户输入模型名。

### 默认模型

默认值来自后端配置：

```yaml
llm:
  model: mock-cpp17
```

如果请求里没有传 `model`，后端会使用配置里的 `LLM_MODEL` 或 `configs/config.yaml` 中的 `llm.model`。

### `mock-cpp17`

`mock-cpp17` 是本地开发和链路测试用模型。

作用：

- 不调用真实外部 LLM。
- 返回固定的 C++17 mock 代码。
- 适合验证前后端请求、AI Solve 记录、代码提取、判题链路、Experiment/Repeat/Compare 流程。

适用场景：

- 本地没有 API Key。
- 只想验证系统流程是否能跑通。
- 写后端测试或联调前端页面。

### OpenAI-compatible 模型

当后端配置为 `provider=openai_compatible` 时，`model` 会原样传给兼容 `/v1/chat/completions` 的模型网关。

示例：

- `gpt-5.4`
- `claude-opus-4-6`
- `gemini-2.5-pro`
- 其它网关实际支持的模型名

注意：

- 系统不会预先校验模型名是否存在。
- 模型名是否可用取决于当前配置的 LLM 网关。
- 请求级 `model` 优先级高于默认配置。
- 如果模型名填错，错误通常会来自 LLM provider 的响应。

### `glm-*` 模型路由

如果配置了 GLM 路由：

```bash
LLM_GLM_BASE_URL=...
LLM_GLM_API_KEY=...
LLM_GLM_MODEL_PREFIX=glm-
```

则模型名以 `glm-` 开头时，会自动路由到 GLM endpoint。

示例：

- `glm-4.5`
- `glm-4-plus`

未配置 GLM 路由时，`glm-*` 不会自动切换 endpoint，会按默认 OpenAI-compatible endpoint 发送。

## Prompt 选择

`prompt_name` 决定题目内容如何组装成 prompt。当前支持以下三个选项。

### `default`

默认提示词模板。

功能：

- 告诉模型正在解决 OJ / 竞赛编程题。
- 要求输出 C++17 代码。
- 要求用 Markdown `cpp` 代码块返回。
- 包含题目标题、题面、输入格式、输出格式和样例。

适用场景：

- 日常单题 solve。
- 作为实验 baseline。
- 不确定该用哪个 prompt 时优先用它。

### `cpp17_minimal`

更短、更直接的 C++17 prompt。

功能：

- 明确要求只返回一个 `cpp` 代码块。
- 明确使用 C++17。
- 明确使用 stdin/stdout。
- 题面字段更紧凑。

适用场景：

- 想减少 prompt token。
- 想观察更短 prompt 对成本和正确率的影响。
- 批量实验中作为低成本 prompt 对照组。

### `strict_cpp17`

约束最严格的 C++17 prompt。

功能：

- 要求输出完整、可编译的 C++17 程序。
- 要求只返回一个 Markdown `cpp` 代码块。
- 要求使用标准输入输出。
- 明确禁止解释、备注和多个代码块。

适用场景：

- 模型容易输出解释文字或多个代码块时。
- 代码提取失败率较高时。
- 对输出格式稳定性要求更高的实验。

## Agent 选择

`agent_name` 决定 AI Solve 的执行策略。当前支持以下三个选项。

### `direct_codegen`

直接生成代码。

执行流程：

1. 使用选定的 prompt 构造完整解题提示词。
2. 调用一次 LLM。
3. 从 LLM 响应中提取 C++17 代码。
4. 提交判题。
5. 记录本次运行的 prompt、原始响应、代码、token、耗时和 verdict。

适用场景：

- 最基础的 baseline。
- 想比较不同 model 或 prompt 的一次生成能力。
- 成本敏感，希望每题最多调用一次模型。

### `direct_codegen_repair`

直接生成，失败后自动修复。

执行流程：

1. 首轮流程与 `direct_codegen` 相同。
2. 如果判题结果是 `AC`，直接结束。
3. 如果不是 `AC`，把上一轮代码和判题反馈组织成 repair prompt。
4. 再次调用 LLM 生成修复后的代码并判题。
5. 最多总共尝试 3 次。

适用场景：

- 希望提高非 AC 题目的修复机会。
- 想比较“单次生成”和“失败后修复”的收益。
- 可以接受更多 token 和更长耗时。

注意：

- 当前只有 `direct_codegen_repair` 支持失败后自修复。
- 修复依据来自上一轮代码和判题反馈，包括 verdict、错误信息、stderr、超时等可用信息。

### `analyze_then_codegen`

先分析，再生成代码。

执行流程：

1. 先调用一次 LLM 生成题目分析，不写代码。
2. 将分析内容和选定的 solve prompt 合并。
3. 再调用一次 LLM 生成最终 C++17 代码。
4. 提取代码并提交判题。

适用场景：

- 题目需要更强的推理和算法分析。
- 想观察“先分析再写代码”对正确率的影响。
- 可以接受至少两次 LLM 调用带来的成本增加。

注意：

- 当前该 agent 不做失败后修复。
- `AISolveRun` 会记录分析内容预览，便于回看模型的推理方向。

## 常见组合建议

| 目标 | 推荐组合 |
| --- | --- |
| 本地跑通流程 | `mock-cpp17` + `default` + `direct_codegen` |
| 建立基础正确率 baseline | 实际模型 + `default` + `direct_codegen` |
| 降低 prompt 成本 | 实际模型 + `cpp17_minimal` + `direct_codegen` |
| 提高输出格式稳定性 | 实际模型 + `strict_cpp17` + `direct_codegen` |
| 提高失败题修复机会 | 实际模型 + `default` 或 `strict_cpp17` + `direct_codegen_repair` |
| 观察分析步骤是否提升正确率 | 实际模型 + `default` 或 `strict_cpp17` + `analyze_then_codegen` |

## 在不同功能中的使用

### 单题 AI Solve

用于对一个题目运行一次 AI 解题。

```json
{
  "problem_id": 35,
  "model": "gpt-5.4",
  "prompt_name": "default",
  "agent_name": "direct_codegen"
}
```

### Experiment

用于一组题目批量运行同一组变量。

适合回答：

- 某个模型在题目集上的整体 AC 率如何？
- 某个 prompt 在同一批题上的成本和正确率如何？
- 某个 agent 是否比 baseline 更稳定？

### Repeat

用于对同一组题目重复运行同一组变量。

适合回答：

- 同一个 model/prompt/agent 是否稳定？
- 某道题是否容易出现随机失败？
- 平均 token 和耗时是否稳定？

### Compare

用于比较 baseline 和 candidate 两组变量。

适合回答：

- 换模型是否更好？
- 换 prompt 是否更省 token？
- 换 agent 是否能提升 AC 率？
- candidate 是否比 baseline 更贵或更慢？

## 结果记录

每次 AI Solve 会记录以下核心信息：

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

这些字段用于问题回溯、成本分析、效果对比和后续实验统计。
