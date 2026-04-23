interface EmptyStateProps {
  title: string;
  message?: string;
}

export function EmptyState({ title, message }: EmptyStateProps) {
  return (
    <div className="empty-state">
      <strong>{title}</strong>
      {message ? <p>{message}</p> : null}
    </div>
  );
}
