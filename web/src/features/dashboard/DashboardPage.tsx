import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';

import { listCompares, listExperiments, listRepeats } from '../../api/experimentApi';
import { listProblems } from '../../api/problemApi';
import { getSubmissionProblemStats } from '../../api/submissionApi';
import type { CompareExperiment, Experiment, RepeatExperiment } from '../../api/types';
import { Table } from '../../components/ui/Table';
import { HealthStatus } from './HealthStatus';
import { RecentRuns } from './RecentRuns';
import { Card } from '../../components/ui/Card';

export function DashboardPage() {
  const problems = useQuery({
    queryKey: ['problems', 'dashboard'],
    queryFn: listProblems,
  });
  const submissionStats = useQuery({
    queryKey: ['submission-problem-stats', 'dashboard'],
    queryFn: getSubmissionProblemStats,
  });
  const experiments = useQuery({
    queryKey: ['experiments', 'dashboard'],
    queryFn: () => listExperiments({ pageSize: 5 }),
  });
  const compares = useQuery({
    queryKey: ['compares', 'dashboard'],
    queryFn: () => listCompares({ pageSize: 5 }),
  });
  const repeats = useQuery({
    queryKey: ['repeats', 'dashboard'],
    queryFn: () => listRepeats({ pageSize: 5 }),
  });
  const submissionCount = (submissionStats.data ?? []).reduce((sum, item) => sum + item.total_submissions, 0);

  return (
    <section className="dashboard">
      <div className="page-heading">
        <div>
          <h1>实验控制台</h1>
          <p>选择模型、prompt、agent，直接运行和回看实验结果。</p>
        </div>
        <HealthStatus />
      </div>

      <div className="dashboard-grid">
        <Card>
          <span className="eyebrow">Workflow</span>
          <h2>快速执行</h2>
          <RecentRuns />
        </Card>
        <Card>
          <span className="eyebrow">Metrics</span>
          <h2>重点指标</h2>
          <div className="metric-list">
            <MetricLine label="题目总数" value={problems.data?.length ?? 0} />
            <MetricLine label="Submission 总数" value={submissionCount} />
            <MetricLine label="最近实验" value={experiments.data?.items.length ?? 0} />
          </div>
        </Card>
        <Card>
          <span className="eyebrow">Trace</span>
          <h2>运行回放</h2>
          <p className="muted">每个 experiment run 可以查看 prompt、模型输出、代码提取和判题结果。</p>
        </Card>
      </div>

      <div className="dashboard-wide">
        <Card>
          <h2>最近运行</h2>
          <RecentRunTable experiments={experiments.data?.items ?? []} compares={compares.data?.items ?? []} repeats={repeats.data?.items ?? []} />
        </Card>
      </div>
    </section>
  );
}

function MetricLine({ label, value }: { label: string; value: number }) {
  return (
    <div className="metric-line">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function RecentRunTable({
  experiments,
  compares,
  repeats,
}: {
  experiments: Experiment[];
  compares: CompareExperiment[];
  repeats: RepeatExperiment[];
}) {
  const rows = [
    ...experiments.map((item) => ({ type: 'experiment', name: item.name, status: item.status, href: `/experiments/${item.id}` })),
    ...compares.map((item) => ({ type: 'compare', name: item.name, status: item.status, href: `/compare/${item.id}` })),
    ...repeats.map((item) => ({ type: 'repeat', name: item.name, status: item.status, href: `/repeat/${item.id}` })),
  ];

  if (rows.length === 0) {
    return <p className="muted">暂无历史运行。</p>;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Type</th>
          <th>Name</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        {rows.map((row) => (
          <tr key={`${row.type}-${row.href}`}>
            <td>{row.type}</td>
            <td>
              <Link to={row.href}>{row.name}</Link>
            </td>
            <td>{row.status}</td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
