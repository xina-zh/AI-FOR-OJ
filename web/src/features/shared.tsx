import { useQuery } from '@tanstack/react-query';
import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';

import { experimentApi } from '../api/experimentApi';
import type { CostSummary, ExperimentOptions, VerdictDistribution } from '../api/types';
import { MetricStrip } from '../components/MetricStrip';

export function PageHeader({ eyebrow, title, actions }: { eyebrow: string; title: string; actions?: ReactNode }) {
  return (
    <div className="page-header">
      <div>
        <p className="section-label">{eyebrow}</p>
        <h1>{title}</h1>
      </div>
      {actions ? <div className="page-actions">{actions}</div> : null}
    </div>
  );
}

export function LoadingBlock() {
  return <div className="panel muted-panel">Loading</div>;
}

export function ErrorBlock({ error }: { error: unknown }) {
  return <div className="panel error-panel">{error instanceof Error ? error.message : 'Request failed'}</div>;
}

export function useExperimentOptions() {
  return useQuery({
    queryKey: ['experiment-options'],
    queryFn: () => experimentApi.getExperimentOptions(),
  });
}

export function defaultVariables(options?: ExperimentOptions) {
  return {
    model: options?.default_model || options?.models[0]?.name || 'mock-model',
    prompt_name: options?.prompts[0]?.name || 'default',
    agent_name: options?.agents[0]?.name || 'direct_codegen',
    tooling_config: '{}',
  };
}

export function idList(raw: string) {
  return raw
    .split(',')
    .map((item) => Number(item.trim()))
    .filter((item) => Number.isFinite(item) && item > 0);
}

export function DistributionStrip({ distribution }: { distribution?: VerdictDistribution }) {
  return (
    <MetricStrip
      items={[
        { label: 'AC', value: distribution?.ac_count ?? 0 },
        { label: 'WA', value: distribution?.wa_count ?? 0 },
        { label: 'RE/TLE', value: (distribution?.re_count ?? 0) + (distribution?.tle_count ?? 0) },
      ]}
    />
  );
}

export function CostStrip({ summary }: { summary?: CostSummary }) {
  return (
    <MetricStrip
      items={[
        { label: 'Tokens', value: summary?.total_tokens ?? 0 },
        { label: 'LLM Latency', value: summary?.total_llm_latency_ms ?? 0, unit: 'ms' },
        { label: 'Total Latency', value: summary?.total_latency_ms ?? 0, unit: 'ms' },
      ]}
    />
  );
}

export function DetailLink({ to, children }: { to: string; children: ReactNode }) {
  return (
    <Link className="inline-link" to={to}>
      {children}
    </Link>
  );
}
