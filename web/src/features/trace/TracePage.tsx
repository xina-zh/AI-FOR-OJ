import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { getExperimentRunTrace } from '../../api/traceApi';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { TraceTimeline } from './TraceTimeline';

export function TracePage() {
  const runId = Number(useParams().id);
  const { data, isLoading, error } = useQuery({
    queryKey: ['experiment-run-trace', runId],
    queryFn: () => getExperimentRunTrace(runId),
    enabled: Number.isFinite(runId) && runId > 0,
  });

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>Trace #{runId}</h1>
          <p>回放 prompt、模型输出、代码提取、判题摘要和测试点结果。</p>
        </div>
        <Link className="button button-secondary" to="/experiments">
          返回实验
        </Link>
      </div>
      {isLoading ? <LoadingBlock label="加载 trace" /> : null}
      {error ? <ErrorPanel error={error} /> : null}
      {data ? <TraceTimeline events={data.events} /> : null}
    </section>
  );
}
