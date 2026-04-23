import { Link } from 'react-router-dom';

import type { CompareExperiment } from '../../api/types';
import { EmptyState } from '../../components/ui/EmptyState';
import { Table } from '../../components/ui/Table';

export function CompareList({ compares }: { compares: CompareExperiment[] }) {
  if (compares.length === 0) {
    return <EmptyState title="暂无对比实验" message="运行一次 compare 后会在这里出现历史记录。" />;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Name</th>
          <th>Dimension</th>
          <th>Delta</th>
          <th>Summary</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {compares.map((compare) => (
          <tr key={compare.id}>
            <td>
              <Link to={`/compare/${compare.id}`}>{compare.name}</Link>
            </td>
            <td>
              {compare.compare_dimension}
              <br />
              <span className="muted">
                {compare.baseline_value}
                {' -> '}
                {compare.candidate_value}
              </span>
            </td>
            <td>
              AC {compare.delta_ac_count}
              <br />
              Failed {compare.delta_failed_count}
            </td>
            <td>{compare.comparison_summary.tradeoff_type || '-'}</td>
            <td>{compare.status}</td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
