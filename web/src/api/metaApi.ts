import { request } from './http';
import type { ExperimentOptions } from './types';

export function getExperimentOptions() {
  return request<ExperimentOptions>('/api/v1/meta/experiment-options');
}
