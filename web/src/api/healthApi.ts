import { request } from './http';
import type { HealthResponse } from './types';

export function getHealth() {
  return request<HealthResponse>('/health');
}
