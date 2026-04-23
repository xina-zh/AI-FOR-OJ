import { useQuery } from '@tanstack/react-query';

import { listCompares, listExperiments, listRepeats } from '../../api/experimentApi';
import type { ExperimentCostSummary } from '../../api/types';
import { Card } from '../../components/ui/Card';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { Table } from '../../components/ui/Table';
import { formatLatency, formatTokens, signedNumber } from '../../lib/format';
import { TokenMetricGrid } from './TokenMetricGrid';

const emptySummary: ExperimentCostSummary = {
  total_token_input: 0,
  total_token_output: 0,
  total_tokens: 0,
  average_token_input: 0,
  average_token_output: 0,
  average_total_tokens: 0,
  total_llm_latency_ms: 0,
  total_latency_ms: 0,
  average_llm_latency_ms: 0,
  average_total_latency_ms: 0,
  run_count: 0,
};

export function TokenAnalyticsPage() {
  const experiments = useQuery({
    queryKey: ['experiments', 'token-analytics'],
    queryFn: () => listExperiments({ pageSize: 20 }),
  });
  const compares = useQuery({
    queryKey: ['compares', 'token-analytics'],
    queryFn: () => listCompares({ pageSize: 20 }),
  });
  const repeats = useQuery({
    queryKey: ['repeats', 'token-analytics'],
    queryFn: () => listRepeats({ pageSize: 20 }),
  });

  const isLoading = experiments.isLoading || compares.isLoading || repeats.isLoading;
  const error = experiments.error || compares.error || repeats.error;
  const latestSummary = experiments.data?.items[0]?.cost_summary ?? repeats.data?.items[0]?.cost_summary ?? emptySummary;

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>Token 分析</h1>
          <p>汇总 single solve、experiment、compare 和 repeat 中的 token 与 latency 指标。</p>
        </div>
      </div>

      {isLoading ? <LoadingBlock label="加载 token 指标" /> : null}
      {error ? <ErrorPanel error={error} /> : null}
      <TokenMetricGrid summary={latestSummary} />

      <Card>
        <h2>Experiment token</h2>
        <Table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Total</th>
              <th>Average</th>
              <th>Latency</th>
            </tr>
          </thead>
          <tbody>
            {(experiments.data?.items ?? []).map((experiment) => (
              <tr key={experiment.id}>
                <td>{experiment.name}</td>
                <td>{formatTokens(experiment.cost_summary.total_tokens)}</td>
                <td>{formatTokens(experiment.cost_summary.average_total_tokens)}</td>
                <td>{formatLatency(experiment.cost_summary.average_total_latency_ms)}</td>
              </tr>
            ))}
          </tbody>
        </Table>
      </Card>

      <Card>
        <h2>Compare delta</h2>
        <Table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Baseline</th>
              <th>Candidate</th>
              <th>Delta</th>
              <th>Latency delta</th>
            </tr>
          </thead>
          <tbody>
            {(compares.data?.items ?? []).map((compare) => (
              <tr key={compare.id}>
                <td>{compare.name}</td>
                <td>{formatTokens(compare.cost_comparison.baseline_total_tokens ?? 0)}</td>
                <td>{formatTokens(compare.cost_comparison.candidate_total_tokens ?? 0)}</td>
                <td>{signedNumber(compare.cost_comparison.delta_total_tokens ?? 0)} tokens</td>
                <td>{signedNumber(compare.cost_comparison.delta_average_total_latency_ms ?? 0)}ms</td>
              </tr>
            ))}
          </tbody>
        </Table>
      </Card>

      <Card>
        <h2>Repeat token</h2>
        <Table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Total</th>
              <th>Runs</th>
              <th>Average</th>
            </tr>
          </thead>
          <tbody>
            {(repeats.data?.items ?? []).map((repeat) => (
              <tr key={repeat.id}>
                <td>{repeat.name}</td>
                <td>{formatTokens(repeat.cost_summary.total_tokens)}</td>
                <td>{repeat.total_run_count}</td>
                <td>{formatTokens(repeat.cost_summary.average_total_tokens)}</td>
              </tr>
            ))}
          </tbody>
        </Table>
      </Card>
    </section>
  );
}
