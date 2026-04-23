import { Link } from 'react-router-dom';

import type { Experiment } from '../../api/types';
import { VerdictDistribution } from '../../components/metrics/VerdictDistribution';
import { EmptyState } from '../../components/ui/EmptyState';
import { Table } from '../../components/ui/Table';
import { formatTokens } from '../../lib/format';

export function ExperimentList({ experiments }: { experiments: Experiment[] }) {
  if (experiments.length === 0) {
    return <EmptyState title="暂无批量实验" message="运行一次 experiment 后会在这里出现历史记录。" />;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Variables</th>
          <th>Result</th>
          <th>Tokens</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {experiments.map((experiment) => (
          <tr key={experiment.id}>
            <td>
              <Link to={`/experiments/${experiment.id}`}>{experiment.name}</Link>
            </td>
            <td>
              {experiment.model}
              <br />
              <span className="muted">
                {experiment.prompt_name} / {experiment.agent_name}
              </span>
            </td>
            <td>
              {experiment.ac_count}/{experiment.total_count} AC
              <VerdictDistribution distribution={experiment.verdict_distribution} />
            </td>
            <td>{formatTokens(experiment.cost_summary.total_tokens)}</td>
            <td>{experiment.status}</td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
