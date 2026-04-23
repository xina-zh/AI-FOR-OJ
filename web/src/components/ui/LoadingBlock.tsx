export function LoadingBlock({ label = '加载中' }: { label?: string }) {
  return <div className="loading-block">{label}</div>;
}
