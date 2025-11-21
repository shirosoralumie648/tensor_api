'use client';

import React, { forwardRef, HTMLAttributes } from 'react';
import { cn } from '@/lib/utils';

interface SpinnerProps extends HTMLAttributes<HTMLDivElement> {
  size?: 'sm' | 'md' | 'lg' | 'xl';
  color?: 'primary' | 'white' | 'currentColor';
  fullScreen?: boolean;
  withText?: string;
}

const sizeClasses: Record<'sm' | 'md' | 'lg' | 'xl', string> = {
  sm: 'w-6 h-6',
  md: 'w-8 h-8',
  lg: 'w-12 h-12',
  xl: 'w-16 h-16',
};

const colorClasses: Record<'primary' | 'white' | 'currentColor', string> = {
  primary: 'text-primary-500',
  white: 'text-white',
  currentColor: 'text-current',
};

/**
 * Spinner 组件
 * 显示加载动画
 */
export const Spinner = forwardRef<HTMLDivElement, SpinnerProps>(
  (
    {
      className,
      size = 'md',
      color = 'primary',
      fullScreen,
      withText,
      ...props
    },
    ref
  ) => {
    const spinner = (
      <div
        ref={ref}
        className={cn(
          'flex items-center justify-center',
          sizeClasses[size],
          className
        )}
        {...props}
      >
        <svg
          className={cn('animate-spin', colorClasses[color], sizeClasses[size])}
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
        >
          <circle
            className="opacity-25"
            cx="12"
            cy="12"
            r="10"
            stroke="currentColor"
            strokeWidth="4"
          />
          <path
            className="opacity-75"
            fill="currentColor"
            d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
          />
        </svg>
      </div>
    );

    if (fullScreen) {
      return (
        <div
          ref={ref}
          className="fixed inset-0 flex items-center justify-center bg-black/20 dark:bg-black/40 z-50"
          {...props}
        >
          <div className="bg-white dark:bg-neutral-800 rounded-lg p-8 flex flex-col items-center gap-4">
            {spinner}
            {withText && (
              <p className="text-sm text-neutral-600 dark:text-neutral-400">{withText}</p>
            )}
          </div>
        </div>
      );
    }

    if (withText) {
      return (
        <div className="flex flex-col items-center gap-2">
          {spinner}
          <p className="text-sm text-neutral-600 dark:text-neutral-400">{withText}</p>
        </div>
      );
    }

    return spinner;
  }
);

Spinner.displayName = 'Spinner';

