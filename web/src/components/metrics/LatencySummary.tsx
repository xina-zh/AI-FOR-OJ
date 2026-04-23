import { formatLatency } from '../../lib/format';

interface LatencySummaryProps {
  llm?: number;
  total?: number;
}

export function LatencySummary({ llm = 0, total = 0 }: LatencySummaryProps) {
  return (
    <div className="metric-row">
      <span>LLM {formatLatency(llm)}</span>
      <strong>Total {formatLatency(total)}</strong>
    </div>
  );
}
