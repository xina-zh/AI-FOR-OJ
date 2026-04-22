import type { ApiClient } from './client';
import { createApiClient } from './client';
import type {
  AISolveRun,
  AISolveRequest,
  CompareExperimentRequest,
  CreateProblemRequest,
  CreateTestCaseRequest,
  Experiment,
  ExperimentCompare,
  ExperimentOptions,
  ExperimentRepeat,
  ExperimentRunTrace,
  Page,
  Problem,
  RepeatExperimentRequest,
  RunExperimentRequest,
  Submission,
  TestCase,
} from './types';

export function createExperimentApi(client: ApiClient = createApiClient()) {
  return {
    getExperimentOptions() {
      return client.get<ExperimentOptions>('/meta/experiment-options');
    },
    listProblems() {
      return client.get<Problem[]>('/problems');
    },
    getProblem(id: number) {
      return client.get<Problem>(`/problems/${id}`);
    },
    createProblem(input: CreateProblemRequest) {
      return client.post<Problem, CreateProblemRequest>('/problems', input);
    },
    listTestCases(problemID: number) {
      return client.get<TestCase[]>(`/problems/${problemID}/testcases`);
    },
    createTestCase(problemID: number, input: CreateTestCaseRequest) {
      return client.post<TestCase, CreateTestCaseRequest>(`/problems/${problemID}/testcases`, input);
    },
    listSubmissions(page = 1, pageSize = 20) {
      return client.get<Page<Submission>>(`/submissions?page=${page}&page_size=${pageSize}`);
    },
    solveAI(input: AISolveRequest) {
      return client.post<AISolveRun, AISolveRequest>('/ai/solve', input);
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
    compareExperiments(input: CompareExperimentRequest) {
      return client.post<ExperimentCompare, CompareExperimentRequest>('/experiments/compare', input);
    },
    listRepeats(page = 1, pageSize = 20) {
      return client.get<Page<ExperimentRepeat>>(`/experiments/repeat?page=${page}&page_size=${pageSize}`);
    },
    getRepeat(id: number) {
      return client.get<ExperimentRepeat>(`/experiments/repeat/${id}`);
    },
    repeatExperiment(input: RepeatExperimentRequest) {
      return client.post<ExperimentRepeat, RepeatExperimentRequest>('/experiments/repeat', input);
    },
    getExperimentRunTrace(id: number) {
      return client.get<ExperimentRunTrace>(`/experiment-runs/${id}/trace`);
    },
  };
}

export const experimentApi = createExperimentApi();
