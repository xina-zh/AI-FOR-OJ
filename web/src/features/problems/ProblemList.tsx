import { Link } from 'react-router-dom';

import type { Problem } from '../../api/types';
import { Badge } from '../../components/ui/Badge';
import { Table } from '../../components/ui/Table';

interface ProblemListProps {
  problems: Problem[];
  deletingProblemId?: number | null;
  onDelete?: (problem: Problem) => void;
}

export function ProblemList({ problems, deletingProblemId = null, onDelete }: ProblemListProps) {
  return (
    <Table>
      <thead>
        <tr>
          <th>ID</th>
          <th>标题</th>
          <th>难度</th>
          <th>限制</th>
          <th>标签</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        {problems.map((problem) => (
          <tr key={problem.id}>
            <td>{problem.id}</td>
            <td>
              <Link to={`/problems/${problem.id}`}>{problem.title}</Link>
            </td>
            <td>
              <Badge tone="info">{problem.difficulty}</Badge>
            </td>
            <td>
              {problem.time_limit_ms}ms / {problem.memory_limit_mb}MB
            </td>
            <td>{problem.tags || '-'}</td>
            <td>
	              <button
	                className="button button-danger"
	                type="button"
	                disabled={!onDelete || deletingProblemId === problem.id}
	                aria-label={`永久删除 ${problem.title}`}
	                onClick={() => onDelete?.(problem)}
	              >
                删除
              </button>
            </td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
