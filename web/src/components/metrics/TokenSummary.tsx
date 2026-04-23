import { formatTokens } from '../../lib/format';

interface TokenSummaryProps {
  input?: number;
  output?: number;
}

export function TokenSummary({ input = 0, output = 0 }: TokenSummaryProps) {
  return (
    <div className="metric-row">
      <span>Input {formatTokens(input)}</span>
      <span>Output {formatTokens(output)}</span>
      <strong>Total {formatTokens(input + output)}</strong>
    </div>
  );
}
