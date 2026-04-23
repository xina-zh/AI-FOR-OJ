import { useMemo } from 'react';

import { readJSON, writeJSON } from '../../lib/storage';

const storageKey = 'ai-for-oj:recent-models';

export function rememberModel(model: string) {
  const trimmed = model.trim();
  if (!trimmed) return;
  const recent = readJSON<string[]>(storageKey, []);
  writeJSON(storageKey, [trimmed, ...recent.filter((item) => item !== trimmed)].slice(0, 8));
}

export function ModelInput({ value, onChange }: { value: string; onChange: (value: string) => void }) {
  const recent = useMemo(() => readJSON<string[]>(storageKey, []), []);

  return (
    <>
      <input
        className="input"
        list="recent-models"
        value={value}
        onChange={(event) => onChange(event.target.value)}
        onBlur={() => rememberModel(value)}
      />
      <datalist id="recent-models">
        {recent.map((model) => (
          <option key={model} value={model} />
        ))}
      </datalist>
    </>
  );
}
