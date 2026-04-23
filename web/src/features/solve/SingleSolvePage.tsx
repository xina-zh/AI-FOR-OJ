import { useMutation } from '@tanstack/react-query';
import { useState } from 'react';

import { solveProblem } from '../../api/aiApi';
import { Card } from '../../components/ui/Card';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { Button } from '../../components/ui/Button';
import { ExperimentVariableForm, type ExperimentVariables } from '../variables/ExperimentVariableForm';
import { ProblemPicker } from '../variables/ProblemPicker';
import { SolveResultPanel } from './SolveResultPanel';

export function SingleSolvePage() {
  const [problemIDs, setProblemIDs] = useState<number[]>([]);
  const [variables, setVariables] = useState<ExperimentVariables>({
    model: '',
    prompt_name: '',
    agent_name: '',
  });
  const mutation = useMutation({
    mutationFn: () =>
      solveProblem({
        problem_id: problemIDs[0],
        model: variables.model,
        prompt_name: variables.prompt_name,
        agent_name: variables.agent_name,
      }),
  });

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>单题 Solve</h1>
          <p>选择一道题和一组变量，直接运行 AI solve。</p>
        </div>
      </div>
      <Card>
        <form
          className="form-grid"
          onSubmit={(event) => {
            event.preventDefault();
            if (problemIDs[0]) {
              mutation.mutate();
            }
          }}
        >
          <ProblemPicker value={problemIDs} onChange={setProblemIDs} />
          <ExperimentVariableForm value={variables} onChange={setVariables} />
          <Button type="submit" variant="primary" disabled={!problemIDs[0] || mutation.isPending}>
            执行 Solve
          </Button>
        </form>
      </Card>
      {mutation.error ? <ErrorPanel error={mutation.error} /> : null}
      {mutation.data ? <SolveResultPanel result={mutation.data} /> : null}
    </section>
  );
}
