import { useQuery } from '@tanstack/react-query';

import { experimentApi } from '../../api/experimentApi';
import { MetricStrip } from '../../components/MetricStrip';
import { PageHeader } from '../shared';

export function DashboardPage() {
  const experiments = useQuery({ queryKey: ['experiments', 1], queryFn: () => experimentApi.listExperiments(1, 20) });
  const compares = useQuery({ queryKey: ['compares', 1], queryFn: () => experimentApi.listCompares(1, 20) });
  const repeats = useQuery({ queryKey: ['repeats', 1], queryFn: () => experimentApi.listRepeats(1, 20) });

  return (
    <section className="route-panel">
      <PageHeader eyebrow="Overview" title="Experiment Dashboard" />
      <MetricStrip
        items={[
          { label: 'Experiments', value: experiments.data?.total ?? 0 },
          { label: 'Compares', value: compares.data?.total ?? 0 },
          { label: 'Repeats', value: repeats.data?.total ?? 0 },
        ]}
      />
    </section>
  );
}
