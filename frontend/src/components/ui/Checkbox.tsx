'use client';

import React, { forwardRef, InputHTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';
import { Check } from 'lucide-react';

interface CheckboxProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: ReactNode;
  helperText?: string;
  error?: boolean | string;
  containerClassName?: string;
  labelClassName?: string;
}

/**
 * Checkbox 组件
 * 带有标签和辅助文本支持
 */
export const Checkbox = forwardRef<HTMLInputElement, CheckboxProps>(
  (
    {
      className,
      label,
      helperText,
      error,
      containerClassName,
      labelClassName,
      disabled,
      ...props
    },
    ref
  ) => {
    const hasError = Boolean(error);
    const checkboxId = props.id || `checkbox-${Math.random()}`;

    return (
      <div className={cn('w-full', containerClassName)}>
        <div className="flex items-center gap-3">
          <div className="relative">
            <input
              ref={ref}
              id={checkboxId}
              type="checkbox"
              disabled={disabled}
              className={cn(
                'w-5 h-5 cursor-pointer appearance-none border-2 rounded transition-all duration-200',
                'focus:outline-none focus:ring-2 focus:ring-offset-2',
                'disabled:bg-neutral-50 disabled:cursor-not-allowed disabled:opacity-50',
                'dark:bg-neutral-800 dark:border-neutral-700',
                hasError
                  ? 'border-error focus:ring-error'
                  : 'border-neutral-300 focus:border-primary-500 focus:ring-primary-500 dark:border-neutral-600',
                'checked:bg-primary-500 checked:border-primary-500 dark:checked:bg-primary-600',
                className
              )}
              {...props}
            />
            {/* Checkmark */}
            <Check className="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 w-3 h-3 text-white pointer-events-none opacity-0 checked:opacity-100 transition-opacity" />
          </div>

          {label && (
            <label
              htmlFor={checkboxId}
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

        {(helperText || error) && (
          <p
            className={cn(
              'mt-1 text-sm ml-8',
              typeof error === 'string'
                ? 'text-error'
                : hasError
                  ? 'text-error'
                  : 'text-neutral-500'
            )}
          >
            {typeof error === 'string' ? error : helperText}
          </p>
        )}
      </div>
    );
  }
);

Checkbox.displayName = 'Checkbox';

