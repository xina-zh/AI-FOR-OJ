import type { RepeatExperiment } from '../../api/types';
import { TokenSummary } from '../../components/metrics/TokenSummary';
import { VerdictDistribution } from '../../components/metrics/VerdictDistribution';
import { Card } from '../../components/ui/Card';
import { Table } from '../../components/ui/Table';
import { formatPercent } from '../../lib/format';

export function RepeatResultPanel({ repeat }: { repeat: RepeatExperiment }) {
  return (
    <Card>
      <div className="result-header">
        <div>
          <span className="eyebrow">Repeat #{repeat.id}</span>
          <h2>{repeat.name}</h2>
        </div>
        <span className="badge badge-info">{repeat.status}</span>
      </div>
      <div className="metric-list">
        <div className="comparison-grid">
          <strong>整体 AC 率 {formatPercent(repeat.overall_ac_rate)}</strong>
          <span>Total runs {repeat.total_run_count}</span>
          <span>AC {repeat.overall_ac_count}</span>
          <span>Failed {repeat.overall_failed_count}</span>
          <span>Best round {repeat.best_round_ac_count}</span>
          <span>Worst round {repeat.worst_round_ac_count}</span>
        </div>
        <TokenSummary input={repeat.cost_summary.total_token_input} output={repeat.cost_summary.total_token_output} />
      </div>
      <RoundSummaryTable repeat={repeat} />
      <ProblemStabilityTable repeat={repeat} />
      <UnstableProblemsTable repeat={repeat} />
    </Card>
  );
}

function RoundSummaryTable({ repeat }: { repeat: RepeatExperiment }) {
  if (repeat.round_summaries.length === 0) {
    return <p className="muted">暂无 round 明细。</p>;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Round</th>
          <th>Experiment</th>
          <th>AC</th>
          <th>Failed</th>
          <th>Verdict</th>
        </tr>
      </thead>
      <tbody>
        {repeat.round_summaries.map((round) => (
          <tr key={round.round_no}>
            <td>Round {round.round_no}</td>
            <td>#{round.experiment_id}</td>
            <td>{round.ac_count}</td>
            <td>{round.failed_count}</td>
            <td>
              <VerdictDistribution distribution={round.verdict_distribution} />
            </td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}

function ProblemStabilityTable({ repeat }: { repeat: RepeatExperiment }) {
  if (repeat.problem_summaries.length === 0) {
    return <p className="muted">暂无 problem stability 明细。</p>;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Problem</th>
          <th>Rounds</th>
          <th>AC Rate</th>
          <th>AC / Failed</th>
          <th>Latest</th>
        </tr>
      </thead>
      <tbody>
        {repeat.problem_summaries.map((problem) => (
          <tr key={problem.problem_id}>
            <td>{problem.problem_id}</td>
            <td>{problem.total_rounds}</td>
            <td>{formatPercent(problem.ac_rate)}</td>
            <td>
              {problem.ac_count}/{problem.failed_count}
            </td>
            <td>{problem.latest_verdict || '-'}</td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}

function UnstableProblemsTable({ repeat }: { repeat: RepeatExperiment }) {
  if (repeat.most_unstable_problems.length === 0) {
    return null;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Unstable Problem</th>
          <th>AC Rate</th>
          <th>Latest</th>
          <th>Score</th>
        </tr>
      </thead>
      <tbody>
        {repeat.most_unstable_problems.map((problem) => (
          <tr key={problem.problem_id}>
            <td>{problem.problem_id}</td>
            <td>{formatPercent(problem.ac_rate)}</td>
            <td>{problem.latest_verdict || '-'}</td>
            <td>{problem.instability_score}</td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
