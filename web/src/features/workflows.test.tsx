import '@testing-library/jest-dom/vitest';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { App } from '../App';

function renderApp(path: string) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[path]}>
        <App />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('console workflows', () => {
  beforeEach(() => {
    vi.stubGlobal(
      'fetch',
      vi.fn(async (input: RequestInfo | URL) => {
        const url = String(input);
        if (url.includes('/meta/experiment-options')) {
          return new Response(
            JSON.stringify({
              default_model: 'mock-model',
              models: [{ name: 'mock-model', label: 'mock-model' }],
              prompts: [{ name: 'default', label: 'default' }],
              agents: [{ name: 'direct_codegen', label: 'direct_codegen' }],
              tooling_options: [{ name: 'sample_judge', label: 'sample_judge' }],
            }),
            { status: 200, headers: { 'Content-Type': 'application/json' } },
          );
        }
        if (url.includes('/experiment-runs/9/trace')) {
          return new Response(
            JSON.stringify({
              experiment_run_id: 9,
              experiment_id: 3,
              problem_id: 2,
              status: 'success',
              timeline: [{ id: 1, sequence_no: 1, step_type: 'llm', content: 'prompt', metadata: '{}', created_at: '2026-04-22T00:00:00Z' }],
            }),
            { status: 200, headers: { 'Content-Type': 'application/json' } },
          );
        }
        return new Response(JSON.stringify({ items: [], page: 1, page_size: 20, total: 0, total_pages: 0 }), {
          status: 200,
          headers: { 'Content-Type': 'application/json' },
        });
      }),
    );
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('renders a usable solve form', async () => {
    renderApp('/solve');

    expect(await screen.findByLabelText('Problem ID')).toBeInTheDocument();
    expect(screen.getByLabelText('Model')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Run solve' })).toBeInTheDocument();
  });

  it('renders an experiment run trace timeline', async () => {
    renderApp('/trace/experiment-runs/9');

    expect(await screen.findByRole('heading', { name: 'Experiment Run Trace' })).toBeInTheDocument();
    expect(screen.getByText('llm')).toBeInTheDocument();
    expect(screen.getByText('prompt')).toBeInTheDocument();
  });
});
