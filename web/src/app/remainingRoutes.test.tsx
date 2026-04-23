import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { App } from './App';
import { router } from './router';

const costSummary = {
  total_token_input: 120,
  total_token_output: 80,
  total_tokens: 200,
  average_token_input: 120,
  average_token_output: 80,
  average_total_tokens: 200,
  total_llm_latency_ms: 500,
  total_latency_ms: 900,
  average_llm_latency_ms: 500,
  average_total_latency_ms: 900,
  run_count: 1,
};

const experiment = {
  id: 31,
  name: 'batch-ui',
  model: 'mock-cpp17',
  prompt_name: 'default',
  agent_name: 'direct_codegen',
  status: 'completed',
  total_count: 1,
  success_count: 1,
  ac_count: 1,
  failed_count: 0,
  verdict_distribution: { ac_count: 1 },
  cost_summary: costSummary,
  created_at: '2026-04-20T10:00:00Z',
  updated_at: '2026-04-20T10:01:00Z',
  runs: [
    {
      id: 99,
      problem_id: 5,
      ai_solve_run_id: 77,
      submission_id: 88,
      attempt_no: 1,
      verdict: 'AC',
      status: 'completed',
      created_at: '2026-04-20T10:00:30Z',
    },
  ],
};

const compare = {
  id: 41,
  name: 'candidate-cost',
  compare_dimension: 'model',
  baseline_value: 'mock-a',
  candidate_value: 'mock-b',
  baseline_prompt_name: 'default',
  candidate_prompt_name: 'default',
  baseline_agent_name: 'direct_codegen',
  candidate_agent_name: 'direct_codegen',
  problem_ids: [5],
  baseline_experiment_id: 31,
  candidate_experiment_id: 32,
  baseline_verdict_distribution: { ac_count: 1 },
  candidate_verdict_distribution: { ac_count: 1 },
  delta_verdict_distribution: {},
  cost_comparison: {
    baseline_total_tokens: 200,
    candidate_total_tokens: 260,
    delta_total_tokens: 60,
    baseline_average_total_latency_ms: 900,
    candidate_average_total_latency_ms: 980,
    delta_average_total_latency_ms: 80,
  },
  comparison_summary: { tradeoff_type: 'same_accuracy_more_cost' },
  improved_count: 0,
  regressed_count: 0,
  changed_non_ac_count: 0,
  problem_summaries: [],
  highlighted_problems: [],
  delta_ac_count: 0,
  delta_failed_count: 0,
  status: 'completed',
  created_at: '2026-04-20T10:00:00Z',
  updated_at: '2026-04-20T10:03:00Z',
};

const repeat = {
  id: 51,
  name: 'repeat-stability',
  model: 'mock-cpp17',
  prompt_name: 'default',
  agent_name: 'direct_codegen',
  problem_ids: [5],
  repeat_count: 2,
  experiment_ids: [31, 32],
  total_problem_count: 1,
  total_run_count: 2,
  overall_ac_count: 2,
  overall_failed_count: 0,
  average_ac_count_per_round: 1,
  average_failed_count_per_round: 0,
  overall_ac_rate: 1,
  best_round_ac_count: 1,
  worst_round_ac_count: 1,
  cost_summary: { ...costSummary, total_token_input: 240, total_token_output: 160, total_tokens: 400, run_count: 2 },
  status: 'completed',
  round_summaries: [],
  problem_summaries: [],
  most_unstable_problems: [],
  created_at: '2026-04-20T10:00:00Z',
  updated_at: '2026-04-20T10:04:00Z',
};

const submission = {
  id: 88,
  problem_id: 5,
  problem_title: 'Echo',
  language: 'cpp17',
  source_type: 'ai',
  verdict: 'AC',
  runtime_ms: 12,
  passed_count: 2,
  total_count: 2,
  created_at: '2026-04-20T10:00:00Z',
  updated_at: '2026-04-20T10:00:01Z',
};

