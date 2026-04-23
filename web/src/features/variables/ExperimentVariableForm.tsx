import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getExperimentOptions } from '../../api/metaApi';
import { Field } from '../../components/ui/Field';
import { Select } from '../../components/ui/Select';
import { ModelInput } from './ModelInput';

export interface ExperimentVariables {
  model: string;
  prompt_name: string;
  agent_name: string;
}

interface ExperimentVariableFormProps {
  value: ExperimentVariables;
  onChange: (value: ExperimentVariables) => void;
}

export function ExperimentVariableForm({ value, onChange }: ExperimentVariableFormProps) {
  const { data } = useQuery({
    queryKey: ['experiment-options'],
    queryFn: getExperimentOptions,
  });

  useEffect(() => {
    if (!data) return;
    const next = {
      model: value.model || data.default_model,
      prompt_name: value.prompt_name || data.prompts[0]?.name || 'default',
      agent_name: value.agent_name || data.agents[0]?.name || 'direct_codegen',
    };
    if (next.model !== value.model || next.prompt_name !== value.prompt_name || next.agent_name !== value.agent_name) {
      onChange(next);
    }
  }, [data, onChange, value]);

  return (
    <div className="variable-grid">
      <Field label="Model">
        <ModelInput value={value.model || data?.default_model || ''} onChange={(model) => onChange({ ...value, model })} />
      </Field>
      <Field label="Prompt">
        <Select value={value.prompt_name || data?.prompts[0]?.name || ''} onChange={(event) => onChange({ ...value, prompt_name: event.target.value })}>
          {data?.prompts.map((option) => (
            <option key={option.name} value={option.name}>
              {option.label}
            </option>
          ))}
        </Select>
      </Field>
      <Field label="Agent">
        <Select value={value.agent_name || data?.agents[0]?.name || ''} onChange={(event) => onChange({ ...value, agent_name: event.target.value })}>
          {data?.agents.map((option) => (
            <option key={option.name} value={option.name}>
              {option.label}
            </option>
          ))}
        </Select>
      </Field>
    </div>
  );
}
