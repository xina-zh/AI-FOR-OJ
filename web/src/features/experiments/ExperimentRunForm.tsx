import { useState } from 'react';

import type { RunExperimentRequest } from '../../api/experimentApi';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { Field } from '../../components/ui/Field';
import { ExperimentVariableForm, type ExperimentVariables } from '../variables/ExperimentVariableForm';
import { ProblemChecklist } from '../variables/ProblemChecklist';

interface ExperimentRunFormProps {
  isSubmitting: boolean;
  onSubmit: (input: RunExperimentRequest) => void;
}

export function ExperimentRunForm({ isSubmitting, onSubmit }: ExperimentRunFormProps) {
  const [name, setName] = useState('');
  const [problemIDs, setProblemIDs] = useState<number[]>([]);
  const [variables, setVariables] = useState<ExperimentVariables>({
    model: '',
    prompt_name: '',
    agent_name: '',
  });

  return (
    <Card>
      <form
        className="form-grid"
        onSubmit={(event) => {
          event.preventDefault();
          if (problemIDs.length === 0) return;
          onSubmit({
            name,
            problem_ids: problemIDs,
            model: variables.model,
            prompt_name: variables.prompt_name,
            agent_name: variables.agent_name,
          });
        }}
      >
        <Field label="实验名称" hint="留空时后端会生成名称">
          <input className="input" value={name} onChange={(event) => setName(event.target.value)} />
        </Field>
        <ProblemChecklist value={problemIDs} onChange={setProblemIDs} />
        <ExperimentVariableForm value={variables} onChange={setVariables} />
        <Button type="submit" variant="primary" disabled={problemIDs.length === 0 || isSubmitting}>
          执行批量实验
        </Button>
      </form>
    </Card>
  );
}
