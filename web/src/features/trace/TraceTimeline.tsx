import type { TraceEvent } from '../../api/types';
import { EmptyState } from '../../components/ui/EmptyState';
import { TraceEventDetail } from './TraceEventDetail';

export function TraceTimeline({ events }: { events: TraceEvent[] }) {
  if (events.length === 0) {
    return <EmptyState title="暂无 trace" message="这个 experiment run 还没有可回放的事件。" />;
  }

  return (
    <div className="timeline">
      {events.map((event) => (
        <TraceEventDetail key={`${event.sequence_no}-${event.step_type}`} event={event} />
      ))}
    </div>
  );
}
