import type { HTMLAttributes } from 'react';

interface BadgeProps extends HTMLAttributes<HTMLSpanElement> {
  tone?: 'success' | 'danger' | 'warning' | 'neutral' | 'info';
}

export function Badge({ tone = 'neutral', className = '', ...props }: BadgeProps) {
  return <span className={`badge badge-${tone} ${className}`.trim()} {...props} />;
}
