import { request } from './http';
import type { ExperimentRunTrace } from './types';

export function getExperimentRunTrace(runId: number) {
  return request<ExperimentRunTrace>(`/api/v1/experiment-runs/${runId}/trace`);
}
