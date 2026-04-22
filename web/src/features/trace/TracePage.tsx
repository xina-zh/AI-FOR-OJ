import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { StatusBadge } from '../../components/StatusBadge';
import { ErrorBlock, LoadingBlock, PageHeader } from '../shared';

export function TracePage() {
  const id = Number(useParams().id);
  const query = useQuery({ queryKey: ['trace', id], queryFn: () => experimentApi.getExperimentRunTrace(id), enabled: Number.isFinite(id) });
  if (query.isLoading) return <LoadingBlock />;
  if (query.error) return <ErrorBlock error={query.error} />;
  const trace = query.data;
  if (!trace) return null;
  return (
    <section className="route-panel">
      <PageHeader eyebrow="Trace" title="Experiment Run Trace" />
      <div className="panel"><StatusBadge value={trace.status} /> Problem {trace.problem_id}</div>
      <DataTable rows={trace.timeline} getRowKey={(row) => row.id} columns={[
        { key: 'seq', header: '#', render: (row) => row.sequence_no },
        { key: 'type', header: 'Type', render: (row) => row.step_type },
        { key: 'content', header: 'Content', render: (row) => row.content },
        { key: 'metadata', header: 'Metadata', render: (row) => row.metadata },
      ]} />
    </section>
  );
}
