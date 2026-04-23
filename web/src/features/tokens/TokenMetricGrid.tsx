import type { ExperimentCostSummary } from '../../api/types';
import { Card } from '../../components/ui/Card';
import { formatLatency, formatTokens } from '../../lib/format';

export function TokenMetricGrid({ summary }: { summary: ExperimentCostSummary }) {
  return (
    <div className="metric-card-grid">
      <Card>
        <span className="eyebrow">Input</span>
        <strong>{formatTokens(summary.total_token_input)}</strong>
        <span className="muted">avg {formatTokens(summary.average_token_input)}</span>
      </Card>
      <Card>
        <span className="eyebrow">Output</span>
        <strong>{formatTokens(summary.total_token_output)}</strong>
        <span className="muted">avg {formatTokens(summary.average_token_output)}</span>
      </Card>
      <Card>
        <span className="eyebrow">Total</span>
        <strong>{formatTokens(summary.total_tokens)}</strong>
        <span className="muted">avg {formatTokens(summary.average_total_tokens)}</span>
      </Card>
      <Card>
        <span className="eyebrow">Latency</span>
        <strong>{formatLatency(summary.total_latency_ms)}</strong>
        <span className="muted">avg {formatLatency(summary.average_total_latency_ms)}</span>
      </Card>
    </div>
  );
}
