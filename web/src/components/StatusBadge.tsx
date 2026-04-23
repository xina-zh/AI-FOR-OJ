const positive = new Set(['AC', 'success', 'completed']);
const negative = new Set(['WA', 'RE', 'CE', 'TLE', 'MLE', 'failed', 'UNJUDGEABLE']);
const pending = new Set(['running', 'queued']);

export function StatusBadge({ value }: { value?: string }) {
  const label = value || 'unknown';
  const normalized = label.toLowerCase();
  const tone = positive.has(label) || positive.has(normalized)
    ? 'positive'
    : negative.has(label) || negative.has(normalized)
      ? 'negative'
      : pending.has(normalized)
        ? 'pending'
        : 'neutral';

  return <span className={`status-badge ${tone}`}>{label}</span>;
}
