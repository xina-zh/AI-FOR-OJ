import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { App } from './App';
import { router } from './router';
import { experimentOptionsFixture } from '../test/fixtures';

const problem = {
  id: 5,
  title: 'Echo',
  description: 'echo input',
  input_spec: 'text',
  output_spec: 'text',
  samples: 'hello',
  time_limit_ms: 1000,
  memory_limit_mb: 128,
  difficulty: 'easy',
  tags: 'io',
};

const costSummary = {
  total_token_input: 120,
  total_token_output: 80,
  total_tokens: 200,
  average_token_input: 60,
  average_token_output: 40,
  average_total_tokens: 100,
  total_llm_latency_ms: 500,
  total_latency_ms: 800,
  average_llm_latency_ms: 250,
  average_total_latency_ms: 400,
  run_count: 2,
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
  name: 'prompt-compare',
  compare_dimension: 'prompt',
  baseline_value: 'default',
  candidate_value: 'strict_cpp17',
  baseline_prompt_name: 'default',
  candidate_prompt_name: 'strict_cpp17',
  baseline_agent_name: 'direct_codegen',
  candidate_agent_name: 'direct_codegen',
  problem_ids: [5],
  baseline_experiment_id: 31,
  candidate_experiment_id: 32,
  baseline_verdict_distribution: { ac_count: 0, wa_count: 1 },
  candidate_verdict_distribution: { ac_count: 1 },
  delta_verdict_distribution: { ac_count: 1, wa_count: -1 },
  cost_comparison: {
    baseline_total_tokens: 200,
    candidate_total_tokens: 240,
    delta_total_tokens: 40,
    baseline_average_total_latency_ms: 400,
    candidate_average_total_latency_ms: 460,
  },
  comparison_summary: {
    accuracy_winner: 'candidate',
    cost_winner: 'baseline',
    latency_winner: 'baseline',
    tradeoff_type: 'accuracy_for_cost',
  },
  improved_count: 1,
  regressed_count: 0,
  changed_non_ac_count: 0,
  problem_summaries: [
    {
      problem_id: 5,
      baseline_verdict: 'WA',
      candidate_verdict: 'AC',
      changed: true,
      change_type: 'improved',
    },
  ],
  highlighted_problems: [
    {
      problem_id: 5,
      baseline_verdict: 'WA',
      candidate_verdict: 'AC',
      changed: true,
      change_type: 'improved',
    },
  ],
  delta_ac_count: 1,
  delta_failed_count: -1,
  status: 'completed',
  created_at: '2026-04-20T10:00:00Z',
  updated_at: '2026-04-20T10:03:00Z',
};

const repeat = {
  id: 51,
  name: 'stability-repeat',
  model: 'mock-cpp17',
  prompt_name: 'default',
  agent_name: 'direct_codegen',
  problem_ids: [5],
  repeat_count: 3,
  experiment_ids: [31, 32, 33],
  total_problem_count: 1,
  total_run_count: 3,
  overall_ac_count: 2,
  overall_failed_count: 1,
  average_ac_count_per_round: 0.67,
  average_failed_count_per_round: 0.33,
  overall_ac_rate: 0.667,
  best_round_ac_count: 1,
  worst_round_ac_count: 0,
  cost_summary: {
    total_token_input: 300,
    total_token_output: 150,
    total_tokens: 450,
    average_token_input: 100,
    average_token_output: 50,
    average_total_tokens: 150,
    total_llm_latency_ms: 900,
    total_latency_ms: 1200,
    average_llm_latency_ms: 300,
    average_total_latency_ms: 400,
    run_count: 3,
  },
  status: 'completed',
  round_summaries: [
    { round_no: 1, experiment_id: 31, ac_count: 1, failed_count: 0, verdict_distribution: { ac_count: 1 }, status: 'completed' },
    { round_no: 2, experiment_id: 32, ac_count: 1, failed_count: 0, verdict_distribution: { ac_count: 1 }, status: 'completed' },
    { round_no: 3, experiment_id: 33, ac_count: 0, failed_count: 1, verdict_distribution: { wa_count: 1 }, status: 'completed' },
  ],
  problem_summaries: [
    {
      problem_id: 5,
      total_rounds: 3,
      ac_count: 2,
      failed_count: 1,
      ac_rate: 0.667,
      verdict_distribution: { ac_count: 2, wa_count: 1 },
      latest_verdict: 'WA',
    },
  ],
  most_unstable_problems: [
    {
      problem_id: 5,
      total_rounds: 3,
      ac_count: 2,
      failed_count: 1,
      ac_rate: 0.667,
      verdict_distribution: { ac_count: 2, wa_count: 1 },
      latest_verdict: 'WA',
      instability_score: 2,
      verdict_kind_count: 2,
    },
  ],
  created_at: '2026-04-20T10:00:00Z',
  updated_at: '2026-04-20T10:04:00Z',
};

