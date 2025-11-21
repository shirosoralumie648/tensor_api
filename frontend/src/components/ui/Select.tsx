'use client';

import React, { forwardRef, SelectHTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';
import { ChevronDown } from 'lucide-react';

interface SelectOption {
  value: string | number;
  label: string;
  disabled?: boolean;
}

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  helperText?: string;
  error?: boolean | string;
  options?: SelectOption[];
  placeholder?: string;
  containerClassName?: string;
  labelClassName?: string;
  children?: ReactNode;
}

/**
 * Select 组件
 * 支持标签、辅助文本和错误状态
 */
export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  (
    {
      className,
      label,
      helperText,
      error,
      options,
      placeholder,
      containerClassName,
      labelClassName,
      disabled,
      children,
      ...props
    },
    ref
  ) => {
    const hasError = Boolean(error);

    return (
      <div className={cn('w-full', containerClassName)}>
        {label && (
          <label className={cn('block text-sm font-medium text-neutral-700 mb-2', labelClassName)}>
            {label}
            {props.required && <span className="text-error ml-1">*</span>}
          </label>
        )}

        <div className="relative">
          <select
            ref={ref}
            disabled={disabled}
            className={cn(
              'w-full px-4 py-2 pr-10 border-2 rounded-lg transition-colors duration-200 appearance-none',
              'focus:outline-none focus:ring-2 focus:ring-offset-0',
              'placeholder:text-neutral-400',
              'disabled:bg-neutral-50 disabled:text-neutral-500 disabled:cursor-not-allowed',
              'dark:bg-neutral-800 dark:border-neutral-700 dark:text-white dark:placeholder:text-neutral-500',
              hasError
                ? 'border-error focus:ring-error focus:border-error'
                : 'border-neutral-300 focus:border-primary-500 focus:ring-primary-500 dark:border-neutral-600',
              className
            )}
            {...props}
          >
            {placeholder && (
              <option value="" disabled>
                {placeholder}
              </option>
            )}
            {options?.map((option) => (
              <option key={option.value} value={option.value} disabled={option.disabled}>
                {option.label}
              </option>
            ))}
            {children}
          </select>

          <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-5 h-5 text-neutral-500 pointer-events-none" />
        </div>

        {(helperText || error) && (
          <p
            className={cn(
              'mt-1 text-sm',
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

Select.displayName = 'Select';

