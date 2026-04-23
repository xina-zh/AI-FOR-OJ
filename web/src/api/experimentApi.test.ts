import { afterEach, describe, expect, it, vi } from 'vitest';

import { listExperiments } from './experimentApi';
import { mockFetch } from '../test/server';

describe('experimentApi', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('uses page and page_size query parameters when listing experiments', async () => {
    const fetchMock = mockFetch({
      items: [],
      page: 2,
      page_size: 5,
      total: 0,
      total_pages: 0,
    });

    const output = await listExperiments({ page: 2, pageSize: 5 });

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/experiments?page=2&page_size=5', expect.any(Object));
    expect(output.page).toBe(2);
    expect(output.page_size).toBe(5);
  });
});
