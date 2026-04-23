import { Link } from 'react-router-dom';

import type { SubmissionSummary } from '../../api/types';
import { VerdictBadge } from '../../components/metrics/VerdictBadge';
import { Table } from '../../components/ui/Table';

export function SubmissionTable({ submissions }: { submissions: SubmissionSummary[] }) {
  if (submissions.length === 0) {
    return <p className="muted">暂无 submission。</p>;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Problem</th>
          <th>Verdict</th>
          <th>Runtime</th>
          <th>Passed</th>
        </tr>
      </thead>
      <tbody>
        {submissions.map((submission) => (
          <tr key={submission.id}>
            <td>
              <Link to={`/submissions/${submission.id}`}>#{submission.id}</Link>
            </td>
            <td>{submission.problem_title}</td>
            <td>
              <VerdictBadge verdict={submission.verdict} />
            </td>
            <td>{submission.runtime_ms}ms</td>
            <td>
              {submission.passed_count}/{submission.total_count}
            </td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
