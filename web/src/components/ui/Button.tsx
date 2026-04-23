import type { ButtonHTMLAttributes, ReactNode } from 'react';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost';
  icon?: ReactNode;
}

export function Button({ variant = 'secondary', icon, children, className = '', ...props }: ButtonProps) {
  return (
    <button className={`button button-${variant} ${className}`.trim()} {...props}>
      {icon ? <span className="button-icon">{icon}</span> : null}
      <span>{children}</span>
    </button>
  );
}
