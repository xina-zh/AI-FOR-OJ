import { jsonBody, request } from './http';
import type { Problem, TestCase } from './types';

export type CreateProblemRequest = Omit<Problem, 'id'>;

export interface CreateTestCaseRequest {
  input: string;
  expected_output: string;
  is_sample: boolean;
}

export function listProblems() {
  return request<Problem[]>('/api/v1/problems');
}

export function getProblem(problemId: number) {
  return request<Problem>(`/api/v1/problems/${problemId}`);
}

export function createProblem(input: CreateProblemRequest) {
  return request<Problem>('/api/v1/problems', {
    method: 'POST',
    body: jsonBody(input),
  });
}

export function deleteProblem(problemId: number) {
  return request<void>(`/api/v1/problems/${problemId}`, {
    method: 'DELETE',
  });
}

export function listTestCases(problemId: number) {
  return request<TestCase[]>(`/api/v1/problems/${problemId}/testcases`);
}

export function createTestCase(problemId: number, input: CreateTestCaseRequest) {
  return request<TestCase>(`/api/v1/problems/${problemId}/testcases`, {
    method: 'POST',
    body: jsonBody(input),
  });
}
