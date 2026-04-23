import { Link } from 'react-router-dom';

import type { AISolveResponse } from '../../api/types';
import { CodeBlock } from '../../components/code/CodeBlock';
import { LatencySummary } from '../../components/metrics/LatencySummary';
import { TokenSummary } from '../../components/metrics/TokenSummary';
import { VerdictBadge } from '../../components/metrics/VerdictBadge';
import { Card } from '../../components/ui/Card';

export function SolveResultPanel({ result }: { result: AISolveResponse }) {
  return (
    <Card>
      <div className="result-header">
        <h2>运行结果</h2>
        <VerdictBadge verdict={result.verdict} />
      </div>
      <TokenSummary input={result.token_input} output={result.token_output} />
      <LatencySummary llm={result.llm_latency_ms} total={result.total_latency_ms} />
      <div className="link-list">
        <Link to={`/ai-runs/${result.ai_solve_run_id}`}>AI Run #{result.ai_solve_run_id}</Link>
        <Link to={`/submissions/${result.submission_id}`}>Submission #{result.submission_id}</Link>
      </div>
      <h3>Prompt Preview</h3>
      <CodeBlock code={result.prompt_preview} />
      {result.raw_response ? (
        <>
          <h3>Raw Response</h3>
          <CodeBlock code={result.raw_response} language="markdown" />
        </>
      ) : null}
      {result.extracted_code ? (
        <>
          <h3>Extracted Code</h3>
          <CodeBlock code={result.extracted_code} language="cpp" />
        </>
      ) : null}
    </Card>
  );
}
