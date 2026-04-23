import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';

import { getExperimentOptions } from '../../api/metaApi';
import { Field } from '../../components/ui/Field';
import { Select } from '../../components/ui/Select';
import { ModelInput } from '../variables/ModelInput';

export interface CompareVariables {
  baseline_model: string;
  candidate_model: string;
  baseline_prompt_name: string;
  candidate_prompt_name: string;
  baseline_agent_name: string;
  candidate_agent_name: string;
  baseline_tooling_config: string;
  candidate_tooling_config: string;
}

interface CompareVariableFormProps {
  value: CompareVariables;
  onChange: (value: CompareVariables) => void;
}

export function CompareVariableForm({ value, onChange }: CompareVariableFormProps) {
  const { data } = useQuery({
    queryKey: ['experiment-options'],
    queryFn: getExperimentOptions,
  });

  useEffect(() => {
    if (!data) return;
    const defaultPrompt = data.prompts[0]?.name || 'default';
    const defaultAgent = data.agents[0]?.name || 'direct_codegen';
    const next = {
      baseline_model: value.baseline_model || data.default_model,
      candidate_model: value.candidate_model || data.default_model,
      baseline_prompt_name: value.baseline_prompt_name || defaultPrompt,
      candidate_prompt_name: value.candidate_prompt_name || defaultPrompt,
      baseline_agent_name: value.baseline_agent_name || defaultAgent,
      candidate_agent_name: value.candidate_agent_name || defaultAgent,
      baseline_tooling_config: value.baseline_tooling_config || '',
      candidate_tooling_config: value.candidate_tooling_config || '',
    };
    if (JSON.stringify(next) !== JSON.stringify(value)) {
      onChange(next);
    }
  }, [data, onChange, value]);

  return (
    <div className="comparison-columns">
      <div className="stack">
        <h2>Baseline</h2>
        <Field label="Baseline Model">
          <ModelInput value={value.baseline_model || data?.default_model || ''} onChange={(baseline_model) => onChange({ ...value, baseline_model })} />
        </Field>
        <Field label="Baseline Prompt">
          <Select value={value.baseline_prompt_name || data?.prompts[0]?.name || ''} onChange={(event) => onChange({ ...value, baseline_prompt_name: event.target.value })}>
            {data?.prompts.map((option) => (
              <option key={option.name} value={option.name}>
                {option.label}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Baseline Agent">
          <Select value={value.baseline_agent_name || data?.agents[0]?.name || ''} onChange={(event) => onChange({ ...value, baseline_agent_name: event.target.value })}>
            {data?.agents.map((option) => (
              <option key={option.name} value={option.name}>
                {option.label}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Baseline Tooling">
          <Select value={value.baseline_tooling_config} onChange={(event) => onChange({ ...value, baseline_tooling_config: event.target.value })}>
            <option value="">none</option>
            {data?.tooling_options.map((option) => (
              <option key={option.name} value={option.name}>
                {option.label}
              </option>
            ))}
          </Select>
        </Field>
      </div>
      <div className="stack">
        <h2>Candidate</h2>
        <Field label="Candidate Model">
          <ModelInput value={value.candidate_model || data?.default_model || ''} onChange={(candidate_model) => onChange({ ...value, candidate_model })} />
        </Field>
        <Field label="Candidate Prompt">
          <Select value={value.candidate_prompt_name || data?.prompts[0]?.name || ''} onChange={(event) => onChange({ ...value, candidate_prompt_name: event.target.value })}>
            {data?.prompts.map((option) => (
              <option key={option.name} value={option.name}>
                {option.label}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Candidate Agent">
          <Select value={value.candidate_agent_name || data?.agents[0]?.name || ''} onChange={(event) => onChange({ ...value, candidate_agent_name: event.target.value })}>
            {data?.agents.map((option) => (
              <option key={option.name} value={option.name}>
                {option.label}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Candidate Tooling">
          <Select value={value.candidate_tooling_config} onChange={(event) => onChange({ ...value, candidate_tooling_config: event.target.value })}>
            <option value="">none</option>
            {data?.tooling_options.map((option) => (
              <option key={option.name} value={option.name}>
                {option.label}
              </option>
            ))}
          </Select>
        </Field>
      </div>
    </div>
  );
}
