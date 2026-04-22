import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { FormEvent, useState } from 'react';
import { useParams } from 'react-router-dom';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { ErrorBlock, LoadingBlock, PageHeader } from '../shared';

export function ProblemDetailPage() {
  const id = Number(useParams().id);
  const queryClient = useQueryClient();
  const [input, setInput] = useState('');
  const [expectedOutput, setExpectedOutput] = useState('');
  const [isSample, setIsSample] = useState(true);
  const problem = useQuery({ queryKey: ['problem', id], queryFn: () => experimentApi.getProblem(id), enabled: Number.isFinite(id) });
  const testCases = useQuery({ queryKey: ['problem-testcases', id], queryFn: () => experimentApi.listTestCases(id), enabled: Number.isFinite(id) });
  const createTestCase = useMutation({
    mutationFn: () => experimentApi.createTestCase(id, { input, expected_output: expectedOutput, is_sample: isSample }),
    onSuccess: () => {
      setInput('');
      setExpectedOutput('');
      queryClient.invalidateQueries({ queryKey: ['problem-testcases', id] });
    },
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    createTestCase.mutate();
  }

  if (problem.isLoading) return <LoadingBlock />;
  if (problem.error) return <ErrorBlock error={problem.error} />;
  if (!problem.data) return null;

  return (
    <section className="route-panel">
      <PageHeader eyebrow="Problem Detail" title={problem.data.title} />
      <div className="panel stack">
        <p>{problem.data.description}</p>
        <div className="split-grid">
          <div>
            <h2>Input</h2>
            <pre className="code-block">{problem.data.input_spec}</pre>
          </div>
          <div>
            <h2>Output</h2>
            <pre className="code-block">{problem.data.output_spec}</pre>
          </div>
        </div>
      </div>
      <form className="panel form-grid" onSubmit={submit}>
        <label>
          <span>Input</span>
          <textarea value={input} onChange={(event) => setInput(event.target.value)} rows={4} />
        </label>
        <label>
          <span>Expected Output</span>
          <textarea value={expectedOutput} onChange={(event) => setExpectedOutput(event.target.value)} rows={4} />
        </label>
        <label className="checkbox-field">
          <input type="checkbox" checked={isSample} onChange={(event) => setIsSample(event.target.checked)} />
          <span>Sample</span>
        </label>
        <button className="primary-button" type="submit">Create test case</button>
      </form>
      {createTestCase.error ? <ErrorBlock error={createTestCase.error} /> : null}
      {testCases.isLoading ? <LoadingBlock /> : null}
      {testCases.error ? <ErrorBlock error={testCases.error} /> : null}
      <DataTable
        rows={testCases.data ?? []}
        getRowKey={(row) => row.id}
        columns={[
          { key: 'input', header: 'Input', render: (row) => row.input },
          { key: 'output', header: 'Expected', render: (row) => row.expected_output },
          { key: 'sample', header: 'Sample', render: (row) => (row.is_sample ? 'yes' : 'no') },
        ]}
      />
    </section>
  );
}
