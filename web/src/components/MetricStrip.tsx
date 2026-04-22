export type MetricItem = {
  label: string;
  value: string | number;
  unit?: string;
};

export function MetricStrip({ items }: { items: MetricItem[] }) {
  return (
    <div className="metric-strip" aria-label="metrics">
      {items.map((item) => (
        <div className="metric-tile compact" key={item.label}>
          <span>{item.label}</span>
          <strong>
            {item.value}
            {item.unit ? <small>{item.unit}</small> : null}
          </strong>
        </div>
      ))}
    </div>
  );
}
