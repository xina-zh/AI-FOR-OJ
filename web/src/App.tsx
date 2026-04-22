import { Route, Routes } from 'react-router-dom';

import { Layout } from './components/Layout';
import { AnalyticsPage } from './features/analytics/AnalyticsPage';
import { CompareDetailPage } from './features/compare/CompareDetailPage';
import { ComparePage } from './features/compare/ComparePage';
import { DashboardPage } from './features/dashboard/DashboardPage';
import { ExperimentDetailPage } from './features/experiments/ExperimentDetailPage';
import { ExperimentsPage } from './features/experiments/ExperimentsPage';
import { ProblemDetailPage } from './features/problems/ProblemDetailPage';
import { ProblemsPage } from './features/problems/ProblemsPage';
import { RepeatDetailPage } from './features/repeat/RepeatDetailPage';
import { RepeatPage } from './features/repeat/RepeatPage';
import { SolvePage } from './features/solve/SolvePage';
import { SubmissionsPage } from './features/submissions/SubmissionsPage';
import { TracePage } from './features/trace/TracePage';

export function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/problems" element={<ProblemsPage />} />
        <Route path="/problems/:id" element={<ProblemDetailPage />} />
        <Route path="/solve" element={<SolvePage />} />
        <Route path="/experiments" element={<ExperimentsPage />} />
        <Route path="/experiments/:id" element={<ExperimentDetailPage />} />
        <Route path="/compare" element={<ComparePage />} />
        <Route path="/compare/:id" element={<CompareDetailPage />} />
        <Route path="/repeat" element={<RepeatPage />} />
        <Route path="/repeat/:id" element={<RepeatDetailPage />} />
        <Route path="/submissions" element={<SubmissionsPage />} />
        <Route path="/trace/experiment-runs/:id" element={<TracePage />} />
        <Route path="/analytics" element={<AnalyticsPage />} />
      </Routes>
    </Layout>
  );
}
