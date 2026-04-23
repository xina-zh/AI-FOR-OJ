import { afterEach, describe, expect, it, vi } from 'vitest';

import { request } from './http';
import { mockFetch } from '../test/server';

describe('request', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('returns parsed JSON for successful responses', async () => {
    mockFetch({ ok: true });

    await expect(request<{ ok: boolean }>('/health')).resolves.toEqual({ ok: true });
  });

  it('throws ApiError with backend error message', async () => {
    mockFetch({ error: 'unknown solve agent' }, { status: 400 });

    await expect(request('/api/v1/ai/solve')).rejects.toMatchObject({
      status: 400,
      message: 'unknown solve agent',
    });
  });
});
