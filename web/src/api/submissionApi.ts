import { jsonBody, request } from './http';
import type { Page, SubmissionDetail, SubmissionProblemStats, SubmissionSummary } from './types';

export interface JudgeSubmissionRequest {
  problem_id: number;
  source_code: string;
  language: string;
}

export interface SubmissionListInput {
  page?: number;
  pageSize?: number;
  problemId?: number;
}

export function listSubmissions(input: SubmissionListInput = {}) {
  const params = new URLSearchParams();
  if (input.page) params.set('page', String(input.page));
  if (input.pageSize) params.set('page_size', String(input.pageSize));
  if (input.problemId) params.set('problem_id', String(input.problemId));
  const query = params.toString();
  return request<Page<SubmissionSummary>>(`/api/v1/submissions${query ? `?${query}` : ''}`);
}

export function getSubmission(submissionId: number) {
  return request<SubmissionDetail>(`/api/v1/submissions/${submissionId}`);
}

export function judgeSubmission(input: JudgeSubmissionRequest) {
  return request<SubmissionSummary>('/api/v1/submissions/judge', {
    method: 'POST',
    body: jsonBody(input),
  });
}

export function getSubmissionProblemStats() {
  return request<SubmissionProblemStats[]>('/api/v1/submissions/stats/problems');
}
