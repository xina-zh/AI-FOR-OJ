import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { FormEvent, useState } from 'react';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { DetailLink, ErrorBlock, LoadingBlock, PageHeader } from '../shared';

export function ProblemsPage() {
  const queryClient = useQueryClient();
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [inputSpec, setInputSpec] = useState('');
  const [outputSpec, setOutputSpec] = useState('');
  const problems = useQuery({ queryKey: ['problems'], queryFn: () => experimentApi.listProblems() });
  const createProblem = useMutation({
    mutationFn: () => experimentApi.createProblem({
      title,
      description,
      input_spec: inputSpec,
      output_spec: outputSpec,
      samples: '',
      time_limit_ms: 1000,
      memory_limit_mb: 256,
      difficulty: 'unknown',
      tags: '',
    }),
    onSuccess: () => {
      setTitle('');
      setDescription('');
      setInputSpec('');
      setOutputSpec('');
      queryClient.invalidateQueries({ queryKey: ['problems'] });
    },
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    createProblem.mutate();
  }

  return (
    <section className="route-panel">
      <PageHeader eyebrow="Library" title="Problems" />
      <form className="panel form-grid" onSubmit={submit}>
        <label>
          <span>Title</span>
          <input value={title} onChange={(event) => setTitle(event.target.value)} />
        </label>
        <label className="wide-field">
          <span>Description</span>
          <textarea value={description} onChange={(event) => setDescription(event.target.value)} rows={4} />
        </label>
        <label>
          <span>Input Spec</span>
          <textarea value={inputSpec} onChange={(event) => setInputSpec(event.target.value)} rows={3} />
        </label>
        <label>
          <span>Output Spec</span>
          <textarea value={outputSpec} onChange={(event) => setOutputSpec(event.target.value)} rows={3} />
        </label>
        <button className="primary-button" type="submit">Create problem</button>
      </form>
      {createProblem.error ? <ErrorBlock error={createProblem.error} /> : null}
      {problems.isLoading ? <LoadingBlock /> : null}
      {problems.error ? <ErrorBlock error={problems.error} /> : null}
      <DataTable
        rows={problems.data ?? []}
        getRowKey={(row) => row.id}
        columns={[
          { key: 'id', header: 'ID', render: (row) => row.id },
          { key: 'title', header: 'Title', render: (row) => <DetailLink to={`/problems/${row.id}`}>{row.title}</DetailLink> },
          { key: 'difficulty', header: 'Difficulty', render: (row) => row.difficulty || '-' },
          { key: 'limits', header: 'Limits', render: (row) => `${row.time_limit_ms}ms / ${row.memory_limit_mb}MB` },
        ]}
      />
    </section>
  );
}
