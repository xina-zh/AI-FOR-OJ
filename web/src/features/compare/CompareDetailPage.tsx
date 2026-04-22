import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { StatusBadge } from '../../components/StatusBadge';
import { CostStrip, DistributionStrip, ErrorBlock, LoadingBlock, PageHeader } from '../shared';

export function CompareDetailPage() {
  const id = Number(useParams().id);
  const query = useQuery({ queryKey: ['compare', id], queryFn: () => experimentApi.getCompare(id), enabled: Number.isFinite(id) });
  if (query.isLoading) return <LoadingBlock />;
  if (query.error) return <ErrorBlock error={query.error} />;
  const compare = query.data;
  if (!compare) return null;

  return (
    <section className="route-panel">
      <PageHeader eyebrow="Compare Detail" title={compare.name} />
      <div className="split-grid">
        <div className="panel stack">
          <h2>Baseline</h2>
          <DistributionStrip distribution={compare.baseline_verdict_distribution} />
          <CostStrip summary={compare.baseline_summary?.cost_summary} />
        </div>
        <div className="panel stack">
          <h2>Candidate</h2>
          <DistributionStrip distribution={compare.candidate_verdict_distribution} />
          <CostStrip summary={compare.candidate_summary?.cost_summary} />
        </div>
      </div>
      <DataTable
        rows={compare.highlighted_problems ?? compare.problem_summaries ?? []}
        getRowKey={(row) => row.problem_id}
        emptyLabel="No changed problems"
        columns={[
          { key: 'problem', header: 'Problem', render: (row) => row.problem_id },
          { key: 'baseline', header: 'Baseline', render: (row) => <StatusBadge value={row.baseline_verdict} /> },
          { key: 'candidate', header: 'Candidate', render: (row) => <StatusBadge value={row.candidate_verdict} /> },
          { key: 'change', header: 'Change', render: (row) => row.change_type },
        ]}
      />
    </section>
  );
}
