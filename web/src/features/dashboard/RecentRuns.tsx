import { Link } from 'react-router-dom';

const actions = [
  { to: '/solve', label: '单题 Solve' },
  { to: '/experiments', label: 'Run Experiment' },
  { to: '/compare', label: 'Compare' },
  { to: '/repeat', label: 'Repeat' },
];

export function RecentRuns() {
  return (
    <div className="quick-actions">
      {actions.map((action) => (
        <Link key={action.to} to={action.to} className="action-link">
          {action.label}
        </Link>
      ))}
    </div>
  );
}
