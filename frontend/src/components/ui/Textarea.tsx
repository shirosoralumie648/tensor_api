'use client';

import React, { forwardRef, TextareaHTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  helperText?: string;
  error?: boolean | string;
  leftIcon?: ReactNode;
  charCount?: boolean;
  maxChars?: number;
  containerClassName?: string;
  labelClassName?: string;
}

/**
 * Textarea 组件
 * 支持标签、字符计数和错误状态
 */
export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  (
    {
      className,
      label,
      helperText,
      error,
      leftIcon,
      charCount,
      maxChars,
      containerClassName,
      labelClassName,
      disabled,
      defaultValue,
      value,
      onChange,
      ...props
    },
    ref
  ) => {
    const hasError = Boolean(error);
    const [charCountValue, setCharCountValue] = React.useState(
      (value || defaultValue || '').toString().length
    );

    const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
      setCharCountValue(e.target.value.length);
      onChange?.(e);
    };

    return (
      <div className={cn('w-full', containerClassName)}>
        {label && (
          <label className={cn('block text-sm font-medium text-neutral-700 mb-2', labelClassName)}>
            {label}
            {props.required && <span className="text-error ml-1">*</span>}
          </label>
        )}

        <textarea
          ref={ref}
          disabled={disabled}
          value={value}
          defaultValue={defaultValue}
          onChange={handleChange}
          maxLength={maxChars}
          className={cn(
            'w-full px-4 py-2 border-2 rounded-lg transition-colors duration-200 resize-vertical',
            'focus:outline-none focus:ring-2 focus:ring-offset-0',
            'placeholder:text-neutral-400',
            'disabled:bg-neutral-50 disabled:text-neutral-500 disabled:cursor-not-allowed',
            'dark:bg-neutral-800 dark:border-neutral-700 dark:text-white dark:placeholder:text-neutral-500',
            leftIcon ? 'pl-10' : '',
            hasError
              ? 'border-error focus:ring-error focus:border-error'
              : 'border-neutral-300 focus:border-primary-500 focus:ring-primary-500 dark:border-neutral-600',
            'min-h-32',
            className
          )}
          {...props}
        />

        <div className="flex items-center justify-between mt-2">
          {(helperText || error) && (
            <p
              className={cn(
                'text-sm',
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

          {charCount && (
            <p className="text-xs text-neutral-500">
              {charCountValue}
              {maxChars && ` / ${maxChars}`}
            </p>
          )}
        </div>
      </div>
    );
  }
);

Textarea.displayName = 'Textarea';

