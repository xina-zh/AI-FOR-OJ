import { Route, Routes, useParams } from 'react-router-dom';

import { Layout } from './components/Layout';
import { MetricStrip } from './components/MetricStrip';

function RoutePanel({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <section className="route-panel" aria-labelledby="page-title">
      <div>
        <p className="section-label">{subtitle}</p>
        <h1 id="page-title">{title}</h1>
      </div>
      <MetricStrip
        items={[
          { label: 'Status', value: 'Ready' },
          { label: 'Scope', value: 'Console' },
          { label: 'Mode', value: 'Local' },
        ]}
      />
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
    <Layout>
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
    </Layout>
  );
}
