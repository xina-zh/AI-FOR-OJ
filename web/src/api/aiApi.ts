import { jsonBody, request } from './http';
import type { AISolveResponse, AISolveRun } from './types';

export interface SolveRequest {
  problem_id: number;
  model: string;
  prompt_name: string;
  agent_name: string;
}

export function solveProblem(input: SolveRequest) {
  return request<AISolveResponse>('/api/v1/ai/solve', {
    method: 'POST',
    body: jsonBody(input),
  });
}

export function getAISolveRun(runId: number) {
  return request<AISolveRun>(`/api/v1/ai/solve-runs/${runId}`);
}
