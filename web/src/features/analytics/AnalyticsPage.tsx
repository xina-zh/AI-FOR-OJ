import { useQuery } from '@tanstack/react-query';

import { experimentApi } from '../../api/experimentApi';
import { MetricStrip } from '../../components/MetricStrip';
import { PageHeader } from '../shared';

export function AnalyticsPage() {
  const experiments = useQuery({ queryKey: ['experiments', 'analytics'], queryFn: () => experimentApi.listExperiments(1, 20) });
  const repeats = useQuery({ queryKey: ['repeats', 'analytics'], queryFn: () => experimentApi.listRepeats(1, 20) });
  const compares = useQuery({ queryKey: ['compares', 'analytics'], queryFn: () => experimentApi.listCompares(1, 20) });
  const totals = (experiments.data?.items ?? []).reduce(
    (acc, item) => ({
      tokens: acc.tokens + (item.cost_summary?.total_tokens ?? 0),
      latency: acc.latency + (item.cost_summary?.total_latency_ms ?? 0),
      ac: acc.ac + item.ac_count,
    }),
    { tokens: 0, latency: 0, ac: 0 },
  );
  for (const item of repeats.data?.items ?? []) {
    totals.tokens += item.cost_summary?.total_tokens ?? 0;
    totals.latency += item.cost_summary?.total_latency_ms ?? 0;
    totals.ac += item.overall_ac_count;
  }
  for (const item of compares.data?.items ?? []) {
    totals.tokens += item.cost_comparison?.baseline_total_tokens ?? 0;
    totals.tokens += item.cost_comparison?.candidate_total_tokens ?? 0;
    totals.latency += item.cost_comparison?.baseline_total_latency_ms ?? 0;
    totals.latency += item.cost_comparison?.candidate_total_latency_ms ?? 0;
    totals.ac += (item.baseline_summary?.ac_count ?? 0) + (item.candidate_summary?.ac_count ?? 0);
  }

  return (
    <section className="route-panel">
      <PageHeader eyebrow="Analytics" title="Token and Latency Analytics" />
      <MetricStrip items={[
        { label: 'Tokens', value: totals.tokens },
        { label: 'Latency', value: totals.latency, unit: 'ms' },
        { label: 'Accepted', value: totals.ac },
      ]} />
    </section>
  );
}
