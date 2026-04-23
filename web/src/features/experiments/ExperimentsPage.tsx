import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { FormEvent, useEffect, useState } from 'react';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { StatusBadge } from '../../components/StatusBadge';
import { ExperimentVariableForm } from '../options/ExperimentVariableForm';
import { defaultVariables, DetailLink, idList, PageHeader, useExperimentOptions } from '../shared';

export function ExperimentsPage() {
  const queryClient = useQueryClient();
  const options = useExperimentOptions();
  const history = useQuery({ queryKey: ['experiments', 1], queryFn: () => experimentApi.listExperiments(1, 20) });
  const [name, setName] = useState('');
  const [problemIDs, setProblemIDs] = useState('1,2');
  const [variables, setVariables] = useState(defaultVariables());
  const run = useMutation({
    mutationFn: () => experimentApi.runExperiment({ name, problem_ids: idList(problemIDs), ...variables }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['experiments'] }),
  });

  useEffect(() => {
    if (options.data) setVariables(defaultVariables(options.data));
  }, [options.data]);

  function submit(event: FormEvent) {
    event.preventDefault();
    run.mutate();
  }

  return (
    <section className="route-panel">
      <PageHeader eyebrow="Batch" title="Experiments" />
      <form className="panel form-grid" onSubmit={submit}>
        <label><span>Name</span><input value={name} onChange={(event) => setName(event.target.value)} /></label>
        <label><span>Problem IDs</span><input value={problemIDs} onChange={(event) => setProblemIDs(event.target.value)} /></label>
        <div className="wide-field">
          <ExperimentVariableForm
            value={variables}
            modelOptions={options.data?.models ?? [{ name: variables.model, label: variables.model }]}
            promptOptions={options.data?.prompts ?? [{ name: variables.prompt_name, label: variables.prompt_name }]}
            agentOptions={options.data?.agents ?? [{ name: variables.agent_name, label: variables.agent_name }]}
            onChange={setVariables}
          />
        </div>
        <button className="primary-button" type="submit">Run experiment</button>
      </form>
      <DataTable
        rows={history.data?.items ?? []}
        getRowKey={(row) => row.id}
        columns={[
          { key: 'id', header: 'ID', render: (row) => <DetailLink to={`/experiments/${row.id}`}>{row.id}</DetailLink> },
          { key: 'name', header: 'Name', render: (row) => row.name },
          { key: 'status', header: 'Status', render: (row) => <StatusBadge value={row.status} /> },
          { key: 'ac', header: 'AC', render: (row) => `${row.ac_count}/${row.total_count}` },
        ]}
      />
    </section>
  );
}
