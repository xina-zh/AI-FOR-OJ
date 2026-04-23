export type Verdict = 'AC' | 'WA' | 'CE' | 'RE' | 'TLE' | 'UNJUDGEABLE' | string;

export interface OptionItem {
  name: string;
  label: string;
}

export interface ExperimentOptions {
  default_model: string;
  prompts: OptionItem[];
  agents: OptionItem[];
  tooling_options: OptionItem[];
}

export interface HealthResponse {
  status: string;
  app: string;
  env: string;
  database: string;
}

export interface Problem {
  id: number;
  title: string;
  description: string;
  input_spec: string;
  output_spec: string;
  samples: string;
  time_limit_ms: number;
  memory_limit_mb: number;
  difficulty: string;
  tags: string;
}

export interface TestCase {
  id: number;
  problem_id: number;
  input: string;
  expected_output: string;
  is_sample: boolean;
}

export interface VerdictDistribution {
  ac_count?: number;
  wa_count?: number;
  ce_count?: number;
  re_count?: number;
  tle_count?: number;
  unjudgeable_count?: number;
  unknown_count?: number;
}

export interface ExperimentCostSummary {
  total_token_input: number;
  total_token_output: number;
  total_tokens: number;
  average_token_input: number;
  average_token_output: number;
  average_total_tokens: number;
  total_llm_latency_ms: number;
  total_latency_ms: number;
  average_llm_latency_ms: number;
  average_total_latency_ms: number;
  run_count: number;
}

export interface ExperimentRun {
  id: number;
  problem_id: number;
  ai_solve_run_id?: number;
  submission_id?: number;
  attempt_no: number;
  verdict?: Verdict;
  status: string;
  error_message?: string;
  attempt_count: number;
  failure_type?: string;
  strategy_path?: string;
  tooling_config: string;
  tool_call_count: number;
  created_at: string;
}

export interface Experiment {
  id: number;
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
  cost_summary: ExperimentCostSummary;
  created_at: string;
  updated_at: string;
  runs: ExperimentRun[];
}

export interface Page<T> {
  items: T[];
  page: number;
  page_size: number;
  total: number;
  total_pages: number;
}

export interface AISolveResponse {
  ai_solve_run_id: number;
  problem_id: number;
  model?: string;
  prompt_name: string;
  agent_name: string;
  prompt_preview: string;
  raw_response?: string;
  extracted_code?: string;
  submission_id: number;
  verdict?: Verdict;
  error_message?: string;
  attempt_count: number;
  failure_type?: string;
  strategy_path?: string;
  tooling_config: string;
  tool_call_count: number;
  token_input: number;
  token_output: number;
  llm_latency_ms: number;
  total_latency_ms: number;
  attempts?: AISolveAttempt[];
}

export interface AISolveRun {
  id: number;
  problem_id: number;
  model?: string;
  prompt_name: string;
  agent_name: string;
  prompt_preview?: string;
  raw_response?: string;
  extracted_code?: string;
  submission_id?: number;
  verdict?: Verdict;
  status: string;
  error_message?: string;
  attempt_count: number;
  failure_type?: string;
  strategy_path?: string;
  tooling_config: string;
  tool_call_count: number;
  token_input: number;
  token_output: number;
  llm_latency_ms: number;
  total_latency_ms: number;
  created_at: string;
  updated_at: string;
  attempts?: AISolveAttempt[];
}

export interface AISolveAttempt {
  id: number;
  attempt_no: number;
  stage: string;
  model: string;
  verdict?: Verdict;
  failure_type?: string;
  repair_reason?: string;
  strategy_path?: string;
  prompt_preview?: string;
  extracted_code?: string;
  judge_passed_count: number;
  judge_total_count: number;
  timed_out: boolean;
  error_message?: string;
  token_input: number;
  token_output: number;
  llm_latency_ms: number;
  total_latency_ms: number;
}

export interface CompareExperiment {
  id: number;
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
  problem_ids: number[];
  baseline_experiment_id: number;
  candidate_experiment_id: number;
  baseline_summary?: Experiment;
  candidate_summary?: Experiment;
  baseline_verdict_distribution: VerdictDistribution;
  candidate_verdict_distribution: VerdictDistribution;
  delta_verdict_distribution: VerdictDistribution;
  cost_comparison: ExperimentCompareCostComparison;
  comparison_summary: ExperimentCompareSummary;
  improved_count: number;
  regressed_count: number;
  changed_non_ac_count: number;
  problem_summaries: ExperimentCompareProblemSummary[];
  highlighted_problems: ExperimentCompareHighlightedProblem[];
  delta_ac_count: number;
  delta_failed_count: number;
  status: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
}

