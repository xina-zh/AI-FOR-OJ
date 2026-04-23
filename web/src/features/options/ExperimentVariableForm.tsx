import type { OptionItem } from '../../api/types';

export type ExperimentVariables = {
  model: string;
  prompt_name: string;
  agent_name: string;
  tooling_config: string;
};

export function ExperimentVariableForm({
  value,
  modelOptions,
  promptOptions,
  agentOptions,
  onChange,
}: {
  value: ExperimentVariables;
  modelOptions: OptionItem[];
  promptOptions: OptionItem[];
  agentOptions: OptionItem[];
  onChange: (value: ExperimentVariables) => void;
}) {
  const update = (patch: Partial<ExperimentVariables>) => onChange({ ...value, ...patch });

  return (
    <fieldset className="variable-form">
      <label>
        <span>Model</span>
        <select value={value.model} onChange={(event) => update({ model: event.target.value })}>
          {modelOptions.map((option) => (
            <option value={option.name} key={option.name}>
              {option.label}
            </option>
          ))}
        </select>
      </label>
      <label>
        <span>Prompt</span>
        <select value={value.prompt_name} onChange={(event) => update({ prompt_name: event.target.value })}>
          {promptOptions.map((option) => (
            <option value={option.name} key={option.name}>
              {option.label}
            </option>
          ))}
        </select>
      </label>
      <label>
        <span>Agent</span>
        <select value={value.agent_name} onChange={(event) => update({ agent_name: event.target.value })}>
          {agentOptions.map((option) => (
            <option value={option.name} key={option.name}>
              {option.label}
            </option>
          ))}
        </select>
      </label>
      <label className="tooling-field">
        <span>Tooling JSON</span>
        <textarea
          value={value.tooling_config}
          onChange={(event) => update({ tooling_config: event.target.value })}
          spellCheck={false}
          rows={5}
        />
      </label>
    </fieldset>
  );
}
