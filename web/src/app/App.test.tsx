import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { App } from './App';

describe('App', () => {
  it('renders the dashboard inside the application shell', async () => {
    render(<App />);

    expect(await screen.findByRole('heading', { name: '实验控制台' })).toBeInTheDocument();
    expect(screen.getAllByRole('link', { name: '单题 Solve' }).length).toBeGreaterThan(0);
    expect(screen.getAllByRole('link', { name: 'Compare' }).length).toBeGreaterThan(0);
  });
});
