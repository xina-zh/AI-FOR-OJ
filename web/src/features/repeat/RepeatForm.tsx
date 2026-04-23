import { useState } from 'react';

import type { RepeatExperimentRequest } from '../../api/experimentApi';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { Field } from '../../components/ui/Field';
import { ExperimentVariableForm, type ExperimentVariables } from '../variables/ExperimentVariableForm';
import { ProblemChecklist } from '../variables/ProblemChecklist';

interface RepeatFormProps {
  isSubmitting: boolean;
  onSubmit: (input: RepeatExperimentRequest) => void;
}

export function RepeatForm({ isSubmitting, onSubmit }: RepeatFormProps) {
  const [name, setName] = useState('');
  const [problemIDs, setProblemIDs] = useState<number[]>([]);
  const [repeatCount, setRepeatCount] = useState(3);
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
            repeat_count: repeatCount,
          });
        }}
      >
        <div className="inline-fields">
          <Field label="实验名称" hint="留空时后端会生成名称">
            <input className="input" value={name} onChange={(event) => setName(event.target.value)} />
          </Field>
          <Field label="重复次数">
            <input className="input" type="number" min={1} max={10} value={repeatCount} onChange={(event) => setRepeatCount(Number(event.target.value))} />
          </Field>
        </div>
        <ProblemChecklist value={problemIDs} onChange={setProblemIDs} />
        <ExperimentVariableForm value={variables} onChange={setVariables} />
        <Button type="submit" variant="primary" disabled={problemIDs.length === 0 || isSubmitting}>
          执行 Repeat
        </Button>
      </form>
    </Card>
  );
}