describe('experiment routes', () => {
  afterEach(() => {
    cleanup();
    void router.navigate('/');
    vi.unstubAllGlobals();
  });

  it('runs a batch experiment from the experiments page', async () => {
    const fetchMock = stubExperimentFetch({
      '/api/v1/experiments?page_size=20': { items: [experiment], page: 1, page_size: 20, total: 1, total_pages: 1 },
      '/api/v1/experiments/run': experiment,
    });

    await router.navigate('/experiments');
    render(<App />);

    expect(await screen.findByRole('heading', { name: '批量实验' })).toBeInTheDocument();
    expect(await screen.findByText('batch-ui')).toBeInTheDocument();
    fireEvent.click(await screen.findByRole('checkbox', { name: '5 · Echo' }));
    fireEvent.click(screen.getByRole('button', { name: '执行批量实验' }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        '/api/v1/experiments/run',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({
            name: '',
            problem_ids: [5],
            model: 'mock-cpp17',
            prompt_name: 'default',
            agent_name: 'direct_codegen',
          }),
        }),
      );
    });
    expect(await screen.findByText('Run #99')).toBeInTheDocument();
  });

  it('runs a compare experiment and renders its summary', async () => {
    const fetchMock = stubExperimentFetch({
      '/api/v1/experiments/compare?page_size=20': { items: [compare], page: 1, page_size: 20, total: 1, total_pages: 1 },
    });

    await router.navigate('/compare');
    render(<App />);

    expect(await screen.findByRole('heading', { name: '对比实验' })).toBeInTheDocument();
    expect(await screen.findByText('prompt-compare')).toBeInTheDocument();
    fireEvent.click(await screen.findByRole('checkbox', { name: '5 · Echo' }));
    fireEvent.change(screen.getByLabelText('Candidate Prompt'), { target: { value: 'strict_cpp17' } });
    fireEvent.click(screen.getByRole('button', { name: '执行 Compare' }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith('/api/v1/experiments/compare', expect.objectContaining({ method: 'POST' }));
    });
    expect((await screen.findAllByText('accuracy_for_cost')).length).toBeGreaterThan(0);
    expect(screen.getByText('improved')).toBeInTheDocument();
  });

  it('runs a repeat experiment and renders stability metrics', async () => {
    const fetchMock = stubExperimentFetch({
      '/api/v1/experiments/repeat?page_size=20': { items: [repeat], page: 1, page_size: 20, total: 1, total_pages: 1 },
    });

    await router.navigate('/repeat');
    render(<App />);

    expect(await screen.findByRole('heading', { name: '重复实验' })).toBeInTheDocument();
    expect(await screen.findByText('stability-repeat')).toBeInTheDocument();
    fireEvent.click(await screen.findByRole('checkbox', { name: '5 · Echo' }));
    fireEvent.change(screen.getByLabelText('重复次数'), { target: { value: '3' } });
    fireEvent.click(screen.getByRole('button', { name: '执行 Repeat' }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith('/api/v1/experiments/repeat', expect.objectContaining({ method: 'POST' }));
    });
    expect(await screen.findByText('整体 AC 率 66.7%')).toBeInTheDocument();
    expect(screen.getByText('Round 3')).toBeInTheDocument();
  });
});

function stubExperimentFetch(routes: Record<string, unknown>) {
  const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input.toString();
    if (url === '/api/v1/meta/experiment-options') {
      return Promise.resolve(Response.json(experimentOptionsFixture));
    }
    if (url === '/api/v1/problems') {
      return Promise.resolve(Response.json([problem]));
    }
    if (url === '/api/v1/experiments/compare' && fetchMethod(input, init) === 'POST') {
      return Promise.resolve(Response.json(compare, { status: 201 }));
    }
    if (url === '/api/v1/experiments/repeat' && fetchMethod(input, init) === 'POST') {
      return Promise.resolve(Response.json(repeat, { status: 201 }));
    }
    if (routes[url]) {
      return Promise.resolve(Response.json(routes[url]));
    }
    return Promise.resolve(Response.json({ error: `unhandled ${url}` }, { status: 500 }));
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}

function fetchMethod(input: RequestInfo | URL, init?: RequestInit) {
  if (init?.method) {
    return init.method;
  }
  if (input instanceof Request) {
    return input.method;
  }
  return 'GET';
}
