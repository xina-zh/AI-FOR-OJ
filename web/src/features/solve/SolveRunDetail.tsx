import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { getAISolveRun } from '../../api/aiApi';
import { CodeBlock } from '../../components/code/CodeBlock';
import { LatencySummary } from '../../components/metrics/LatencySummary';
import { TokenSummary } from '../../components/metrics/TokenSummary';
import { VerdictBadge } from '../../components/metrics/VerdictBadge';
import { Card } from '../../components/ui/Card';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';

export function SolveRunDetail() {
  const runId = Number(useParams().id);
  const { data, isLoading, error } = useQuery({
    queryKey: ['ai-run', runId],
    queryFn: () => getAISolveRun(runId),
    enabled: Number.isFinite(runId) && runId > 0,
  });

  if (isLoading) return <LoadingBlock />;
  if (error) return <ErrorPanel error={error} />;
  if (!data) return null;

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>AI Run #{data.id}</h1>
          <p>{data.model} · {data.prompt_name} · {data.agent_name}</p>
        </div>
        <VerdictBadge verdict={data.verdict} />
      </div>
      <Card>
        <TokenSummary input={data.token_input} output={data.token_output} />
        <LatencySummary llm={data.llm_latency_ms} total={data.total_latency_ms} />
        <h2>Prompt</h2>
        <CodeBlock code={data.prompt_preview || ''} />
        <h2>Raw Response</h2>
        <CodeBlock code={data.raw_response || ''} language="markdown" />
        <h2>Extracted Code</h2>
        <CodeBlock code={data.extracted_code || ''} language="cpp" />
      </Card>
    </section>
  );
}
