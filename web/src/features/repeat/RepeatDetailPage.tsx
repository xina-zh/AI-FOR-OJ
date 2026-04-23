import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { StatusBadge } from '../../components/StatusBadge';
import { CostStrip, ErrorBlock, LoadingBlock, PageHeader } from '../shared';

export function RepeatDetailPage() {
  const id = Number(useParams().id);
  const query = useQuery({ queryKey: ['repeat', id], queryFn: () => experimentApi.getRepeat(id), enabled: Number.isFinite(id) });
  if (query.isLoading) return <LoadingBlock />;
  if (query.error) return <ErrorBlock error={query.error} />;
  const repeat = query.data;
  if (!repeat) return null;
  return (
    <section className="route-panel">
      <PageHeader eyebrow="Repeat Detail" title={repeat.name} />
      <CostStrip summary={repeat.cost_summary} />
      <DataTable rows={repeat.round_summaries ?? []} getRowKey={(row) => row.round_no} columns={[
        { key: 'round', header: 'Round', render: (row) => row.round_no },
        { key: 'experiment', header: 'Experiment', render: (row) => row.experiment_id },
        { key: 'ac', header: 'AC', render: (row) => row.ac_count },
        { key: 'status', header: 'Status', render: (row) => <StatusBadge value={row.status} /> },
      ]} />
    </section>
  );
}
