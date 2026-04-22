import { NavLink, Route, Routes, useParams } from 'react-router-dom';

type RouteSummary = {
  label: string;
  path: string;
  section: string;
};

const navItems: RouteSummary[] = [
  { label: 'Dashboard', path: '/', section: 'Overview' },
  { label: 'Problems', path: '/problems', section: 'Library' },
  { label: 'Solve', path: '/solve', section: 'AI Solve' },
  { label: 'Experiments', path: '/experiments', section: 'Batch' },
  { label: 'Compare', path: '/compare', section: 'Analysis' },
  { label: 'Repeat', path: '/repeat', section: 'Stability' },
  { label: 'Submissions', path: '/submissions', section: 'Judge' },
];

function RoutePanel({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <section className="route-panel" aria-labelledby="page-title">
      <div>
        <p className="section-label">{subtitle}</p>
        <h1 id="page-title">{title}</h1>
      </div>
      <div className="metric-grid" aria-label="summary metrics">
        <div className="metric-tile">
          <span>Status</span>
          <strong>Ready</strong>
        </div>
        <div className="metric-tile">
          <span>Scope</span>
          <strong>Console</strong>
        </div>
        <div className="metric-tile">
          <span>Mode</span>
          <strong>Local</strong>
        </div>
      </div>
    </section>
  );
}

function DetailPanel({ kind }: { kind: string }) {
  const { id } = useParams();
  return <RoutePanel title={`${kind} ${id ?? ''}`.trim()} subtitle="Detail" />;
}

function TracePanel() {
  const { id } = useParams();
  return <RoutePanel title={`Experiment Run ${id ?? ''}`.trim()} subtitle="Trace" />;
}

export function App() {
  return (
    <div className="app-shell">
      <aside className="sidebar" aria-label="primary navigation">
        <div className="brand">
          <span className="brand-mark">OJ</span>
          <div>
            <strong>AI-For-Oj</strong>
            <span>Experiment Console</span>
          </div>
        </div>
        <nav className="nav-list">
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

      <main className="content">
        <Routes>
          <Route path="/" element={<RoutePanel title="Experiment Dashboard" subtitle="Overview" />} />
          <Route path="/problems" element={<RoutePanel title="Problems" subtitle="Library" />} />
          <Route path="/solve" element={<RoutePanel title="Single Solve" subtitle="AI Solve" />} />
          <Route path="/experiments" element={<RoutePanel title="Experiments" subtitle="Batch" />} />
          <Route path="/experiments/:id" element={<DetailPanel kind="Experiment" />} />
          <Route path="/compare" element={<RoutePanel title="Compare" subtitle="Analysis" />} />
          <Route path="/compare/:id" element={<DetailPanel kind="Compare" />} />
          <Route path="/repeat" element={<RoutePanel title="Repeat" subtitle="Stability" />} />
          <Route path="/repeat/:id" element={<DetailPanel kind="Repeat" />} />
          <Route path="/submissions" element={<RoutePanel title="Submissions" subtitle="Judge" />} />
          <Route path="/trace/experiment-runs/:id" element={<TracePanel />} />
        </Routes>
      </main>
    </div>
  );
}
