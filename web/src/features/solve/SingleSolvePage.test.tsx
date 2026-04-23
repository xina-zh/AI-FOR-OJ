import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { SingleSolvePage } from './SingleSolvePage';
import { experimentOptionsFixture } from '../../test/fixtures';
import { renderWithProviders } from '../../test/render';

describe('SingleSolvePage', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('submits solve variables and renders the result', async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url === '/api/v1/meta/experiment-options') {
        return Promise.resolve(Response.json(experimentOptionsFixture));
      }
      if (url === '/api/v1/problems') {
        return Promise.resolve(
          Response.json([
            {
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
            },
          ]),
        );
      }
      if (url === '/api/v1/ai/solve') {
        return Promise.resolve(
          Response.json(
            {
              ai_solve_run_id: 9,
              problem_id: 5,
              model: 'mock-cpp17',
              prompt_name: 'default',
              agent_name: 'direct_codegen',
              prompt_preview: 'solve echo',
              raw_response: '```cpp\nint main(){return 0;}\n```',
              extracted_code: 'int main(){return 0;}',
              submission_id: 11,
              verdict: 'AC',
              token_input: 12,
              token_output: 8,
              llm_latency_ms: 30,
              total_latency_ms: 60,
            },
            { status: 201 },
          ),
        );
      }
      return Promise.resolve(Response.json({ error: `unhandled ${url}` }, { status: 500 }));
    });
    vi.stubGlobal('fetch', fetchMock);

    renderWithProviders(<SingleSolvePage />);

    expect(await screen.findByText('5 · Echo')).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('题目'), { target: { value: '5' } });
    fireEvent.change(await screen.findByLabelText('Tooling'), { target: { value: 'sample_judge' } });
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '执行 Solve' })).not.toBeDisabled();
    });
    fireEvent.click(screen.getByRole('button', { name: '执行 Solve' }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        '/api/v1/ai/solve',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({
            problem_id: 5,
            model: 'mock-cpp17',
            prompt_name: 'default',
            agent_name: 'direct_codegen',
            tooling_config: 'sample_judge',
          }),
        }),
      );
    });
    expect(await screen.findByText('AC')).toBeInTheDocument();
    expect(screen.getByText(/Total 20/)).toBeInTheDocument();
  });
});
