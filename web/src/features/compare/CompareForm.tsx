import { useState } from 'react';

import type { CompareExperimentRequest } from '../../api/experimentApi';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { Field } from '../../components/ui/Field';
import { ProblemChecklist } from '../variables/ProblemChecklist';
import { CompareVariableForm, type CompareVariables } from './CompareVariableForm';

interface CompareFormProps {
  isSubmitting: boolean;
  onSubmit: (input: CompareExperimentRequest) => void;
}

const emptyVariables: CompareVariables = {
  baseline_model: '',
  candidate_model: '',
  baseline_prompt_name: '',
  candidate_prompt_name: '',
  baseline_agent_name: '',
  candidate_agent_name: '',
  baseline_tooling_config: '',
  candidate_tooling_config: '',
};

export function CompareForm({ isSubmitting, onSubmit }: CompareFormProps) {
  const [name, setName] = useState('');
  const [problemIDs, setProblemIDs] = useState<number[]>([]);
  const [variables, setVariables] = useState<CompareVariables>(emptyVariables);

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
            ...variables,
          });
        }}
      >
        <Field label="实验名称" hint="留空时后端会生成名称">
          <input className="input" value={name} onChange={(event) => setName(event.target.value)} />
        </Field>
        <ProblemChecklist value={problemIDs} onChange={setProblemIDs} />
        <CompareVariableForm value={variables} onChange={setVariables} />
        <Button type="submit" variant="primary" disabled={problemIDs.length === 0 || isSubmitting}>
          执行 Compare
        </Button>
      </form>
    </Card>
  );
}
