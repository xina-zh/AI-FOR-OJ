import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { getRepeat } from '../../api/experimentApi';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { RepeatResultPanel } from './RepeatResultPanel';

export function RepeatDetailPage() {
  const repeatId = Number(useParams().id);
  const { data, isLoading, error } = useQuery({
    queryKey: ['repeat', repeatId],
    queryFn: () => getRepeat(repeatId),
    enabled: Number.isFinite(repeatId) && repeatId > 0,
  });

  if (isLoading) return <LoadingBlock label="加载 repeat" />;
  if (error) return <ErrorPanel error={error} />;
  if (!data) return null;

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>{data.name}</h1>
          <p>
            {data.model} · {data.prompt_name} · {data.agent_name} · {data.repeat_count}x
          </p>
        </div>
        <Link className="button button-secondary" to="/repeat">
          返回 Repeat
        </Link>
      </div>
      <RepeatResultPanel repeat={data} />
    </section>
  );
}
