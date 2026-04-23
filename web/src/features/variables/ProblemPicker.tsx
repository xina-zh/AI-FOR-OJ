import { useQuery } from '@tanstack/react-query';

import { listProblems } from '../../api/problemApi';
import { Field } from '../../components/ui/Field';
import { Select } from '../../components/ui/Select';

export function ProblemPicker({
  value,
  onChange,
  multiple = false,
}: {
  value: number[];
  onChange: (value: number[]) => void;
  multiple?: boolean;
}) {
  const { data = [] } = useQuery({
    queryKey: ['problems'],
    queryFn: listProblems,
  });

  return (
    <Field label="题目">
      <Select
        value={multiple ? value.map(String) : String(value[0] ?? '')}
        multiple={multiple}
        onChange={(event) => {
          if (multiple) {
            onChange(Array.from(event.currentTarget.selectedOptions).map((option) => Number(option.value)));
            return;
          }
          onChange(event.target.value ? [Number(event.target.value)] : []);
        }}
      >
        {!multiple ? <option value="">选择题目</option> : null}
        {data.map((problem) => (
          <option key={problem.id} value={problem.id}>
            {problem.id} · {problem.title}
          </option>
        ))}
      </Select>
    </Field>
  );
}
