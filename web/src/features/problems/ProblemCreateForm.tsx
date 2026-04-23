import { useState } from 'react';

import type { CreateProblemRequest } from '../../api/problemApi';
import { Button } from '../../components/ui/Button';
import { Field } from '../../components/ui/Field';
import { TextArea } from '../../components/ui/TextArea';

const initialForm: CreateProblemRequest = {
  title: '',
  description: '',
  input_spec: '',
  output_spec: '',
  samples: '',
  time_limit_ms: 1000,
  memory_limit_mb: 128,
  difficulty: 'easy',
  tags: '',
};

export function ProblemCreateForm({ onSubmit }: { onSubmit: (input: CreateProblemRequest) => void }) {
  const [form, setForm] = useState<CreateProblemRequest>(initialForm);

  return (
    <form
      className="form-grid"
      onSubmit={(event) => {
        event.preventDefault();
        onSubmit(form);
      }}
    >
      <Field label="标题">
        <input
          className="input"
          value={form.title}
          onChange={(event) => setForm({ ...form, title: event.target.value })}
          required
        />
      </Field>
      <Field label="难度">
        <input
          className="input"
          value={form.difficulty}
          onChange={(event) => setForm({ ...form, difficulty: event.target.value })}
          required
        />
      </Field>
      <Field label="题面">
        <TextArea
          value={form.description}
          onChange={(event) => setForm({ ...form, description: event.target.value })}
          required
        />
      </Field>
      <Field label="输入说明">
        <TextArea
          value={form.input_spec}
          onChange={(event) => setForm({ ...form, input_spec: event.target.value })}
          required
        />
      </Field>
      <Field label="输出说明">
        <TextArea
          value={form.output_spec}
          onChange={(event) => setForm({ ...form, output_spec: event.target.value })}
          required
        />
      </Field>
      <Field label="样例">
        <TextArea value={form.samples} onChange={(event) => setForm({ ...form, samples: event.target.value })} />
      </Field>
      <Field label="标签">
        <input className="input" value={form.tags} onChange={(event) => setForm({ ...form, tags: event.target.value })} />
      </Field>
      <div className="inline-fields">
        <Field label="时间限制 ms">
          <input
            className="input"
            type="number"
            value={form.time_limit_ms}
            onChange={(event) => setForm({ ...form, time_limit_ms: Number(event.target.value) })}
            required
          />
        </Field>
        <Field label="内存 MB">
          <input
            className="input"
            type="number"
            value={form.memory_limit_mb}
            onChange={(event) => setForm({ ...form, memory_limit_mb: Number(event.target.value) })}
            required
          />
        </Field>
      </div>
      <Button type="submit" variant="primary">
        创建题目
      </Button>
    </form>
  );
}
