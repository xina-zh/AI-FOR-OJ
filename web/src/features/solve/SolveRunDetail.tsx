import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { getAISolveRun } from '../../api/aiApi';
import type { AISolveAttempt } from '../../api/types';
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
        <div className="metric-row">
          <span>Attempts {data.attempt_count}</span>
          {data.failure_type ? <span>Failure {data.failure_type}</span> : null}
          {data.strategy_path ? <span>Strategy {data.strategy_path}</span> : null}
          {data.tooling_config ? <span>{data.tooling_config}</span> : null}
          {data.tool_call_count ? <span>Tool calls {data.tool_call_count}</span> : null}
        </div>
        <TokenSummary input={data.token_input} output={data.token_output} />
        <LatencySummary llm={data.llm_latency_ms} total={data.total_latency_ms} />
        <h2>Prompt</h2>
        <CodeBlock code={data.prompt_preview || ''} />
        <h2>Raw Response</h2>
        <CodeBlock code={data.raw_response || ''} language="markdown" />
        <h2>Extracted Code</h2>
        <CodeBlock code={data.extracted_code || ''} language="cpp" />
      </Card>
      <AISolveAttempts attempts={data.attempts || []} />
    </section>
  );
}

function AISolveAttempts({ attempts }: { attempts: AISolveAttempt[] }) {
  if (attempts.length === 0) {
    return null;
  }

  return (
    <Card>
      <h2>Attempts</h2>
      <div className="stack">
        {attempts.map((attempt) => (
          <div className="stack" key={attempt.id}>
            <div className="result-header">
              <div>
                <span className="eyebrow">{attempt.stage}</span>
                <h3>Attempt #{attempt.attempt_no}</h3>
              </div>
              <VerdictBadge verdict={attempt.verdict} />
            </div>
            <div className="metric-row">
              <span>{attempt.model}</span>
              {attempt.failure_type ? <span>Failure {attempt.failure_type}</span> : null}
              {attempt.strategy_path ? <span>Strategy {attempt.strategy_path}</span> : null}
              <span>
                Judge {attempt.judge_passed_count}/{attempt.judge_total_count}
              </span>
              {attempt.timed_out ? <span>Timed out</span> : null}
            </div>
            {attempt.repair_reason ? <p>{attempt.repair_reason}</p> : null}
            {attempt.error_message ? <p className="error-panel">{attempt.error_message}</p> : null}
            <TokenSummary input={attempt.token_input} output={attempt.token_output} />
            <LatencySummary llm={attempt.llm_latency_ms} total={attempt.total_latency_ms} />
            {attempt.prompt_preview ? (
              <>
                <h4>Prompt</h4>
                <CodeBlock code={attempt.prompt_preview} />
              </>
            ) : null}
            {attempt.extracted_code ? (
              <>
                <h4>Extracted Code</h4>
                <CodeBlock code={attempt.extracted_code} language="cpp" />
              </>
            ) : null}
          </div>
        ))}
      </div>
    </Card>
  );
}
