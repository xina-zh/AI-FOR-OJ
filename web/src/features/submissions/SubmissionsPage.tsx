import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';

import { listProblems } from '../../api/problemApi';
import { listSubmissions } from '../../api/submissionApi';
import { Card } from '../../components/ui/Card';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { Field } from '../../components/ui/Field';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { Select } from '../../components/ui/Select';
import { SubmissionTable } from './SubmissionTable';

export function SubmissionsPage() {
  const [problemId, setProblemId] = useState<number | undefined>();
  const problems = useQuery({
    queryKey: ['problems', 'submissions-filter'],
    queryFn: listProblems,
  });
  const { data, isLoading, error } = useQuery({
    queryKey: ['submissions', problemId],
    queryFn: () => listSubmissions({ pageSize: 20, problemId }),
  });

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>提交记录</h1>
          <p>查看 judge 结果、运行输出和 testcase 明细。</p>
        </div>
      </div>
      {isLoading ? <LoadingBlock label="加载 submissions" /> : null}
      {error ? <ErrorPanel error={error} /> : null}
      <Card>
        <div className="form-grid">
          <Field label="题目筛选">
            <Select
              value={String(problemId ?? '')}
              onChange={(event) => {
                setProblemId(event.target.value ? Number(event.target.value) : undefined);
              }}
            >
              <option value="">全部题目</option>
              {(problems.data ?? []).map((problem) => (
                <option key={problem.id} value={problem.id}>
                  {problem.id} · {problem.title}
                </option>
              ))}
            </Select>
          </Field>
        </div>
        <SubmissionTable submissions={data?.items ?? []} />
      </Card>
    </section>
  );
}
