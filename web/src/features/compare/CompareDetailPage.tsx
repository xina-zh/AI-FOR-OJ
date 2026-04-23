import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { getCompare } from '../../api/experimentApi';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { CompareResultPanel } from './CompareResultPanel';

export function CompareDetailPage() {
  const compareId = Number(useParams().id);
  const { data, isLoading, error } = useQuery({
    queryKey: ['compare', compareId],
    queryFn: () => getCompare(compareId),
    enabled: Number.isFinite(compareId) && compareId > 0,
  });

  if (isLoading) return <LoadingBlock label="加载 compare" />;
  if (error) return <ErrorPanel error={error} />;
  if (!data) return null;

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>{data.name}</h1>
          <p>
            {data.baseline_value} {'->'} {data.candidate_value}
          </p>
        </div>
        <Link className="button button-secondary" to="/compare">
          返回对比
        </Link>
      </div>
      <CompareResultPanel compare={data} />
    </section>
  );
}
