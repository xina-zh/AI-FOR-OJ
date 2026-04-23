import { NavLink } from 'react-router-dom';

const links = [
  { to: '/', label: 'Dashboard' },
  { to: '/problems', label: '题目' },
  { to: '/solve', label: '单题 Solve' },
  { to: '/experiments', label: 'Experiment' },
  { to: '/compare', label: 'Compare' },
  { to: '/repeat', label: 'Repeat' },
  { to: '/tokens', label: 'Token' },
  { to: '/submissions', label: 'Submissions' },
];

export function Sidebar() {
  return (
    <aside className="sidebar" aria-label="主导航">
      <div className="brand">
        <span className="brand-mark">OJ</span>
        <span>AI-For-OJ</span>
      </div>
      <nav className="nav-list">
        {links.map((link) => (
          <NavLink
            key={link.to}
            to={link.to}
            end={link.to === '/'}
            className={({ isActive }) => (isActive ? 'nav-link active' : 'nav-link')}
          >
            {link.label}
          </NavLink>
        ))}
      </nav>
    </aside>
  );
}
