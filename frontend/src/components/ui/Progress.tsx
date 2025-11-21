'use client';

import React, { forwardRef, HTMLAttributes } from 'react';
import { cn } from '@/lib/utils';

interface ProgressProps extends HTMLAttributes<HTMLDivElement> {
  value: number;
  max?: number;
  label?: string;
  showPercentage?: boolean;
  size?: 'sm' | 'md' | 'lg';
  color?: 'primary' | 'success' | 'warning' | 'error';
  striped?: boolean;
  animated?: boolean;
}

const sizeClasses: Record<'sm' | 'md' | 'lg', string> = {
  sm: 'h-1',
  md: 'h-2',
  lg: 'h-4',
};

const colorClasses: Record<'primary' | 'success' | 'warning' | 'error', string> = {
  primary: 'bg-primary-500 dark:bg-primary-600',
  success: 'bg-emerald-500 dark:bg-emerald-600',
  warning: 'bg-amber-500 dark:bg-amber-600',
  error: 'bg-red-500 dark:bg-red-600',
};

/**
 * Progress 组件
 * 显示操作进度的条形图组件
 */
export const Progress = forwardRef<HTMLDivElement, ProgressProps>(
  (
    {
      className,
      value,
      max = 100,
      label,
      showPercentage,
      size = 'md',
      color = 'primary',
      striped,
      animated,
      ...props
    },
    ref
  ) => {
    const percentage = Math.min((value / max) * 100, 100);

    return (
      <div ref={ref} className={cn('w-full', className)} {...props}>
        {(label || showPercentage) && (
          <div className="flex items-center justify-between mb-2">
            {label && <span className="text-sm font-medium text-neutral-700 dark:text-neutral-300">{label}</span>}
            {showPercentage && (
              <span className="text-sm font-medium text-neutral-600 dark:text-neutral-400">
                {Math.round(percentage)}%
              </span>
            )}
          </div>
        )}

        <div
          className={cn(
            'w-full rounded-full overflow-hidden bg-neutral-200 dark:bg-neutral-700',
            sizeClasses[size]
          )}
        >
          <div
            className={cn(
              'h-full transition-all duration-300 rounded-full',
              colorClasses[color],
              striped && 'bg-gradient-to-r from-transparent via-white to-transparent opacity-20',
              animated && striped && 'animate-pulse'
            )}
            style={{ width: `${percentage}%` }}
            role="progressbar"
            aria-valuenow={Math.round(percentage)}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label={label}
          />
        </div>
      </div>
    );
  }
);

Progress.displayName = 'Progress';

