import { afterEach, describe, expect, it, vi } from 'vitest';

import { getExperimentOptions } from './metaApi';
import { experimentOptionsFixture } from '../test/fixtures';
import { mockFetch } from '../test/server';

describe('metaApi', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('fetches experiment options', async () => {
    const fetchMock = mockFetch(experimentOptionsFixture);

    const options = await getExperimentOptions();

    expect(fetchMock).toHaveBeenCalledWith('/api/v1/meta/experiment-options', expect.any(Object));
    expect(options.default_model).toBe('mock-cpp17');
    expect(options.prompts).toHaveLength(3);
    expect(options.agents[1].name).toBe('direct_codegen_repair');
  });
});
