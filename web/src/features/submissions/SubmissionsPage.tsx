import { useQuery } from '@tanstack/react-query';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { StatusBadge } from '../../components/StatusBadge';
import { PageHeader } from '../shared';

export function SubmissionsPage() {
  const query = useQuery({ queryKey: ['submissions', 1], queryFn: () => experimentApi.listSubmissions(1, 20) });
  return (
    <section className="route-panel">
      <PageHeader eyebrow="Judge" title="Submissions" />
      <DataTable rows={query.data?.items ?? []} getRowKey={(row) => row.id} columns={[
        { key: 'id', header: 'ID', render: (row) => row.id },
        { key: 'problem', header: 'Problem', render: (row) => row.problem_title || row.problem_id },
        { key: 'verdict', header: 'Verdict', render: (row) => <StatusBadge value={row.verdict} /> },
        { key: 'passed', header: 'Passed', render: (row) => `${row.passed_count}/${row.total_count}` },
      ]} />
    </section>
  );
}
