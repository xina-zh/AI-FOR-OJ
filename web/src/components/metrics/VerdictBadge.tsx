import { verdictTone } from '../../lib/verdict';
import { Badge } from '../ui/Badge';

export function VerdictBadge({ verdict }: { verdict?: string }) {
  return <Badge tone={verdictTone(verdict)}>{verdict || 'UNKNOWN'}</Badge>;
}
