import { useMutation } from '@tanstack/react-query';
import { FormEvent, useEffect, useState } from 'react';

import type { AISolveRun } from '../../api/types';
import { DataTable } from '../../components/DataTable';
import { MetricStrip } from '../../components/MetricStrip';
import { StatusBadge } from '../../components/StatusBadge';
import { ExperimentVariableForm } from '../options/ExperimentVariableForm';
import { defaultVariables, ErrorBlock, LoadingBlock, PageHeader, useExperimentOptions } from '../shared';
import { experimentApi } from '../../api/experimentApi';

export function SolvePage() {
  const options = useExperimentOptions();
  const [problemID, setProblemID] = useState('1');
  const [variables, setVariables] = useState(defaultVariables());
  const solve = useMutation({ mutationFn: () => experimentApi.solveAI({ problem_id: Number(problemID), ...variables }) });

  useEffect(() => {
    if (options.data) {
      setVariables(defaultVariables(options.data));
    }
  }, [options.data]);

  function submit(event: FormEvent) {
    event.preventDefault();
    solve.mutate();
  }

  return (
    <section className="route-panel">
      <PageHeader eyebrow="AI Solve" title="Single Solve" />
      {options.isLoading ? <LoadingBlock /> : null}
      {options.error ? <ErrorBlock error={options.error} /> : null}
      <form className="panel form-grid" onSubmit={submit}>
        <label>
          <span>Problem ID</span>
          <input value={problemID} onChange={(event) => setProblemID(event.target.value)} inputMode="numeric" />
        </label>
        <div className="wide-field">
          <ExperimentVariableForm
            value={variables}
            modelOptions={options.data?.models ?? [{ name: variables.model, label: variables.model }]}
            promptOptions={options.data?.prompts ?? [{ name: variables.prompt_name, label: variables.prompt_name }]}
            agentOptions={options.data?.agents ?? [{ name: variables.agent_name, label: variables.agent_name }]}
            onChange={setVariables}
          />
        </div>
        <button className="primary-button" type="submit">Run solve</button>
      </form>
      {solve.error ? <ErrorBlock error={solve.error} /> : null}
      {solve.data ? <SolveResult run={solve.data} /> : null}
    </section>
  );
}

function SolveResult({ run }: { run: AISolveRun }) {
  return (
    <div className="panel stack">
      <MetricStrip
        items={[
          { label: 'Input Tokens', value: run.token_input ?? 0 },
          { label: 'Output Tokens', value: run.token_output ?? 0 },
          { label: 'Latency', value: run.total_latency_ms ?? 0, unit: 'ms' },
        ]}
      />
      <p><StatusBadge value={run.verdict || run.status} /> Submission {run.submission_id ?? '-'}</p>
      <pre className="code-block">{run.extracted_code || run.raw_response || 'No code returned'}</pre>
      <DataTable
        rows={run.attempts ?? []}
        getRowKey={(row) => row.id}
        emptyLabel="No attempt history"
        columns={[
          { key: 'no', header: 'Attempt', render: (row) => row.attempt_no },
          { key: 'stage', header: 'Stage', render: (row) => row.stage },
          { key: 'verdict', header: 'Verdict', render: (row) => <StatusBadge value={row.verdict} /> },
          { key: 'tokens', header: 'Tokens', render: (row) => row.token_input + row.token_output },
        ]}
      />
    </div>
  );
}
