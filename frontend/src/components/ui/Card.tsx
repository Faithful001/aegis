import React from 'react';
import { cn } from './Button';

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  hover?: boolean;
}

export const Card: React.FC<CardProps> = ({ className, children, hover = true, ...props }) => {
  return (
    <div
      className={cn(
        'bg-surface border border-surface-border rounded-2xl p-5 shadow-glow-card transition-all duration-200',
        hover && 'hover:border-zinc-700/80 hover:bg-surface-hover',
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
};
