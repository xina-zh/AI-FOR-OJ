import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { VerdictBadge } from './VerdictBadge';

describe('VerdictBadge', () => {
  it('renders the verdict label', () => {
    render(<VerdictBadge verdict="AC" />);

    expect(screen.getByText('AC')).toBeInTheDocument();
  });
});
