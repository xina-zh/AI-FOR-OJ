import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

import { createProblem, deleteProblem, listProblems } from '../../api/problemApi';
import type { Problem } from '../../api/types';
import { Card } from '../../components/ui/Card';
import { EmptyState } from '../../components/ui/EmptyState';
import { ErrorPanel } from '../../components/ui/ErrorPanel';
import { LoadingBlock } from '../../components/ui/LoadingBlock';
import { ProblemCreateForm } from './ProblemCreateForm';
import { ProblemList } from './ProblemList';

export function ProblemsPage() {
  const queryClient = useQueryClient();
  const { data, isLoading, error } = useQuery({
    queryKey: ['problems'],
    queryFn: listProblems,
  });
  const createMutation = useMutation({
    mutationFn: createProblem,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['problems'] });
    },
  });
	  const deleteMutation = useMutation({
	    mutationFn: deleteProblem,
	    onSuccess: (_data, problemId) => {
	      void queryClient.invalidateQueries({ queryKey: ['problems'] });
	      void queryClient.invalidateQueries({ queryKey: ['problem', problemId] });
	      void queryClient.invalidateQueries({ queryKey: ['testcases', problemId] });
	      void queryClient.invalidateQueries({ queryKey: ['submissions'] });
	      void queryClient.invalidateQueries({ queryKey: ['submission-problem-stats'] });
	      void queryClient.invalidateQueries({ queryKey: ['experiments'] });
	      void queryClient.invalidateQueries({ queryKey: ['compares'] });
	      void queryClient.invalidateQueries({ queryKey: ['repeats'] });
	    },
	  });
  const deletingProblemId = deleteMutation.isPending ? (deleteMutation.variables ?? null) : null;

  function handleDelete(problem: Problem) {
    const confirmed = window.confirm(
      `永久删除题目 #${problem.id}「${problem.title}」？这会删除题目、测试点、相关提交和相关实验记录，不能恢复。`,
    );
    if (!confirmed) return;
    deleteMutation.mutate(problem.id);
  }

  return (
    <section className="page-section">
      <div className="page-heading">
        <div>
          <h1>题目</h1>
          <p>维护实验题目和测试点。</p>
        </div>
      </div>
      {isLoading ? <LoadingBlock /> : null}
      {error ? <ErrorPanel error={error} /> : null}
      {data && data.length > 0 ? (
        <ProblemList problems={data} deletingProblemId={deletingProblemId} onDelete={handleDelete} />
      ) : null}
      {data && data.length === 0 ? <EmptyState title="暂无题目" /> : null}
      {deleteMutation.error ? <ErrorPanel error={deleteMutation.error} /> : null}
      <Card>
        <h2>创建题目</h2>
        <ProblemCreateForm onSubmit={(input) => createMutation.mutate(input)} />
      </Card>
    </section>
  );
}
