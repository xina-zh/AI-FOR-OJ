import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { listExperiments, runExperiment } from '../../api/experimentApi';
import { Card } from '../../components/ui/Card';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { ExperimentList } from './ExperimentList';
import { ExperimentResultPanel } from './ExperimentResultPanel';
import { ExperimentRunForm } from './ExperimentRunForm';

export function ExperimentRunPage() {
  const queryClient = useQueryClient();
  const experiments = useQuery({
    queryKey: ['experiments'],
    queryFn: () => listExperiments({ pageSize: 20 }),
  });
  const mutation = useMutation({
    mutationFn: runExperiment,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['experiments'] });
    },
  });

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>批量实验</h1>
          <p>选择多道题和一组变量，运行一次完整 experiment。</p>
        </div>
      </div>
      <ExperimentRunForm isSubmitting={mutation.isPending} onSubmit={mutation.mutate} />
      {mutation.error ? <ErrorPanel error={mutation.error} /> : null}
      {mutation.data ? <ExperimentResultPanel experiment={mutation.data} /> : null}
      <Card>
        <h2>历史实验</h2>
        {experiments.isLoading ? <LoadingBlock label="加载 experiment 历史" /> : null}
        {experiments.error ? <ErrorPanel error={experiments.error} /> : null}
        {experiments.data ? <ExperimentList experiments={experiments.data.items} /> : null}
      </Card>
    </section>
  );
}
