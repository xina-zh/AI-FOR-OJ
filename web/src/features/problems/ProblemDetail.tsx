import { useQuery } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';

import { getProblem } from '../../api/problemApi';
import { Card } from '../../components/ui/Card';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { TestCasePanel } from './TestCasePanel';

interface SampleCase {
  input: string;
  output: string;
}

export function ProblemDetail() {
  const problemId = Number(useParams().id);
  const { data, isLoading, error } = useQuery({
    queryKey: ['problem', problemId],
    queryFn: () => getProblem(problemId),
    enabled: Number.isFinite(problemId) && problemId > 0,
  });

  if (isLoading) return <LoadingBlock />;
  if (error) return <ErrorPanel error={error} />;
  if (!data) return null;

  const samples = parseSamples(data.samples);

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>{data.title}</h1>
          <p>{data.difficulty} · {data.time_limit_ms}ms · {data.memory_limit_mb}MB</p>
        </div>
      </div>
      <Card>
        <h2>题面</h2>
        <p>{data.description}</p>
        <h2>输入</h2>
        <p>{data.input_spec}</p>
        <h2>输出</h2>
        <p>{data.output_spec}</p>
        {samples.length > 0 ? (
          <>
            <h2>样例</h2>
            {samples.map((sample, index) => (
              <div className="stack" key={`${sample.input}-${index}`}>
                <h3>样例输入 {index + 1}</h3>
                <pre className="plain-pre">{sample.input}</pre>
                <h3>样例输出 {index + 1}</h3>
                <pre className="plain-pre">{sample.output}</pre>
              </div>
            ))}
          </>
        ) : data.samples ? (
          <>
            <h2>样例</h2>
            <pre className="plain-pre">{data.samples}</pre>
          </>
        ) : null}
      </Card>
      <TestCasePanel problemId={data.id} />
    </section>
  );
}

function parseSamples(raw: string): SampleCase[] {
  if (!raw.trim()) {
    return [];
  }

  try {
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.flatMap((item) => {
      if (!isSampleCase(item)) {
        return [];
      }
      return [{ input: item.input, output: item.output }];
    });
  } catch {
    return [];
  }
}

function isSampleCase(value: unknown): value is SampleCase {
  if (!value || typeof value !== 'object') {
    return false;
  }
  const item = value as { input?: unknown; output?: unknown };
  return typeof item.input === 'string' && typeof item.output === 'string';
}
