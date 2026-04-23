export function DiffBlock({ before, after }: { before: string; after: string }) {
  return (
    <div className="diff-block">
      <pre>
        <code>{before}</code>
      </pre>
      <pre>
        <code>{after}</code>
      </pre>
    </div>
  );
}
