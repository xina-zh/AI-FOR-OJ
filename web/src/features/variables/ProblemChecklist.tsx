import { useQuery } from '@tanstack/react-query';

import { listProblems } from '../../api/problemApi';

interface ProblemChecklistProps {
  value: number[];
  onChange: (value: number[]) => void;
}

export function ProblemChecklist({ value, onChange }: ProblemChecklistProps) {
  const { data = [] } = useQuery({
    queryKey: ['problems'],
    queryFn: listProblems,
  });

  const selected = new Set(value);

  return (
    <fieldset className="field problem-checklist">
      <legend className="field-label">题目</legend>
      <div className="checklist-grid">
        {data.map((problem) => (
          <label key={problem.id} className="checkbox-row checklist-item">
            <input
              type="checkbox"
              checked={selected.has(problem.id)}
              onChange={(event) => {
                if (event.target.checked) {
                  onChange([...value, problem.id]);
                  return;
                }
                onChange(value.filter((id) => id !== problem.id));
              }}
            />
            <span>
              {problem.id} · {problem.title}
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}
