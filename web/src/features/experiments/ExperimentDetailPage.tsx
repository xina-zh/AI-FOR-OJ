import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { StatusBadge } from '../../components/StatusBadge';
import { CostStrip, DetailLink, DistributionStrip, ErrorBlock, LoadingBlock, PageHeader } from '../shared';

export function ExperimentDetailPage() {
  const id = Number(useParams().id);
  const query = useQuery({ queryKey: ['experiment', id], queryFn: () => experimentApi.getExperiment(id), enabled: Number.isFinite(id) });

  if (query.isLoading) return <LoadingBlock />;
  if (query.error) return <ErrorBlock error={query.error} />;
  const experiment = query.data;
  if (!experiment) return null;

  return (
    <section className="route-panel">
      <PageHeader eyebrow="Experiment Detail" title={experiment.name} />
      <DistributionStrip distribution={experiment.verdict_distribution} />
      <CostStrip summary={experiment.cost_summary} />
      <DataTable
        rows={experiment.runs}
        getRowKey={(row) => row.id}
        columns={[
          { key: 'problem', header: 'Problem', render: (row) => row.problem_id },
          { key: 'verdict', header: 'Verdict', render: (row) => <StatusBadge value={row.verdict || row.status} /> },
          { key: 'solve', header: 'AI Run', render: (row) => row.ai_solve_run_id ?? '-' },
          { key: 'trace', header: 'Trace', render: (row) => <DetailLink to={`/trace/experiment-runs/${row.id}`}>Open</DetailLink> },
        ]}
      />
    </section>
  );
}
