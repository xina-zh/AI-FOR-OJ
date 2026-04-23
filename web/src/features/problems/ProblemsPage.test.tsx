import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { ProblemsPage } from './ProblemsPage';
import { mockFetchRoutes, renderWithProviders } from '../../test/render';

describe('ProblemsPage', () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it('renders problems returned by the backend', async () => {
    mockFetchRoutes({
      '/api/v1/problems': [
        {
          id: 1,
          title: 'A+B',
          description: 'sum two numbers',
          input_spec: 'a b',
          output_spec: 'sum',
          samples: '1 2 -> 3',
          time_limit_ms: 1000,
          memory_limit_mb: 128,
          difficulty: 'easy',
          tags: 'math',
        },
      ],
    });

    renderWithProviders(<ProblemsPage />);

    expect(await screen.findByRole('heading', { name: '题目' })).toBeInTheDocument();
    expect(await screen.findByText('A+B')).toBeInTheDocument();
    expect(screen.getByText('easy')).toBeInTheDocument();
  });

  it('confirms and permanently deletes a problem', async () => {
    const fetchMock = vi.fn((input: RequestInfo | URL, init?: RequestInit) => {
      const url = typeof input === 'string' ? input : input.toString();
      if (url === '/api/v1/problems' && !init?.method) {
        return Promise.resolve(
          new Response(
            JSON.stringify([
              {
                id: 1,
                title: 'A+B',
                description: 'sum two numbers',
                input_spec: 'a b',
                output_spec: 'sum',
                samples: '1 2 -> 3',
                time_limit_ms: 1000,
                memory_limit_mb: 128,
                difficulty: 'easy',
                tags: 'math',
              },
            ]),
            { status: 200, headers: { 'Content-Type': 'application/json' } },
          ),
        );
      }
      if (url === '/api/v1/problems/1' && init?.method === 'DELETE') {
        return Promise.resolve(new Response(null, { status: 204 }));
      }
      return Promise.resolve(
        new Response(JSON.stringify({ error: `unhandled request: ${url}` }), {
          status: 500,
          headers: { 'Content-Type': 'application/json' },
        }),
      );
    });
    vi.stubGlobal('fetch', fetchMock);
    vi.spyOn(window, 'confirm').mockReturnValue(true);

    renderWithProviders(<ProblemsPage />);

    fireEvent.click(await screen.findByRole('button', { name: '永久删除 A+B' }));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith('/api/v1/problems/1', expect.objectContaining({ method: 'DELETE' }));
    });
  });

  it('does not delete when confirmation is cancelled', async () => {
    mockFetchRoutes({
      '/api/v1/problems': [
        {
          id: 1,
          title: 'A+B',
          description: 'sum two numbers',
          input_spec: 'a b',
          output_spec: 'sum',
          samples: '1 2 -> 3',
          time_limit_ms: 1000,
          memory_limit_mb: 128,
          difficulty: 'easy',
          tags: 'math',
        },
      ],
    });
    const confirmSpy = vi.spyOn(window, 'confirm').mockReturnValue(false);

    renderWithProviders(<ProblemsPage />);

    fireEvent.click(await screen.findByRole('button', { name: '永久删除 A+B' }));

    expect(confirmSpy).toHaveBeenCalled();
    expect(fetch).not.toHaveBeenCalledWith('/api/v1/problems/1', expect.objectContaining({ method: 'DELETE' }));
  });
});
