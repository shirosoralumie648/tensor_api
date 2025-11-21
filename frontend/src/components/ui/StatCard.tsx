'use client';

import React, { forwardRef, HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface StatCardProps extends HTMLAttributes<HTMLDivElement> {
  label: string;
  value: string | number;
  icon?: ReactNode;
  trend?: {
    value: number;
    direction: 'up' | 'down';
  };
  color?: 'primary' | 'success' | 'warning' | 'error' | 'info';
}

const colorClasses = {
  primary: 'text-primary-500 bg-primary-50 dark:bg-primary-900/30',
  success: 'text-green-500 bg-green-50 dark:bg-green-900/30',
  warning: 'text-yellow-500 bg-yellow-50 dark:bg-yellow-900/30',
  error: 'text-red-500 bg-red-50 dark:bg-red-900/30',
  info: 'text-blue-500 bg-blue-50 dark:bg-blue-900/30',
};

/**
 * StatCard 统计卡片组件
 * 用于展示数据统计和指标
 */
export const StatCard = forwardRef<HTMLDivElement, StatCardProps>(
  (
    {
      label,
      value,
      icon,
      trend,
      color = 'primary',
      className,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          'p-6 rounded-lg border border-neutral-200 dark:border-neutral-700 bg-white dark:bg-neutral-800',
          className
        )}
        {...props}
      >
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <p className="text-sm font-medium text-neutral-600 dark:text-neutral-400 mb-2">
              {label}
            </p>
            <div className="flex items-baseline gap-2">
              <p className="text-3xl font-bold text-neutral-900 dark:text-white">
                {value}
              </p>
              {trend && (
                <div
                  className={cn(
                    'text-xs font-semibold px-2 py-1 rounded',
                    trend.direction === 'up'
                      ? 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-900/30'
                      : 'text-red-600 bg-red-100 dark:text-red-400 dark:bg-red-900/30'
                  )}
                >
                  {trend.direction === 'up' ? '↑' : '↓'} {Math.abs(trend.value)}%
                </div>
              )}
            </div>
          </div>
          {icon && (
            <div className={cn('p-3 rounded-lg', colorClasses[color])}>
              {icon}
            </div>
          )}
        </div>
      </div>
    );
  }
);

StatCard.displayName = 'StatCard';

export default StatCard;

