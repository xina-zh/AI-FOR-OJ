import type { ReactNode } from 'react';
import { NavLink } from 'react-router-dom';

export type NavItem = {
  label: string;
  path: string;
  section: string;
};

export const defaultNavItems: NavItem[] = [
  { label: 'Dashboard', path: '/', section: 'Overview' },
  { label: 'Problems', path: '/problems', section: 'Library' },
  { label: 'Solve', path: '/solve', section: 'AI Solve' },
  { label: 'Experiments', path: '/experiments', section: 'Batch' },
  { label: 'Compare', path: '/compare', section: 'Analysis' },
  { label: 'Repeat', path: '/repeat', section: 'Stability' },
  { label: 'Submissions', path: '/submissions', section: 'Judge' },
  { label: 'Analytics', path: '/analytics', section: 'Metrics' },
];

export function Layout({
  children,
  navItems = defaultNavItems,
}: {
  children: ReactNode;
  navItems?: NavItem[];
}) {
  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand">
          <span className="brand-mark">OJ</span>
          <div>
            <strong>AI-For-Oj</strong>
            <span>Experiment Console</span>
          </div>
        </div>
        <nav className="nav-list" aria-label="primary navigation">
          {navItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              end={item.path === '/'}
              className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}
            >
              <span>{item.label}</span>
              <small>{item.section}</small>
            </NavLink>
          ))}
        </nav>
      </aside>
      <main className="content">{children}</main>
    </div>
  );
}
