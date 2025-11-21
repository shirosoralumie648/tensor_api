'use client';

import React, { forwardRef, InputHTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  helperText?: string;
  error?: boolean | string;
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  containerClassName?: string;
  labelClassName?: string;
}

/**
 * Input 组件
 * 支持标签、辅助文本、图标和错误状态
 */
export const Input = forwardRef<HTMLInputElement, InputProps>(
  (
    {
      className,
      label,
      helperText,
      error,
      leftIcon,
      rightIcon,
      containerClassName,
      labelClassName,
      type = 'text',
      disabled,
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
          {leftIcon && (
            <div className="absolute left-0 top-0 h-full flex items-center justify-center px-3 text-neutral-500 pointer-events-none">
              {leftIcon}
            </div>
          )}

          <input
            ref={ref}
            type={type}
            disabled={disabled}
            className={cn(
              'w-full px-4 py-2 border-2 rounded-lg transition-colors duration-200',
              'focus:outline-none focus:ring-2 focus:ring-offset-0',
              'placeholder:text-neutral-400',
              'disabled:bg-neutral-50 disabled:text-neutral-500 disabled:cursor-not-allowed',
              'dark:bg-neutral-800 dark:border-neutral-700 dark:text-white dark:placeholder:text-neutral-500',
              leftIcon ? 'pl-10' : '',
              rightIcon ? 'pr-10' : '',
              hasError
                ? 'border-error focus:ring-error focus:border-error'
                : 'border-neutral-300 focus:border-primary-500 focus:ring-primary-500 dark:border-neutral-600',
              className
            )}
            {...props}
          />

          {rightIcon && (
            <div className="absolute right-0 top-0 h-full flex items-center justify-center px-3 text-neutral-500 pointer-events-none">
              {rightIcon}
            </div>
          )}
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

Input.displayName = 'Input';

