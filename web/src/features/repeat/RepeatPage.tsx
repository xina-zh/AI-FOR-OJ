import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { listRepeats, repeatExperiment } from '../../api/experimentApi';
import { Card } from '../../components/ui/Card';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { RepeatForm } from './RepeatForm';
import { RepeatList } from './RepeatList';
import { RepeatResultPanel } from './RepeatResultPanel';

export function RepeatPage() {
  const queryClient = useQueryClient();
  const repeats = useQuery({
    queryKey: ['repeats'],
    queryFn: () => listRepeats({ pageSize: 20 }),
  });
  const mutation = useMutation({
    mutationFn: repeatExperiment,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['repeats'] });
    },
  });

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>重复实验</h1>
          <p>用相同变量重复运行同一批题，观察稳定性和最不稳定题目。</p>
        </div>
      </div>
      <RepeatForm isSubmitting={mutation.isPending} onSubmit={mutation.mutate} />
      {mutation.error ? <ErrorPanel error={mutation.error} /> : null}
      {mutation.data ? <RepeatResultPanel repeat={mutation.data} /> : null}
      <Card>
        <h2>Repeat 历史</h2>
        {repeats.isLoading ? <LoadingBlock label="加载 repeat 历史" /> : null}
        {repeats.error ? <ErrorPanel error={repeats.error} /> : null}
        {repeats.data ? <RepeatList repeats={repeats.data.items} /> : null}
      </Card>
    </section>
  );
}
