import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { cleanup, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { MemoryRouter, Route, Routes } from 'react-router-dom';

import { ProblemDetail } from './ProblemDetail';
import { mockFetchRoutes } from '../../test/render';

function renderProblemDetail(problemId: number) {
  const client = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter initialEntries={[`/problems/${problemId}`]}>
        <Routes>
          <Route path="/problems/:id" element={<ProblemDetail />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('ProblemDetail', () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it('renders JSON samples as separate input and output blocks', async () => {
    mockFetchRoutes({
      '/api/v1/problems/35': {
        id: 35,
        title: 'Find B with C in A',
        description: 'description',
        input_spec: 'input spec',
        output_spec: 'output spec',
        samples: JSON.stringify([{ input: '2\n5\nabcde a', output: '1 2 4\n1 2 5' }]),
        time_limit_ms: 1000,
        memory_limit_mb: 512,
        difficulty: 'unknown',
        tags: '',
      },
      '/api/v1/problems/35/testcases': [],
    });

    renderProblemDetail(35);

    expect(await screen.findByRole('heading', { name: '样例输入 1' })).toBeInTheDocument();
    expect(screen.getByText((_, element) => element?.tagName === 'PRE' && element.textContent === '2\n5\nabcde a')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: '样例输出 1' })).toBeInTheDocument();
    expect(screen.getByText((_, element) => element?.tagName === 'PRE' && element.textContent === '1 2 4\n1 2 5')).toBeInTheDocument();
    expect(screen.queryByText(/\[\{"input"/)).not.toBeInTheDocument();
  });
});
