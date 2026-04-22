import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { FormEvent, useEffect, useState } from 'react';

import { experimentApi } from '../../api/experimentApi';
import { DataTable } from '../../components/DataTable';
import { StatusBadge } from '../../components/StatusBadge';
import { ExperimentVariableForm } from '../options/ExperimentVariableForm';
import { defaultVariables, DetailLink, idList, PageHeader, useExperimentOptions } from '../shared';

export function ComparePage() {
  const queryClient = useQueryClient();
  const options = useExperimentOptions();
  const history = useQuery({ queryKey: ['compares', 1], queryFn: () => experimentApi.listCompares(1, 20) });
  const [name, setName] = useState('compare');
  const [problemIDs, setProblemIDs] = useState('1,2');
  const [baseline, setBaseline] = useState(defaultVariables());
  const [candidate, setCandidate] = useState(defaultVariables());
  const compare = useMutation({
    mutationFn: () => experimentApi.compareExperiments({
      name,
      problem_ids: idList(problemIDs),
      baseline_model: baseline.model,
      candidate_model: candidate.model,
      baseline_prompt_name: baseline.prompt_name,
      candidate_prompt_name: candidate.prompt_name,
      baseline_agent_name: baseline.agent_name,
      candidate_agent_name: candidate.agent_name,
      baseline_tooling_config: baseline.tooling_config,
      candidate_tooling_config: candidate.tooling_config,
    }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['compares'] }),
  });

  useEffect(() => {
    if (options.data) {
      setBaseline(defaultVariables(options.data));
      setCandidate(defaultVariables(options.data));
    }
  }, [options.data]);

  function submit(event: FormEvent) {
    event.preventDefault();
    compare.mutate();
  }

  return (
    <section className="route-panel">
      <PageHeader eyebrow="Analysis" title="Compare" />
      <form className="panel form-grid" onSubmit={submit}>
        <label><span>Name</span><input value={name} onChange={(event) => setName(event.target.value)} /></label>
        <label><span>Problem IDs</span><input value={problemIDs} onChange={(event) => setProblemIDs(event.target.value)} /></label>
        <div className="split-grid wide-field">
          <div className="stack">
            <h2>Baseline</h2>
            <ExperimentVariableForm
              value={baseline}
              modelOptions={options.data?.models ?? [{ name: baseline.model, label: baseline.model }]}
              promptOptions={options.data?.prompts ?? [{ name: baseline.prompt_name, label: baseline.prompt_name }]}
              agentOptions={options.data?.agents ?? [{ name: baseline.agent_name, label: baseline.agent_name }]}
              onChange={setBaseline}
            />
          </div>
          <div className="stack">
            <h2>Candidate</h2>
            <ExperimentVariableForm
              value={candidate}
              modelOptions={options.data?.models ?? [{ name: candidate.model, label: candidate.model }]}
              promptOptions={options.data?.prompts ?? [{ name: candidate.prompt_name, label: candidate.prompt_name }]}
              agentOptions={options.data?.agents ?? [{ name: candidate.agent_name, label: candidate.agent_name }]}
              onChange={setCandidate}
            />
          </div>
        </div>
        <button className="primary-button" type="submit">Run compare</button>
      </form>
      <DataTable
        rows={history.data?.items ?? []}
        getRowKey={(row) => row.id}
        columns={[
          { key: 'id', header: 'ID', render: (row) => <DetailLink to={`/compare/${row.id}`}>{row.id}</DetailLink> },
          { key: 'name', header: 'Name', render: (row) => row.name },
          { key: 'dimension', header: 'Dimension', render: (row) => row.compare_dimension },
          { key: 'status', header: 'Status', render: (row) => <StatusBadge value={row.status} /> },
        ]}
      />
    </section>
  );
}
