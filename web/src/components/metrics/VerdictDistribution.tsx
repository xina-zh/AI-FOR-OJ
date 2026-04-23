import type { VerdictDistribution as Distribution } from '../../api/types';

const labels: Array<[keyof Distribution, string]> = [
  ['ac_count', 'AC'],
  ['wa_count', 'WA'],
  ['ce_count', 'CE'],
  ['re_count', 'RE'],
  ['tle_count', 'TLE'],
  ['unjudgeable_count', 'UNJUDGEABLE'],
];

export function VerdictDistribution({ distribution }: { distribution: Distribution }) {
  return (
    <div className="distribution">
      {labels.map(([key, label]) => (
        <span key={key}>
          {label} <strong>{distribution[key] ?? 0}</strong>
        </span>
      ))}
    </div>
  );
}
