import { afterEach, describe, expect, it, vi } from 'vitest';

import { deleteProblem } from './problemApi';

describe('problemApi', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('sends DELETE for a problem', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
    vi.stubGlobal('fetch', fetchMock);

    await deleteProblem(42);

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/problems/42', expect.objectContaining({ method: 'DELETE' }));
  });
});
