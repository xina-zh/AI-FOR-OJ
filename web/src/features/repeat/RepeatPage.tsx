import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { FormEvent, useEffect, useState } from 'react';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { StatusBadge } from '../../components/StatusBadge';
import { ExperimentVariableForm } from '../options/ExperimentVariableForm';
import { defaultVariables, DetailLink, idList, PageHeader, useExperimentOptions } from '../shared';

export function RepeatPage() {
  const queryClient = useQueryClient();
  const options = useExperimentOptions();
  const history = useQuery({ queryKey: ['repeats', 1], queryFn: () => experimentApi.listRepeats(1, 20) });
  const [name, setName] = useState('repeat');
  const [problemIDs, setProblemIDs] = useState('1,2');
  const [repeatCount, setRepeatCount] = useState(3);
  const [variables, setVariables] = useState(defaultVariables());
  const repeat = useMutation({
    mutationFn: () => experimentApi.repeatExperiment({ name, problem_ids: idList(problemIDs), repeat_count: repeatCount, ...variables }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['repeats'] }),
  });
  useEffect(() => { if (options.data) setVariables(defaultVariables(options.data)); }, [options.data]);
  function submit(event: FormEvent) { event.preventDefault(); repeat.mutate(); }
  return (
    <section className="route-panel">
      <PageHeader eyebrow="Stability" title="Repeat" />
      <form className="panel form-grid" onSubmit={submit}>
        <label><span>Name</span><input value={name} onChange={(event) => setName(event.target.value)} /></label>
        <label><span>Problem IDs</span><input value={problemIDs} onChange={(event) => setProblemIDs(event.target.value)} /></label>
        <label><span>Repeat Count</span><input value={repeatCount} onChange={(event) => setRepeatCount(Number(event.target.value))} inputMode="numeric" /></label>
        <div className="wide-field">
          <ExperimentVariableForm
            value={variables}
            modelOptions={options.data?.models ?? [{ name: variables.model, label: variables.model }]}
            promptOptions={options.data?.prompts ?? [{ name: variables.prompt_name, label: variables.prompt_name }]}
            agentOptions={options.data?.agents ?? [{ name: variables.agent_name, label: variables.agent_name }]}
            onChange={setVariables}
          />
        </div>
        <button className="primary-button" type="submit">Run repeat</button>
      </form>
      <DataTable rows={history.data?.items ?? []} getRowKey={(row) => row.id} columns={[
        { key: 'id', header: 'ID', render: (row) => <DetailLink to={`/repeat/${row.id}`}>{row.id}</DetailLink> },
        { key: 'name', header: 'Name', render: (row) => row.name },
        { key: 'rate', header: 'AC Rate', render: (row) => `${Math.round((row.overall_ac_rate ?? 0) * 100)}%` },
        { key: 'status', header: 'Status', render: (row) => <StatusBadge value={row.status} /> },
      ]} />
    </section>
  );
}
