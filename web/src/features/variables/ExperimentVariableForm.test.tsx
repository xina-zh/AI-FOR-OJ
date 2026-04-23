import { screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { ExperimentVariableForm } from './ExperimentVariableForm';
import { experimentOptionsFixture } from '../../test/fixtures';
import { mockFetchRoutes, renderWithProviders } from '../../test/render';

describe('ExperimentVariableForm', () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it('loads prompt and agent options from metadata', async () => {
    mockFetchRoutes({
      '/api/v1/meta/experiment-options': experimentOptionsFixture,
    });

    renderWithProviders(
      <ExperimentVariableForm
        value={{ model: '', prompt_name: '', agent_name: '', tooling_config: '' }}
        onChange={() => undefined}
      />,
    );

    expect(await screen.findByDisplayValue('mock-cpp17')).toBeInTheDocument();
    expect(await screen.findByRole('option', { name: 'strict_cpp17' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'direct_codegen_repair' })).toBeInTheDocument();
    expect(screen.getByRole('option', { name: 'sample_judge' })).toBeInTheDocument();
  });
});
