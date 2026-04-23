import { formatLatency, formatTokens } from '../../lib/format';

interface CostComparisonProps {
  baselineTokens: number;
  candidateTokens: number;
  deltaTokens: number;
  baselineLatency: number;
  candidateLatency: number;
}

export function CostComparison({
  baselineTokens,
  candidateTokens,
  deltaTokens,
  baselineLatency,
  candidateLatency,
}: CostComparisonProps) {
  return (
    <div className="comparison-grid">
      <span>Baseline tokens {formatTokens(baselineTokens)}</span>
      <span>Candidate tokens {formatTokens(candidateTokens)}</span>
      <strong>Delta {formatTokens(deltaTokens)}</strong>
      <span>Baseline latency {formatLatency(baselineLatency)}</span>
      <span>Candidate latency {formatLatency(candidateLatency)}</span>
    </div>
  );
}
