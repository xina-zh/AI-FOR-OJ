import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';

import { createTestCase, listTestCases } from '../../api/problemApi';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { EmptyState } from '../../components/ui/EmptyState';
import { Field } from '../../components/ui/Field';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { TextArea } from '../../components/ui/TextArea';

export function TestCasePanel({ problemId }: { problemId: number }) {
  const queryClient = useQueryClient();
  const [input, setInput] = useState('');
  const [expectedOutput, setExpectedOutput] = useState('');
  const [isSample, setIsSample] = useState(false);
  const { data, isLoading } = useQuery({
    queryKey: ['testcases', problemId],
    queryFn: () => listTestCases(problemId),
  });
  const mutation = useMutation({
    mutationFn: () => createTestCase(problemId, { input, expected_output: expectedOutput, is_sample: isSample }),
    onSuccess: () => {
      setInput('');
      setExpectedOutput('');
      setIsSample(false);
      void queryClient.invalidateQueries({ queryKey: ['testcases', problemId] });
    },
  });

  return (
    <Card>
      <h2>测试点</h2>
      {isLoading ? <LoadingBlock /> : null}
      {!isLoading && data?.length === 0 ? <EmptyState title="暂无测试点" /> : null}
      <div className="stack">
        {data?.map((testcase) => (
          <div className="testcase-row" key={testcase.id}>
            <strong>#{testcase.id}</strong>
            <span>{testcase.is_sample ? 'sample' : 'hidden'}</span>
            <code>{testcase.input.slice(0, 80)}</code>
          </div>
        ))}
      </div>
      <form
        className="form-grid"
        onSubmit={(event) => {
          event.preventDefault();
          mutation.mutate();
        }}
      >
        <Field label="输入">
          <TextArea value={input} onChange={(event) => setInput(event.target.value)} required />
        </Field>
        <Field label="期望输出">
          <TextArea value={expectedOutput} onChange={(event) => setExpectedOutput(event.target.value)} required />
        </Field>
        <label className="checkbox-row">
          <input type="checkbox" checked={isSample} onChange={(event) => setIsSample(event.target.checked)} />
          样例
        </label>
        <Button type="submit" variant="secondary" disabled={mutation.isPending}>
          添加测试点
        </Button>
      </form>
    </Card>
  );
}
