import { Table } from '../../components/ui/Table';

interface TestCaseResult {
  testcase_id: number;
  index: number;
  verdict: string;
  runtime_ms: number;
  stdout?: string;
  stderr?: string;
  exit_code: number;
  timed_out: boolean;
}

export function TestCaseResultTable({ results }: { results: TestCaseResult[] }) {
  if (results.length === 0) {
    return <p className="muted">暂无 testcase 结果。</p>;
  }

  return (
    <Table>
      <thead>
        <tr>
          <th>Case</th>
          <th>Verdict</th>
          <th>Runtime</th>
          <th>Output</th>
          <th>Exit</th>
        </tr>
      </thead>
      <tbody>
        {results.map((result) => (
          <tr key={`${result.testcase_id}-${result.index}`}>
            <td>#{result.index}</td>
            <td>{result.verdict}</td>
            <td>{result.runtime_ms}ms</td>
            <td>{result.stdout || result.stderr || '-'}</td>
            <td>{result.timed_out ? 'timeout' : result.exit_code}</td>
          </tr>
        ))}
      </tbody>
    </Table>
  );
}
