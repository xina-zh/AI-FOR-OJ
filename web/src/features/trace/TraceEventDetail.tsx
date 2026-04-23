import type { TraceEvent } from '../../api/types';
import { CodeBlock } from '../../components/code/CodeBlock';
import { Card } from '../../components/ui/Card';

export function TraceEventDetail({ event }: { event: TraceEvent }) {
  const isCode = event.step_type === 'extracted_code';
  const isJSON = event.content.trim().startsWith('{') || event.content.trim().startsWith('[');
  const title = event.title || traceEventTitle(event.step_type);

  return (
    <Card>
      <div className="result-header">
        <div>
          <span className="eyebrow">Step {event.sequence_no}</span>
          <h2>{title}</h2>
        </div>
        <span className="badge badge-neutral">{event.step_type}</span>
      </div>
      <CodeBlock code={event.content} language={isCode ? 'cpp' : isJSON ? 'json' : 'text'} />
      {event.metadata ? <CodeBlock code={event.metadata} language="json" /> : null}
    </Card>
  );
}

function traceEventTitle(stepType: string) {
  return stepType
    .split('_')
    .filter(Boolean)
    .map((part) => part[0].toUpperCase() + part.slice(1))
    .join(' ');
}