export interface ExperimentCompareCostComparison {
  baseline_total_token_input?: number;
  baseline_total_token_output?: number;
  baseline_total_tokens?: number;
  candidate_total_token_input?: number;
  candidate_total_token_output?: number;
  candidate_total_tokens?: number;
  delta_total_tokens?: number;
  baseline_average_total_latency_ms?: number;
  candidate_average_total_latency_ms?: number;
  delta_average_total_latency_ms?: number;
}

export interface ExperimentCompareSummary {
  accuracy_winner?: string;
  cost_winner?: string;
  latency_winner?: string;
  tradeoff_type?: string;
  [key: string]: boolean | number | string | undefined;
}

export interface ExperimentCompareProblemSummary {
  problem_id: number;
  baseline_verdict?: Verdict;
  candidate_verdict?: Verdict;
  changed: boolean;
  change_type: string;
  baseline_status?: string;
  candidate_status?: string;
  baseline_submission_id?: number;
  candidate_submission_id?: number;
}

export interface ExperimentCompareHighlightedProblem extends ExperimentCompareProblemSummary {}

export interface RepeatExperiment {
  id: number;
  name: string;
  model: string;
  prompt_name: string;
  agent_name: string;
  tooling_config: string;
  problem_ids: number[];
  repeat_count: number;
  experiment_ids: number[];
  total_problem_count: number;
  total_run_count: number;
  overall_ac_count: number;
  overall_failed_count: number;
  average_ac_count_per_round: number;
  average_failed_count_per_round: number;
  overall_ac_rate: number;
  best_round_ac_count: number;
  worst_round_ac_count: number;
  cost_summary: ExperimentCostSummary;
  status: string;
  error_message?: string;
  round_summaries: ExperimentRepeatRoundSummary[];
  problem_summaries: ExperimentRepeatProblemSummary[];
  most_unstable_problems: ExperimentRepeatUnstableProblem[];
  created_at: string;
  updated_at: string;
}

export interface ExperimentRepeatRoundSummary {
  round_no: number;
  experiment_id: number;
  ac_count: number;
  failed_count: number;
  verdict_distribution: VerdictDistribution;
  status: string;
}

export interface ExperimentRepeatProblemSummary {
  problem_id: number;
  total_rounds: number;
  ac_count: number;
  failed_count: number;
  ac_rate: number;
  verdict_distribution: VerdictDistribution;
  latest_verdict?: Verdict;
}

export interface ExperimentRepeatUnstableProblem extends ExperimentRepeatProblemSummary {
  instability_score: number;
  verdict_kind_count: number;
}

export interface SubmissionSummary {
  id: number;
  problem_id: number;
  problem_title: string;
  language: string;
  source_type: string;
  verdict: Verdict;
  runtime_ms: number;
  passed_count: number;
  total_count: number;
  created_at: string;
  updated_at: string;
}

export interface SubmissionProblemStats {
  problem_id: number;
  problem_title: string;
  total_submissions: number;
  ac_count: number;
  wa_count?: number;
  ce_count?: number;
  re_count?: number;
  tle_count?: number;
  latest_submission_at?: string;
}

export interface SubmissionDetail extends SubmissionSummary {
  source_code: string;
  memory_kb: number;
  compile_stderr?: string;
  run_stdout?: string;
  run_stderr?: string;
  exit_code: number;
  timed_out: boolean;
  exec_stage?: string;
  error_message?: string;
  judge_result?: SubmissionJudgeResult;
  testcase_results?: SubmissionTestCaseResult[];
}

export interface SubmissionJudgeResult {
  verdict: Verdict;
  runtime_ms: number;
  memory_kb: number;
  passed_count: number;
  total_count: number;
}

export interface SubmissionTestCaseResult {
  testcase_id: number;
  index: number;
  verdict: Verdict;
  runtime_ms: number;
  stdout?: string;
  stderr?: string;
  exit_code: number;
  timed_out: boolean;
}

export interface TraceEvent {
  id?: number;
  sequence_no: number;
  step_type: string;
  title?: string;
  content: string;
  metadata?: string;
  created_at: string;
}

export interface ExperimentRunTrace {
  experiment_run_id: number;
  timeline: TraceEvent[];
}
