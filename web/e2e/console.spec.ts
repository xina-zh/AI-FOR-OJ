import { expect, test, type Page } from '@playwright/test';

const now = '2026-04-22T00:00:00Z';

test.beforeEach(async ({ page }) => {
  await mockApi(page);
});

test('loads the dashboard with experiment totals', async ({ page }) => {
  await page.goto('/');

  await expect(page.getByRole('heading', { name: 'Experiment Dashboard' })).toBeVisible();
  await expect(page.getByText('Experiments')).toBeVisible();
  await expect(page.getByText('Compares')).toBeVisible();
  await expect(page.getByText('Repeats')).toBeVisible();
});

test('renders options metadata in the solve form', async ({ page }) => {
  await page.goto('/solve');

  await expect(page.getByRole('heading', { name: 'Single Solve' })).toBeVisible();
  await expect(page.getByLabel('Model')).toHaveValue('mock-cpp17');
  await expect(page.getByLabel('Prompt')).toHaveValue('strict_cpp17');
  await expect(page.getByLabel('Agent')).toHaveValue('tooling_codegen_v1');
  await expect(page.getByLabel('Tooling JSON')).toHaveValue('{}');
});

test('renders the problem list from the mocked API', async ({ page }) => {
  await page.goto('/problems');

  await expect(page.getByRole('heading', { name: 'Problems' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'A+B Problem' })).toBeVisible();
  await expect(page.getByText('1000ms / 128MB')).toBeVisible();
});

test('submits the solve form with mocked API data', async ({ page }) => {
  await page.goto('/solve');

  await page.getByLabel('Problem ID').fill('1');
  await page.getByRole('button', { name: 'Run solve' }).click();

  await expect(page.getByText('Submission 99')).toBeVisible();
  await expect(page.getByText('initial_codegen')).toBeVisible();
  await expect(page.getByText('int main')).toBeVisible();
});

test('renders compare detail baseline and candidate summaries', async ({ page }) => {
  await page.goto('/compare/12');

  await expect(page.getByRole('heading', { name: 'e2e-compare' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Baseline' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Candidate' })).toBeVisible();
  await expect(page.getByText('regressed')).toBeVisible();
});

async function mockApi(page: Page) {
  await page.route('**/api/v1/meta/experiment-options', async (route) => {
    await route.fulfill({
      json: {
        default_model: 'mock-cpp17',
        models: [{ name: 'mock-cpp17', label: 'mock-cpp17' }],
        prompts: [{ name: 'strict_cpp17', label: 'strict_cpp17' }],
        agents: [{ name: 'tooling_codegen_v1', label: 'tooling_codegen_v1' }],
        tooling_options: [{ name: 'sample_judge', label: 'sample_judge' }],
      },
    });
  });
  await page.route('**/api/v1/problems', async (route) => {
    await route.fulfill({ json: problems });
  });
  await page.route('**/api/v1/ai/solve', async (route) => {
    await route.fulfill({ json: solveRun });
  });
  await page.route('**/api/v1/experiments?page=1&page_size=20', async (route) => {
    await route.fulfill({ json: pageOf([experiment]) });
  });
  await page.route('**/api/v1/experiments/compare?page=1&page_size=20', async (route) => {
    await route.fulfill({ json: pageOf([compare]) });
  });
  await page.route('**/api/v1/experiments/repeat?page=1&page_size=20', async (route) => {
    await route.fulfill({ json: pageOf([repeat]) });
  });
  await page.route('**/api/v1/experiments/compare/12', async (route) => {
    await route.fulfill({ json: compare });
  });
}

function pageOf<T>(items: T[]) {
  return { items, page: 1, page_size: 20, total: items.length, total_pages: 1 };
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
    created_at: now,
  },
];

const costSummary = {
  total_token_input: 10,
  total_token_output: 20,
  total_tokens: 30,
  total_llm_latency_ms: 40,
  total_latency_ms: 50,
  run_count: 1,
};

const solveRun = {
  id: 88,
  ai_solve_run_id: 88,
  problem_id: 1,
  model: 'mock-cpp17',
  prompt_name: 'strict_cpp17',
  agent_name: 'tooling_codegen_v1',
  prompt_preview: 'Solve A+B',
  raw_response: '```cpp\nint main() { return 0; }\n```',
  extracted_code: 'int main() { return 0; }',
  tooling_config: '{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}',
  tool_call_count: 1,
  submission_id: 99,
  verdict: 'AC',
  status: 'success',
  attempt_count: 1,
  token_input: 11,
  token_output: 22,
  llm_latency_ms: 33,
  total_latency_ms: 44,
  attempts: [
    {
      id: 1,
      attempt_no: 1,
      stage: 'initial_codegen',
      model: 'mock-cpp17',
      verdict: 'AC',
      failure_type: '',
      repair_reason: '',
      token_input: 11,
      token_output: 22,
      llm_latency_ms: 33,
      total_latency_ms: 44,
    },
  ],
  created_at: now,
};

const experiment = {
  id: 5,
  name: 'e2e-batch',
  model: 'mock-cpp17',
  prompt_name: 'strict_cpp17',
  agent_name: 'tooling_codegen_v1',
  tooling_config: '{}',
  status: 'completed',
  total_count: 1,
  success_count: 1,
  ac_count: 1,
  failed_count: 0,
  verdict_distribution: { ac_count: 1 },
  cost_summary: costSummary,
  created_at: now,
  updated_at: now,
  runs: [],
};

const compare = {
  id: 12,
  name: 'e2e-compare',
  compare_dimension: 'agent',
  baseline_value: 'direct_codegen',
  candidate_value: 'tooling_codegen_v1',
  baseline_prompt_name: 'strict_cpp17',
  candidate_prompt_name: 'strict_cpp17',
  baseline_agent_name: 'direct_codegen',
  candidate_agent_name: 'tooling_codegen_v1',
  baseline_tooling_config: '{}',
  candidate_tooling_config: '{"enabled":["sample_judge"],"max_calls":1,"per_tool_max_calls":{}}',
  baseline_experiment_id: 5,
  candidate_experiment_id: 6,
  baseline_summary: experiment,
  candidate_summary: { ...experiment, id: 6, agent_name: 'tooling_codegen_v1' },
  baseline_verdict_distribution: { ac_count: 1 },
  candidate_verdict_distribution: { wa_count: 1 },
  highlighted_problems: [
    { problem_id: 1, baseline_verdict: 'AC', candidate_verdict: 'WA', changed: true, change_type: 'regressed' },
  ],
  status: 'completed',
  created_at: now,
  updated_at: now,
};

const repeat = {
  id: 20,
  name: 'e2e-repeat',
  model: 'mock-cpp17',
  prompt_name: 'strict_cpp17',
  agent_name: 'tooling_codegen_v1',
  tooling_config: '{}',
  problem_ids: [1],
  repeat_count: 2,
  experiment_ids: [5, 6],
  total_problem_count: 1,
  total_run_count: 2,
  overall_ac_count: 1,
  overall_failed_count: 1,
  overall_ac_rate: 0.5,
  cost_summary: costSummary,
  status: 'completed',
  created_at: now,
  updated_at: now,
};
