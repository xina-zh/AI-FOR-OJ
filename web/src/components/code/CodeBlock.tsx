export function CodeBlock({ code, language = 'text' }: { code: string; language?: string }) {
  return (
    <pre className="code-block" data-language={language}>
      <code>{code}</code>
    </pre>
  );
}
