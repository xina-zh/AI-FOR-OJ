import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { getExperiment } from '../../api/experimentApi';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { ExperimentResultPanel } from './ExperimentResultPanel';

export function ExperimentDetailPage() {
  const experimentId = Number(useParams().id);
  const { data, isLoading, error } = useQuery({
    queryKey: ['experiment', experimentId],
    queryFn: () => getExperiment(experimentId),
    enabled: Number.isFinite(experimentId) && experimentId > 0,
  });

  if (isLoading) return <LoadingBlock label="加载 experiment" />;
  if (error) return <ErrorPanel error={error} />;
  if (!data) return null;

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>{data.name}</h1>
          <p>
            {data.model} · {data.prompt_name} · {data.agent_name}
          </p>
        </div>
        <Link className="button button-secondary" to="/experiments">
          返回实验
        </Link>
      </div>
      <ExperimentResultPanel experiment={data} />
    </section>
  );
}
