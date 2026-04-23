import type { TableHTMLAttributes } from 'react';

export function Table(props: TableHTMLAttributes<HTMLTableElement>) {
  return <table className="table" {...props} />;
}
