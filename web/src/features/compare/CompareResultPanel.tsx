import { Link } from 'react-router-dom';

import type { CompareExperiment } from '../../api/types';
import { CostComparison } from '../../components/metrics/CostComparison';
import { VerdictDistribution } from '../../components/metrics/VerdictDistribution';
import { Card } from '../../components/ui/Card';
import { Table } from '../../components/ui/Table';

export function CompareResultPanel({ compare }: { compare: CompareExperiment }) {
  return (
    <Card>
      <div className="result-header">
        <div>
          <span className="eyebrow">Compare #{compare.id}</span>
          <h2>{compare.name}</h2>
        </div>
        <span className="badge badge-info">{compare.status}</span>
      </div>
      <div className="metric-list">
        <div className="comparison-grid">
          <span>Dimension {compare.compare_dimension}</span>
          <span>Baseline {compare.baseline_value}</span>
          <strong>Candidate {compare.candidate_value}</strong>
          <span>AC delta {compare.delta_ac_count}</span>
          <span>Failed delta {compare.delta_failed_count}</span>
          <strong>{compare.comparison_summary.tradeoff_type || '-'}</strong>
        </div>
        <CostComparison
          baselineTokens={compare.cost_comparison.baseline_total_tokens ?? 0}
          candidateTokens={compare.cost_comparison.candidate_total_tokens ?? 0}
          deltaTokens={compare.cost_comparison.delta_total_tokens ?? 0}
          baselineLatency={compare.cost_comparison.baseline_average_total_latency_ms ?? 0}
          candidateLatency={compare.cost_comparison.candidate_average_total_latency_ms ?? 0}
        />
        <div className="comparison-columns">
          <div>
            <h2>Baseline verdict</h2>
            <VerdictDistribution distribution={compare.baseline_verdict_distribution} />
          </div>
          <div>
            <h2>Candidate verdict</h2>
            <VerdictDistribution distribution={compare.candidate_verdict_distribution} />
          </div>
        </div>
      </div>
      <HighlightedProblemsTable compare={compare} />
    </Card>
  );
}

function HighlightedProblemsTable({ compare }: { compare: CompareExperiment }) {
  const rows = compare.highlighted_problems.length > 0 ? compare.highlighted_problems : compare.problem_summaries;
  if (rows.length === 0) {
    return <p className="muted">暂无题目差异。</p>;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Problem</th>
          <th>Baseline</th>
          <th>Candidate</th>
          <th>Submissions</th>
          <th>Change</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((row) => (
          <tr key={`${row.problem_id}-${row.change_type}`}>
            <td>{row.problem_id}</td>
            <td>{row.baseline_verdict || row.baseline_status || '-'}</td>
            <td>{row.candidate_verdict || row.candidate_status || '-'}</td>
            <td>
              <div className="link-list">
                {row.baseline_submission_id ? <Link to={`/submissions/${row.baseline_submission_id}`}>Baseline #{row.baseline_submission_id}</Link> : null}
                {row.candidate_submission_id ? <Link to={`/submissions/${row.candidate_submission_id}`}>Candidate #{row.candidate_submission_id}</Link> : null}
                {!row.baseline_submission_id && !row.candidate_submission_id ? '-' : null}
              </div>
            </td>
            <td>{row.change_type}</td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
