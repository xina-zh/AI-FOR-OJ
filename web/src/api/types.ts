export type ID = number;

export type Page<T> = {
  items: T[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
};

export type OptionItem = {
  name: string;
  label: string;
};

export type ToolingConfig = {
  enabled: string[];
  max_calls: number;
  per_tool_max_calls: Record<string, number>;
};

export type ExperimentOptions = {
  default_model: string;
  models: OptionItem[];
  prompts: OptionItem[];
  agents: OptionItem[];
  tooling_options: OptionItem[];
};

export type Problem = {
  id: ID;
  title: string;
  description: string;
  input_spec: string;
  output_spec: string;
  samples?: string;
  time_limit_ms: number;
  memory_limit_mb: number;
  difficulty: string;
  tags?: string;
  created_at?: string;
  updated_at?: string;
};

export type CreateProblemRequest = {
  title: string;
  description: string;
  input_spec: string;
  output_spec: string;
  samples: string;
  time_limit_ms: number;
  memory_limit_mb: number;
  difficulty: string;
  tags: string;
};

export type TestCase = {
  id: ID;
  problem_id: ID;
  input: string;
  expected_output: string;
  is_sample: boolean;
};

export type CreateTestCaseRequest = {
  input: string;
  expected_output: string;
  is_sample: boolean;
};

export type Submission = {
  id: ID;
  problem_id: ID;
  problem_title?: string;
  language: string;
  source_type: string;
  verdict: string;
  runtime_ms: number;
  memory_kb?: number;
  memory_exceeded?: boolean;
  passed_count: number;
  total_count: number;
  created_at: string;
  updated_at?: string;
};

export type AISolveAttempt = {
  id: ID;
  attempt_no: number;
  stage: string;
  model: string;
  verdict: string;
  failure_type: string;
  repair_reason: string;
  token_input: number;
  token_output: number;
  llm_latency_ms: number;
  total_latency_ms: number;
};

export type AISolveRun = {
  id: ID;
  problem_id: ID;
  model: string;
  prompt_name: string;
  agent_name: string;
  prompt_preview?: string;
  raw_response?: string;
  extracted_code?: string;
  tooling_config: string;
  tool_call_count: number;
  submission_id?: ID;
  verdict: string;
  status: string;
  error_message?: string;
  attempt_count: number;
  failure_type?: string;
  strategy_path?: string;
  token_input: number;
  token_output: number;
  llm_latency_ms: number;
  total_latency_ms: number;
  attempts?: AISolveAttempt[];
  created_at: string;
  updated_at?: string;
};

export type AISolveRequest = {
  problem_id: ID;
  model: string;
  prompt_name: string;
  agent_name: string;
  tooling_config: string;
};

export type VerdictDistribution = {
  ac_count?: number;
  wa_count?: number;
  ce_count?: number;
  re_count?: number;
  tle_count?: number;
  unjudgeable_count?: number;
  unknown_count?: number;
};

export type CostSummary = {
  total_token_input: number;
  total_token_output: number;
  total_tokens: number;
  average_token_input?: number;
  average_token_output?: number;
  average_total_tokens?: number;
  total_llm_latency_ms: number;
  total_latency_ms: number;
  average_llm_latency_ms?: number;
  average_total_latency_ms?: number;
  run_count?: number;
};

export type ExperimentRun = {
  id: ID;
  problem_id: ID;
  ai_solve_run_id?: ID;
  submission_id?: ID;
  attempt_no: number;
  verdict?: string;
  status: string;
  error_message?: string;
  attempt_count: number;
  failure_type?: string;
  strategy_path?: string;
  tooling_config: string;
  tool_call_count: number;
  created_at: string;
};

export type Experiment = {
  id: ID;
  name: string;
  model: string;
  prompt_name: string;
  agent_name: string;
  tooling_config: string;
  status: string;
  total_count: number;
  success_count: number;
  ac_count: number;
  failed_count: number;
  verdict_distribution: VerdictDistribution;
  cost_summary: CostSummary;
  created_at: string;
  updated_at: string;
  runs: ExperimentRun[];
};

export type ExperimentCompare = {
  id: ID;
  name: string;
  compare_dimension: string;
  baseline_value: string;
  candidate_value: string;
  baseline_prompt_name: string;
  candidate_prompt_name: string;
  baseline_agent_name: string;
  candidate_agent_name: string;
  baseline_tooling_config: string;
  candidate_tooling_config: string;
  baseline_experiment_id?: ID;
  candidate_experiment_id?: ID;
  baseline_summary?: Experiment;
  candidate_summary?: Experiment;
  baseline_verdict_distribution?: VerdictDistribution;
  candidate_verdict_distribution?: VerdictDistribution;
  delta_verdict_distribution?: VerdictDistribution;
  cost_comparison?: ExperimentCompareCostComparison;
  improved_count?: number;
  regressed_count?: number;
  changed_non_ac_count?: number;
  problem_summaries?: ExperimentCompareProblemSummary[];
  highlighted_problems?: ExperimentCompareProblemSummary[];
  delta_ac_count?: number;
  delta_failed_count?: number;
  status: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
};

export type ExperimentCompareCostComparison = {
  baseline_total_tokens: number;
  baseline_total_latency_ms: number;
  candidate_total_tokens: number;
  candidate_total_latency_ms: number;
  delta_total_tokens: number;
  delta_total_latency_ms: number;
};

export type ExperimentCompareProblemSummary = {
  problem_id: ID;
  baseline_verdict?: string;
  candidate_verdict?: string;
  changed: boolean;
  change_type: string;
  baseline_submission_id?: ID;
  candidate_submission_id?: ID;
};

export type CompareExperimentRequest = {
  name: string;
  problem_ids: ID[];
  baseline_model: string;
  candidate_model: string;
  baseline_prompt_name: string;
  candidate_prompt_name: string;
  baseline_agent_name: string;
  candidate_agent_name: string;
  baseline_tooling_config: string;
  candidate_tooling_config: string;
};

export type ExperimentRepeatRoundSummary = {
  round_no: number;
  experiment_id: ID;
  ac_count: number;
  failed_count: number;
  verdict_distribution: VerdictDistribution;
  status: string;
};

export type ExperimentRepeatProblemSummary = {
  problem_id: ID;
  total_rounds: number;
  ac_count: number;
  failed_count: number;
  ac_rate: number;
  latest_verdict?: string;
  verdict_distribution: VerdictDistribution;
};

export type ExperimentRepeat = {
  id: ID;
  name: string;
  model: string;
  prompt_name: string;
  agent_name: string;
  tooling_config: string;
  problem_ids: ID[];
  repeat_count: number;
  experiment_ids: ID[];
  total_problem_count: number;
  total_run_count: number;
  overall_ac_count: number;
  overall_failed_count: number;
  overall_ac_rate: number;
  cost_summary: CostSummary;
  status: string;
  error_message?: string;
  round_summaries?: ExperimentRepeatRoundSummary[];
  problem_summaries?: ExperimentRepeatProblemSummary[];
  most_unstable_problems?: ExperimentRepeatProblemSummary[];
  created_at: string;
  updated_at: string;
};

export type RepeatExperimentRequest = {
  name: string;
  problem_ids: ID[];
  model: string;
  prompt_name: string;
  agent_name: string;
  tooling_config: string;
  repeat_count: number;
};

export type TraceEvent = {
  id: ID;
  sequence_no: number;
  step_type: string;
  content: string;
  metadata: string;
  created_at: string;
};

export type ExperimentRunTrace = {
  experiment_run_id: ID;
  experiment_id: ID;
  problem_id: ID;
  ai_solve_run_id?: ID;
  submission_id?: ID;
  attempt_no: number;
  verdict?: string;
  status: string;
  error_message?: string;
  timeline: TraceEvent[];
  ai_solve_run?: AISolveRun;
  submission?: Submission;
};

export type RunExperimentRequest = {
  name: string;
  problem_ids: ID[];
  model: string;
  prompt_name: string;
  agent_name: string;
  tooling_config: string;
};
