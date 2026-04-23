import { Link } from 'react-router-dom';

import type { Experiment } from '../../api/types';
import { TokenSummary } from '../../components/metrics/TokenSummary';
import { VerdictDistribution } from '../../components/metrics/VerdictDistribution';
import { VerdictBadge } from '../../components/metrics/VerdictBadge';
import { Card } from '../../components/ui/Card';
import { Table } from '../../components/ui/Table';
import { formatLatency } from '../../lib/format';

export function ExperimentResultPanel({ experiment }: { experiment: Experiment }) {
  return (
    <Card>
      <div className="result-header">
        <div>
          <span className="eyebrow">Experiment #{experiment.id}</span>
          <h2>{experiment.name}</h2>
        </div>
        <span className="badge badge-info">{experiment.status}</span>
      </div>
      <div className="metric-list">
        <VerdictDistribution distribution={experiment.verdict_distribution} />
        <TokenSummary input={experiment.cost_summary.total_token_input} output={experiment.cost_summary.total_token_output} />
        <span>Total latency {formatLatency(experiment.cost_summary.total_latency_ms)}</span>
      </div>
      <ExperimentRunsTable runs={experiment.runs} />
    </Card>
  );
}

function ExperimentRunsTable({ runs }: { runs: Experiment['runs'] }) {
  if (runs.length === 0) {
    return <p className="muted">暂无 run 明细。</p>;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Run</th>
          <th>Problem</th>
          <th>Verdict</th>
          <th>Status</th>
          <th>Links</th>
        </tr>
      </thead>
      <tbody>
        {runs.map((run) => (
          <tr key={run.id}>
            <td>Run #{run.id}</td>
            <td>{run.problem_id}</td>
            <td>{run.verdict ? <VerdictBadge verdict={run.verdict} /> : '-'}</td>
            <td>{run.status}</td>
            <td>
              <div className="link-list">
                <Link to={`/trace/experiment-runs/${run.id}`}>Trace #{run.id}</Link>
                {run.ai_solve_run_id ? <Link to={`/ai-runs/${run.ai_solve_run_id}`}>AI Run #{run.ai_solve_run_id}</Link> : null}
                {run.submission_id ? <Link to={`/submissions/${run.submission_id}`}>Submission #{run.submission_id}</Link> : null}
              </div>
            </td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
