export function formatTokens(value: number | undefined | null) {
  return Number(value ?? 0).toLocaleString('en-US');
}

export function formatLatency(ms: number | undefined | null) {
  const value = Number(ms ?? 0);
  if (value >= 1000) {
    return `${(value / 1000).toFixed(2)}s`;
  }
  return `${value}ms`;
}

export function formatPercent(value: number | undefined | null) {
  return `${((value ?? 0) * 100).toFixed(1)}%`;
}

export function signedNumber(value: number | undefined | null) {
  const normalized = Number(value ?? 0);
  if (normalized > 0) {
    return `+${normalized}`;
  }
  return String(normalized);
}
