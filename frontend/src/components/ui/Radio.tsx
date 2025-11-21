'use client';

import React, { forwardRef, InputHTMLAttributes } from 'react';
import { cn } from '@/lib/utils';

interface RadioProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  helperText?: string;
  containerClassName?: string;
  labelClassName?: string;
}

/**
 * Radio 组件
 * 支持标签和辅助文本
 */
export const Radio = forwardRef<HTMLInputElement, RadioProps>(
  (
    {
      className,
      label,
      helperText,
      containerClassName,
      labelClassName,
      disabled,
      ...props
    },
    ref
  ) => {
    const radioId = props.id || `radio-${Math.random()}`;

    return (
      <div className={cn('w-full', containerClassName)}>
        <div className="flex items-center gap-3">
          <div className="relative">
            <input
              ref={ref}
              id={radioId}
              type="radio"
              disabled={disabled}
              className={cn(
                'w-5 h-5 cursor-pointer appearance-none border-2 rounded-full transition-all duration-200',
                'focus:outline-none focus:ring-2 focus:ring-offset-2',
                'disabled:bg-neutral-50 disabled:cursor-not-allowed disabled:opacity-50',
                'dark:bg-neutral-800 dark:border-neutral-700',
                'border-neutral-300 focus:border-primary-500 focus:ring-primary-500 dark:border-neutral-600',
                'checked:border-primary-500 dark:checked:border-primary-600',
                className
              )}
              {...props}
            />
            {/* Radio dot */}
            <div className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-2 h-2 bg-primary-500 rounded-full pointer-events-none opacity-0 checked:opacity-100 transition-opacity dark:bg-primary-600" />
          </div>

          {label && (
            <label
              htmlFor={radioId}
              className={cn(
                'text-sm font-medium cursor-pointer select-none',
                disabled ? 'text-neutral-400 cursor-not-allowed' : 'text-neutral-700',
                'dark:text-neutral-300',
                labelClassName
              )}
            >
              {label}
              {props.required && <span className="text-error ml-1">*</span>}
            </label>
          )}
        </div>

        {helperText && (
          <p className={cn('mt-1 text-sm text-neutral-500 ml-8')}>
            {helperText}
          </p>
        )}
      </div>
    );
  }
);

Radio.displayName = 'Radio';

