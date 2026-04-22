import type { ApiClient } from './client';
import { createApiClient } from './client';
import type {
  AISolveRun,
  Experiment,
  ExperimentCompare,
  ExperimentOptions,
  ExperimentRepeat,
  ExperimentRunTrace,
  Page,
  Problem,
  RunExperimentRequest,
  Submission,
} from './types';

export function createExperimentApi(client: ApiClient = createApiClient()) {
  return {
    getExperimentOptions() {
      return client.get<ExperimentOptions>('/meta/experiment-options');
    },
    listProblems() {
      return client.get<Problem[]>('/problems');
    },
    listSubmissions(page = 1, pageSize = 20) {
      return client.get<Page<Submission>>(`/submissions?page=${page}&page_size=${pageSize}`);
    },
    getAISolveRun(id: number) {
      return client.get<AISolveRun>(`/ai/solve-runs/${id}`);
    },
    runExperiment(input: RunExperimentRequest) {
      return client.post<Experiment, RunExperimentRequest>('/experiments/run', input);
    },
    listExperiments(page = 1, pageSize = 20) {
      return client.get<Page<Experiment>>(`/experiments?page=${page}&page_size=${pageSize}`);
    },
    getExperiment(id: number) {
      return client.get<Experiment>(`/experiments/${id}`);
    },
    listCompares(page = 1, pageSize = 20) {
      return client.get<Page<ExperimentCompare>>(`/experiments/compare?page=${page}&page_size=${pageSize}`);
    },
    getCompare(id: number) {
      return client.get<ExperimentCompare>(`/experiments/compare/${id}`);
    },
    listRepeats(page = 1, pageSize = 20) {
      return client.get<Page<ExperimentRepeat>>(`/experiments/repeat?page=${page}&page_size=${pageSize}`);
    },
    getRepeat(id: number) {
      return client.get<ExperimentRepeat>(`/experiments/repeat/${id}`);
    },
    getExperimentRunTrace(id: number) {
      return client.get<ExperimentRunTrace>(`/experiment-runs/${id}/trace`);
    },
  };
}

export const experimentApi = createExperimentApi();
