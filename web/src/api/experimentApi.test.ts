import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import { createApiClient } from './client';
import { createExperimentApi } from './experimentApi';

describe('experimentApi', () => {
  const fetchMock = vi.fn();

  beforeEach(() => {
    fetchMock.mockReset();
    vi.stubGlobal('fetch', fetchMock);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('loads experiment options from the default api base url', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          default_model: 'mock-model',
          models: [{ name: 'mock-model', label: 'mock-model' }],
          prompts: [{ name: 'default', label: 'default' }],
          agents: [{ name: 'direct_codegen', label: 'direct_codegen' }],
          tooling_options: [{ name: 'sample_judge', label: 'sample_judge' }],
        }),
        { status: 200, headers: { 'Content-Type': 'application/json' } },
      ),
    );

    const api = createExperimentApi(createApiClient());
    const options = await api.getExperimentOptions();

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/meta/experiment-options', {
      headers: { Accept: 'application/json' },
    });
    expect(options.default_model).toBe('mock-model');
    expect(options.tooling_options[0].name).toBe('sample_judge');
  });

  it('posts experiment runs with json payloads', async () => {
    fetchMock.mockResolvedValueOnce(
      new Response(JSON.stringify({ id: 10, name: 'smoke', runs: [] }), {
        status: 201,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    const api = createExperimentApi(createApiClient('http://localhost:8080/api/v1'));
    const result = await api.runExperiment({
      name: 'smoke',
      problem_ids: [1, 2],
      model: 'mock-model',
      prompt_name: 'default',
      agent_name: 'direct_codegen',
      tooling_config: '{}',
    });

    expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/v1/experiments/run', {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        name: 'smoke',
        problem_ids: [1, 2],
        model: 'mock-model',
        prompt_name: 'default',
        agent_name: 'direct_codegen',
        tooling_config: '{}',
      }),
    });
    expect(result.id).toBe(10);
  });
});
