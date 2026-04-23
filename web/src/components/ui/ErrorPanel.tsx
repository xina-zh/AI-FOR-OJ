interface ErrorPanelProps {
  title?: string;
  error: unknown;
}

export function ErrorPanel({ title = '请求失败', error }: ErrorPanelProps) {
  const message = error instanceof Error ? error.message : String(error);
  return (
    <div className="error-panel" role="alert">
      <strong>{title}</strong>
      <p>{message}</p>
    </div>
  );
}