describe('remaining routes', () => {
  afterEach(() => {
    cleanup();
    void router.navigate('/');
    vi.unstubAllGlobals();
  });

  it('renders token analytics from experiment, compare, and repeat history', async () => {
    stubFetch({
      '/api/v1/experiments?page_size=20': { items: [experiment], page: 1, page_size: 20, total: 1, total_pages: 1 },
      '/api/v1/experiments/compare?page_size=20': { items: [compare], page: 1, page_size: 20, total: 1, total_pages: 1 },
      '/api/v1/experiments/repeat?page_size=20': { items: [repeat], page: 1, page_size: 20, total: 1, total_pages: 1 },
    });

    await router.navigate('/tokens');
    render(<App />);

    expect(await screen.findByRole('heading', { name: 'Token 分析' })).toBeInTheDocument();
    expect(await screen.findByText('batch-ui')).toBeInTheDocument();
    expect(await screen.findByText('candidate-cost')).toBeInTheDocument();
    expect(screen.getByText('+60 tokens')).toBeInTheDocument();
    expect(await screen.findByText('repeat-stability')).toBeInTheDocument();
  });

  it('renders an experiment run trace timeline', async () => {
    stubFetch({
      '/api/v1/experiment-runs/99/trace': {
        experiment_run_id: 99,
        events: [
          { sequence_no: 1, step_type: 'prompt', title: 'Prompt', content: 'solve echo', created_at: '2026-04-20T10:00:00Z' },
          { sequence_no: 2, step_type: 'extracted_code', title: 'Extracted Code', content: 'int main(){}', created_at: '2026-04-20T10:00:01Z' },
        ],
      },
    });

    await router.navigate('/trace/experiment-runs/99');
    render(<App />);

    expect(await screen.findByRole('heading', { name: 'Trace #99' })).toBeInTheDocument();
    expect(await screen.findByText('Prompt')).toBeInTheDocument();
    expect(screen.getByText('solve echo')).toBeInTheDocument();
    expect(screen.getByText('int main(){}')).toBeInTheDocument();
  });

  it('renders submission list and submission detail pages', async () => {
    stubFetch({
      '/api/v1/submissions?page_size=20': { items: [submission], page: 1, page_size: 20, total: 1, total_pages: 1 },
      '/api/v1/submissions/88': {
        ...submission,
        source_code: '#include <bits/stdc++.h>\nint main(){}',
        memory_kb: 1024,
        compile_stderr: '',
        run_stdout: 'ok',
        run_stderr: '',
        exit_code: 0,
        timed_out: false,
        exec_stage: 'run',
        judge_result: { verdict: 'AC', runtime_ms: 12, memory_kb: 1024, passed_count: 2, total_count: 2 },
        testcase_results: [{ testcase_id: 1, index: 1, verdict: 'AC', runtime_ms: 5, stdout: 'ok', exit_code: 0, timed_out: false }],
      },
    });

    await router.navigate('/submissions');
    render(<App />);

    expect(await screen.findByRole('heading', { name: '提交记录' })).toBeInTheDocument();
    expect(await screen.findByText('Echo')).toBeInTheDocument();

    await router.navigate('/submissions/88');

    expect(await screen.findByRole('heading', { name: 'Submission #88' })).toBeInTheDocument();
    expect(screen.getByText(/#include <bits\/stdc\+\+\.h>/)).toBeInTheDocument();
    expect(screen.getAllByText('ok').length).toBeGreaterThan(0);
  });

  it('renders experiment, compare, and repeat detail routes', async () => {
    stubFetch({
      '/api/v1/experiments/31': experiment,
      '/api/v1/experiments/compare/41': compare,
      '/api/v1/experiments/repeat/51': repeat,
    });

    await router.navigate('/experiments/31');
    render(<App />);

    expect(await screen.findByRole('heading', { level: 1, name: 'batch-ui' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Trace #99' })).toHaveAttribute('href', '/trace/experiment-runs/99');
    expect(screen.getByRole('link', { name: 'Submission #88' })).toHaveAttribute('href', '/submissions/88');

    await router.navigate('/compare/41');
    expect(await screen.findByRole('heading', { level: 1, name: 'candidate-cost' })).toBeInTheDocument();
    expect(screen.getByText('same_accuracy_more_cost')).toBeInTheDocument();

    await router.navigate('/repeat/51');
    expect(await screen.findByRole('heading', { level: 1, name: 'repeat-stability' })).toBeInTheDocument();
    expect(screen.getByText('整体 AC 率 100.0%')).toBeInTheDocument();
  });

  it('renders dashboard totals and recent run lists', async () => {
    stubFetch({
      '/health': { status: 'ok', app: 'ai-for-oj', env: 'test', database: 'ok' },
      '/api/v1/problems': [{ id: 5, title: 'Echo' }],
      '/api/v1/submissions/stats/problems': [{ problem_id: 5, problem_title: 'Echo', total_submissions: 3, ac_count: 2 }],
      '/api/v1/experiments?page_size=5': { items: [experiment], page: 1, page_size: 5, total: 1, total_pages: 1 },
      '/api/v1/experiments/compare?page_size=5': { items: [compare], page: 1, page_size: 5, total: 1, total_pages: 1 },
      '/api/v1/experiments/repeat?page_size=5': { items: [repeat], page: 1, page_size: 5, total: 1, total_pages: 1 },
    });

    await router.navigate('/');
    render(<App />);

    expect(await screen.findByText('batch-ui')).toBeInTheDocument();
    expect(screen.getByText('题目总数')).toBeInTheDocument();
    expect(screen.getAllByText('1').length).toBeGreaterThan(0);
    expect(screen.getByText('Submission 总数')).toBeInTheDocument();
    expect(screen.getByText('3')).toBeInTheDocument();
    expect(await screen.findByText('batch-ui')).toBeInTheDocument();
    expect(screen.getByText('candidate-cost')).toBeInTheDocument();
    expect(screen.getByText('repeat-stability')).toBeInTheDocument();
  });
});

function stubFetch(routes: Record<string, unknown>) {
  const fetchMock = vi.fn((input: RequestInfo | URL) => {
    const url = typeof input === 'string' ? input : input.toString();
    const response = routes[url];
    if (response === undefined) {
      return Promise.resolve(Response.json({ error: `unhandled ${url}` }, { status: 500 }));
    }
    return Promise.resolve(Response.json(response));
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}
