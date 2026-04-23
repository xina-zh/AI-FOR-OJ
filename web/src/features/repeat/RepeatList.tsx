import { Link } from 'react-router-dom';

import type { RepeatExperiment } from '../../api/types';
import { EmptyState } from '../../components/ui/EmptyState';
import { Table } from '../../components/ui/Table';
import { formatPercent, formatTokens } from '../../lib/format';

export function RepeatList({ repeats }: { repeats: RepeatExperiment[] }) {
  if (repeats.length === 0) {
    return <EmptyState title="暂无重复实验" message="运行一次 repeat 后会在这里出现历史记录。" />;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Variables</th>
          <th>Stability</th>
          <th>Tokens</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {repeats.map((repeat) => (
          <tr key={repeat.id}>
            <td>
              <Link to={`/repeat/${repeat.id}`}>{repeat.name}</Link>
            </td>
            <td>
              {repeat.model}
              <br />
              <span className="muted">
                {repeat.prompt_name} / {repeat.agent_name} / {repeat.repeat_count}x
              </span>
            </td>
            <td>
              {formatPercent(repeat.overall_ac_rate)}
              <br />
              <span className="muted">
                best {repeat.best_round_ac_count}, worst {repeat.worst_round_ac_count}
              </span>
            </td>
            <td>{formatTokens(repeat.cost_summary.total_tokens)}</td>
            <td>{repeat.status}</td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
