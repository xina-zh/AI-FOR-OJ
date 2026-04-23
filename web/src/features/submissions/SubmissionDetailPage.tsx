import { useQuery } from '@tanstack/react-query';
import { Link, useParams } from 'react-router-dom';

import { getSubmission } from '../../api/submissionApi';
import { CodeBlock } from '../../components/code/CodeBlock';
import { VerdictBadge } from '../../components/metrics/VerdictBadge';
import { Card } from '../../components/ui/Card';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { TestCaseResultTable } from './TestCaseResultTable';

export function SubmissionDetailPage() {
  const submissionId = Number(useParams().id);
  const { data, isLoading, error } = useQuery({
    queryKey: ['submission', submissionId],
    queryFn: () => getSubmission(submissionId),
    enabled: Number.isFinite(submissionId) && submissionId > 0,
  });

  if (isLoading) return <LoadingBlock label="加载 submission" />;
  if (error) return <ErrorPanel error={error} />;
  if (!data) return null;

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>Submission #{data.id}</h1>
          <p>
            {data.problem_title} · {data.language} · {data.source_type}
          </p>
        </div>
        <Link className="button button-secondary" to="/submissions">
          返回列表
        </Link>
      </div>

      <Card>
        <div className="metric-row">
          <VerdictBadge verdict={data.verdict} />
          <span>{data.runtime_ms}ms</span>
          <span>{data.memory_kb}KB</span>
          <span>
            {data.passed_count}/{data.total_count}
          </span>
          <span>{data.timed_out ? 'timeout' : `exit ${data.exit_code}`}</span>
        </div>
      </Card>

      <Card>
        <h2>Source code</h2>
        <CodeBlock code={data.source_code} language={data.language} />
      </Card>

      <Card>
        <h2>Judge output</h2>
        {data.compile_stderr ? <CodeBlock code={data.compile_stderr} language="text" /> : null}
        {data.run_stdout ? <CodeBlock code={data.run_stdout} language="text" /> : null}
        {data.run_stderr ? <CodeBlock code={data.run_stderr} language="text" /> : null}
        {data.error_message ? <CodeBlock code={data.error_message} language="text" /> : null}
      </Card>

      <Card>
        <h2>Test cases</h2>
        <TestCaseResultTable results={data.testcase_results ?? []} />
      </Card>
    </section>
  );
}
