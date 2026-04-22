import '@testing-library/jest-dom/vitest';

import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it } from 'vitest';

import { Layout } from './Layout';

describe('Layout', () => {
  it('renders the console shell navigation and page content', () => {
    render(
      <MemoryRouter initialEntries={['/experiments']}>
        <Layout>
          <h1>Experiment History</h1>
        </Layout>
      </MemoryRouter>,
    );

    expect(screen.getByText('AI-For-Oj')).toBeInTheDocument();
    expect(screen.getByRole('navigation', { name: 'primary navigation' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: /Experiments Batch/i })).toHaveClass('active');
    expect(screen.getByRole('heading', { name: 'Experiment History' })).toBeInTheDocument();
  });
});
