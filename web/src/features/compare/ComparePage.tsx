import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { compareExperiments, listCompares } from '../../api/experimentApi';
import { Card } from '../../components/ui/Card';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { CompareForm } from './CompareForm';
import { CompareList } from './CompareList';
import { CompareResultPanel } from './CompareResultPanel';

export function ComparePage() {
  const queryClient = useQueryClient();
  const compares = useQuery({
    queryKey: ['compares'],
    queryFn: () => listCompares({ pageSize: 20 }),
  });
  const mutation = useMutation({
    mutationFn: compareExperiments,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['compares'] });
    },
  });

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>对比实验</h1>
          <p>对同一批题运行 baseline 和 candidate，比较 verdict、token 和延迟。</p>
        </div>
      </div>
      <CompareForm isSubmitting={mutation.isPending} onSubmit={mutation.mutate} />
      {mutation.error ? <ErrorPanel error={mutation.error} /> : null}
      {mutation.data ? <CompareResultPanel compare={mutation.data} /> : null}
      <Card>
        <h2>Compare 历史</h2>
        {compares.isLoading ? <LoadingBlock label="加载 compare 历史" /> : null}
        {compares.error ? <ErrorPanel error={compares.error} /> : null}
        {compares.data ? <CompareList compares={compares.data.items} /> : null}
      </Card>
    </section>
  );
}
