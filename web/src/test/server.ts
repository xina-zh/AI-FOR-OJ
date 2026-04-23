import { vi } from 'vitest';

export function mockFetch(response: unknown, init?: ResponseInit) {
  const fetchMock = vi.fn().mockResolvedValue(
    new Response(JSON.stringify(response), {
      status: init?.status ?? 200,
      headers: { 'Content-Type': 'application/json', ...(init?.headers ?? {}) },
    }),
  );
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
}
