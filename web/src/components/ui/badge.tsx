import * as React from 'react';
import { cn } from '@/lib/utils';

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'success' | 'warning' | 'error' | 'outline';
}

function Badge({ className, variant = 'default', ...props }: BadgeProps) {
  return (
    <div
      className={cn(
        'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors',
        {
          'border-transparent bg-slate-900 text-white': variant === 'default',
          'border-green-200 bg-green-50 text-green-700': variant === 'success',
          'border-yellow-200 bg-yellow-50 text-yellow-700': variant === 'warning',
          'border-red-200 bg-red-50 text-red-700': variant === 'error',
          'border-slate-300 text-slate-700': variant === 'outline',
        },
        className
      )}
      {...props}
    />
  );
}

export { Badge };
