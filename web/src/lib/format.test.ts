import { describe, expect, it } from 'vitest';

import { formatLatency, formatTokens, signedNumber } from './format';

describe('format helpers', () => {
  it('formats token counts with locale separators', () => {
    expect(formatTokens(123456)).toBe('123,456');
  });

  it('formats latency in seconds when values are large', () => {
    expect(formatLatency(2300)).toBe('2.30s');
  });

  it('formats signed numeric deltas', () => {
    expect(signedNumber(60)).toBe('+60');
    expect(signedNumber(-3)).toBe('-3');
    expect(signedNumber(0)).toBe('0');
  });
});
