import { expect, test, type Page } from '@playwright/test';

const now = '2026-04-20T08:00:00Z';

test.beforeEach(async ({ page }) => {
  await mockApi(page);
});

test('loads dashboard with health and recent runs', async ({ page }) => {
  await page.goto('/');

  await expect(page.getByRole('heading', { name: '实验控制台' })).toBeVisible();
  await expect(page.getByText('ok')).toBeVisible();
  await expect(page.getByText('题目总数')).toBeVisible();
  await expect(page.getByRole('link', { name: 'e2e-batch' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'e2e-compare' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'e2e-repeat' })).toBeVisible();
});

test('runs a single solve through mocked API', async ({ page }) => {
  await page.goto('/solve');

  await page.getByLabel('题目').selectOption('1');
  await page.getByRole('button', { name: '执行 Solve' }).click();

  await expect(page.getByRole('heading', { name: '运行结果' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'AI Run #88' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Submission #99' })).toBeVisible();
  await expect(page.getByText('Prompt Preview')).toBeVisible();
  await expect(page.getByText('int main')).toBeVisible();
});

test('runs compare through mocked API', async ({ page }) => {
  await page.goto('/compare');

  await page.getByLabel('实验名称').fill('e2e-compare-run');
  await page.getByLabel('1 · A+B Problem').check();
  await page.getByRole('button', { name: '执行 Compare' }).click();

  await expect(page.getByText('Compare #12')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'e2e-compare-run' })).toBeVisible();
  await expect(page.getByText('Candidate direct_codegen_repair')).toBeVisible();
  await expect(page.getByText('regressed')).toBeVisible();
});

test('opens trace playback from mocked API', async ({ page }) => {
  await page.goto('/trace/experiment-runs/7');

  await expect(page.getByRole('heading', { name: 'Trace #7' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Prompt' })).toBeVisible();
  await expect(page.getByText('Solve A+B')).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Extracted Code' })).toBeVisible();
  await expect(page.getByText('#include <bits/stdc++.h>')).toBeVisible();
});

async function mockApi(page: Page) {
  await page.route('**/health', async (route) => {
    await route.fulfill({ json: { status: 'ok', app: 'ai-for-oj', env: 'test', database: 'ok' } });
  });
  await page.route('**/api/v1/meta/experiment-options', async (route) => {
    await route.fulfill({
      json: {
        default_model: 'mock-cpp17',
        prompts: [
          { name: 'default', label: 'default' },
          { name: 'strict_cpp17', label: 'strict_cpp17' },
        ],
        agents: [
          { name: 'direct_codegen', label: 'direct_codegen' },
          { name: 'direct_codegen_repair', label: 'direct_codegen_repair' },
        ],
      },
    });
  });
  await page.route('**/api/v1/problems', async (route) => {
    await route.fulfill({ json: problems });
  });
  await page.route('**/api/v1/submissions/stats/problems', async (route) => {
    await route.fulfill({
      json: [
        { problem_id: 1, problem_title: 'A+B Problem', total_submissions: 3, ac_count: 2, wa_count: 1, latest_submission_at: now },
      ],
    });
  });
  await page.route('**/api/v1/experiments?**', async (route) => {
    await route.fulfill({ json: pageOf([experiment]) });
  });
  await page.route('**/api/v1/experiments/compare?**', async (route) => {
    await route.fulfill({ json: pageOf([compare]) });
  });
  await page.route('**/api/v1/experiments/repeat?**', async (route) => {
    await route.fulfill({ json: pageOf([repeat]) });
  });
  await page.route('**/api/v1/ai/solve', async (route) => {
    await route.fulfill({
      json: {
        ai_solve_run_id: 88,
        problem_id: 1,
        model: 'mock-cpp17',
        prompt_name: 'default',
        agent_name: 'direct_codegen',
        prompt_preview: 'Solve A+B',
        raw_response: '```cpp\nint main() { return 0; }\n```',
        extracted_code: 'int main() { return 0; }',
        submission_id: 99,
        verdict: 'AC',
        token_input: 11,
        token_output: 22,
        llm_latency_ms: 33,
        total_latency_ms: 44,
      },
    });
  });
  await page.route('**/api/v1/experiments/compare', async (route) => {
    await route.fulfill({ json: { ...compare, name: 'e2e-compare-run' } });
  });
  await page.route('**/api/v1/experiment-runs/7/trace', async (route) => {
    await route.fulfill({
      json: {
        experiment_run_id: 7,
        events: [
          { sequence_no: 1, step_type: 'prompt', title: 'Prompt', content: 'Solve A+B', created_at: now },
          {
            sequence_no: 2,
            step_type: 'extracted_code',
            title: 'Extracted Code',
            content: '#include <bits/stdc++.h>\nint main() { return 0; }',
            created_at: now,
          },
        ],
      },
    });
  });
}

const problems = [
  {
    id: 1,
    title: 'A+B Problem',
    description: 'Add two numbers.',
    input_spec: 'a b',
    output_spec: 'a+b',
    samples: '1 2 -> 3',
    time_limit_ms: 1000,
    memory_limit_mb: 128,
    difficulty: 'easy',
    tags: 'math',
  },
];

const costSummary = {
  total_token_input: 10,
  total_token_output: 20,
  total_tokens: 30,
  average_token_input: 10,
  average_token_output: 20,
  average_total_tokens: 30,
  total_llm_latency_ms: 40,
  total_latency_ms: 50,
  average_llm_latency_ms: 40,
  average_total_latency_ms: 50,
  run_count: 1,
};

const experiment = {
  id: 5,
  name: 'e2e-batch',
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
  created_at: now,
  updated_at: now,
  runs: [
    { id: 7, problem_id: 1, ai_solve_run_id: 88, submission_id: 99, attempt_no: 1, verdict: 'AC', status: 'completed', created_at: now },
  ],
};

const compare = {
  id: 12,
  name: 'e2e-compare',
  compare_dimension: 'agent',
  baseline_value: 'direct_codegen',
  candidate_value: 'direct_codegen_repair',
  baseline_prompt_name: 'default',
  candidate_prompt_name: 'default',
  baseline_agent_name: 'direct_codegen',
  candidate_agent_name: 'direct_codegen_repair',
  problem_ids: [1],
  baseline_experiment_id: 5,
  candidate_experiment_id: 6,
  baseline_verdict_distribution: { ac_count: 1 },
  candidate_verdict_distribution: { wa_count: 1 },
  delta_verdict_distribution: { ac_count: -1, wa_count: 1 },
  cost_comparison: {
    baseline_total_tokens: 30,
    candidate_total_tokens: 45,
    delta_total_tokens: 15,
    baseline_average_total_latency_ms: 50,
    candidate_average_total_latency_ms: 60,
  },
  comparison_summary: { tradeoff_type: 'quality_regression' },
  improved_count: 0,
  regressed_count: 1,
  changed_non_ac_count: 0,
  problem_summaries: [
    { problem_id: 1, baseline_verdict: 'AC', candidate_verdict: 'WA', changed: true, change_type: 'regressed' },
  ],
  highlighted_problems: [
    { problem_id: 1, baseline_verdict: 'AC', candidate_verdict: 'WA', changed: true, change_type: 'regressed' },
  ],
  delta_ac_count: -1,
  delta_failed_count: 1,
  status: 'completed',
  created_at: now,
  updated_at: now,
};

const repeat = {
  id: 20,
  name: 'e2e-repeat',
  model: 'mock-cpp17',
  prompt_name: 'default',
  agent_name: 'direct_codegen',
  problem_ids: [1],
  repeat_count: 2,
  experiment_ids: [5, 6],
  total_problem_count: 1,
  total_run_count: 2,
  overall_ac_count: 1,
  overall_failed_count: 1,
  average_ac_count_per_round: 0.5,
  average_failed_count_per_round: 0.5,
  overall_ac_rate: 0.5,
  best_round_ac_count: 1,
  worst_round_ac_count: 0,
  cost_summary: { ...costSummary, run_count: 2 },
  status: 'completed',
  round_summaries: [],
  problem_summaries: [],
  most_unstable_problems: [],
  created_at: now,
  updated_at: now,
};

function pageOf<T>(items: T[]) {
  return {
    items,
    page: 1,
    page_size: 20,
    total: items.length,
    total_pages: 1,
  };
}
