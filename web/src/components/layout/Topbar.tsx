export function Topbar() {
  return (
    <header className="topbar">
      <div>
        <span className="eyebrow">Local experiment platform</span>
        <strong>模型实验与评测控制台</strong>
      </div>
      <span className="status-dot" aria-label="local backend target">
        127.0.0.1:8080
      </span>
    </header>
  );
}
