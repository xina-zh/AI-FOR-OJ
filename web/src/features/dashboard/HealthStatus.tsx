import { useQuery } from '@tanstack/react-query';

import { getHealth } from '../../api/healthApi';
import { Badge } from '../../components/ui/Badge';

export function HealthStatus() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['health'],
    queryFn: getHealth,
  });

  if (isLoading) {
    return <Badge tone="info">检查中</Badge>;
  }
  if (error) {
    return <Badge tone="danger">后端不可用</Badge>;
  }
  return <Badge tone={data?.status === 'ok' ? 'success' : 'warning'}>{data?.status ?? 'unknown'}</Badge>;
}
