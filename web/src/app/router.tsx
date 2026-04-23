import { createBrowserRouter } from 'react-router-dom';

import { AppShell } from '../components/layout/AppShell';
import { ComparePage } from '../features/compare/ComparePage';
import { DashboardPage } from '../features/dashboard/DashboardPage';
import { ExperimentDetailPage } from '../features/experiments/ExperimentDetailPage';
import { ExperimentRunPage } from '../features/experiments/ExperimentRunPage';
import { ProblemDetail } from '../features/problems/ProblemDetail';
import { ProblemsPage } from '../features/problems/ProblemsPage';
import { CompareDetailPage } from '../features/compare/CompareDetailPage';
import { RepeatDetailPage } from '../features/repeat/RepeatDetailPage';
import { RepeatPage } from '../features/repeat/RepeatPage';
import { SingleSolvePage } from '../features/solve/SingleSolvePage';
import { SolveRunDetail } from '../features/solve/SolveRunDetail';
import { SubmissionDetailPage } from '../features/submissions/SubmissionDetailPage';
import { SubmissionsPage } from '../features/submissions/SubmissionsPage';
import { TokenAnalyticsPage } from '../features/tokens/TokenAnalyticsPage';
import { TracePage } from '../features/trace/TracePage';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppShell />,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'problems', element: <ProblemsPage /> },
      { path: 'problems/:id', element: <ProblemDetail /> },
      { path: 'solve', element: <SingleSolvePage /> },
      { path: 'ai-runs/:id', element: <SolveRunDetail /> },
      { path: 'experiments', element: <ExperimentRunPage /> },
      { path: 'experiments/:id', element: <ExperimentDetailPage /> },
      { path: 'compare', element: <ComparePage /> },
      { path: 'compare/:id', element: <CompareDetailPage /> },
      { path: 'repeat', element: <RepeatPage /> },
      { path: 'repeat/:id', element: <RepeatDetailPage /> },
      { path: 'tokens', element: <TokenAnalyticsPage /> },
      { path: 'trace/experiment-runs/:id', element: <TracePage /> },
      { path: 'submissions', element: <SubmissionsPage /> },
      { path: 'submissions/:id', element: <SubmissionDetailPage /> },
    ],
  },
]);
