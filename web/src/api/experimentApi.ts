import { jsonBody, request } from './http';
import type { CompareExperiment, Experiment, Page, RepeatExperiment } from './types';

interface PageInput {
  page?: number;
  pageSize?: number;
}

export interface RunExperimentRequest {
  name: string;
  problem_ids: number[];
  model: string;
  prompt_name: string;
  agent_name: string;
  tooling_config: string;
}

export interface CompareExperimentRequest {
  name: string;
  problem_ids: number[];
  baseline_model: string;
  candidate_model: string;
  baseline_prompt_name: string;
  candidate_prompt_name: string;
  baseline_agent_name: string;
  candidate_agent_name: string;
  baseline_tooling_config: string;
  candidate_tooling_config: string;
}

export interface RepeatExperimentRequest {
  name: string;
  problem_ids: number[];
  model: string;
  prompt_name: string;
  agent_name: string;
  tooling_config: string;
  repeat_count: number;
}

export function listExperiments(input: PageInput = {}) {
  return request<Page<Experiment>>(`/api/v1/experiments${pageQuery(input)}`);
}

export function getExperiment(experimentId: number) {
  return request<Experiment>(`/api/v1/experiments/${experimentId}`);
}

export function runExperiment(input: RunExperimentRequest) {
  return request<Experiment>('/api/v1/experiments/run', {
    method: 'POST',
    body: jsonBody(input),
  });
}

export function listCompares(input: PageInput = {}) {
  return request<Page<CompareExperiment>>(`/api/v1/experiments/compare${pageQuery(input)}`);
}

export function getCompare(compareId: number) {
  return request<CompareExperiment>(`/api/v1/experiments/compare/${compareId}`);
}

export function compareExperiments(input: CompareExperimentRequest) {
  return request<CompareExperiment>('/api/v1/experiments/compare', {
    method: 'POST',
    body: jsonBody(input),
  });
}

export function listRepeats(input: PageInput = {}) {
  return request<Page<RepeatExperiment>>(`/api/v1/experiments/repeat${pageQuery(input)}`);
}

export function getRepeat(repeatId: number) {
  return request<RepeatExperiment>(`/api/v1/experiments/repeat/${repeatId}`);
}

export function repeatExperiment(input: RepeatExperimentRequest) {
  return request<RepeatExperiment>('/api/v1/experiments/repeat', {
    method: 'POST',
    body: jsonBody(input),
  });
}

function pageQuery(input: PageInput) {
  const params = new URLSearchParams();
  if (input.page) {
    params.set('page', String(input.page));
  }
  if (input.pageSize) {
    params.set('page_size', String(input.pageSize));
  }
  const query = params.toString();
  return query ? `?${query}` : '';
}
