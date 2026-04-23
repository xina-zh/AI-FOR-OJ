import type { ReactNode } from 'react';

interface TabItem {
  id: string;
  label: string;
}

interface TabsProps {
  items: TabItem[];
  activeId: string;
  onChange: (id: string) => void;
  children?: ReactNode;
}

export function Tabs({ items, activeId, onChange, children }: TabsProps) {
  return (
    <div className="tabs">
      <div className="tab-list" role="tablist">
        {items.map((item) => (
          <button
            key={item.id}
            className={item.id === activeId ? 'tab active' : 'tab'}
            type="button"
            role="tab"
            aria-selected={item.id === activeId}
            onClick={() => onChange(item.id)}
          >
            {item.label}
          </button>
        ))}
      </div>
      {children}
    </div>
  );
}
