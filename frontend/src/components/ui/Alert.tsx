'use client';

import React, { forwardRef, HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';
import { AlertCircle, CheckCircle, AlertTriangle, Info, X } from 'lucide-react';

type AlertVariant = 'success' | 'error' | 'warning' | 'info';

const variantConfig: Record<
  AlertVariant,
  {
    bgColor: string;
    borderColor: string;
    textColor: string;
    icon: ReactNode;
  }
> = {
  success: {
    bgColor: 'bg-emerald-50 dark:bg-emerald-950',
    borderColor: 'border-emerald-200 dark:border-emerald-800',
    textColor: 'text-emerald-800 dark:text-emerald-200',
    icon: <CheckCircle className="w-5 h-5 text-emerald-600 dark:text-emerald-400" />,
  },
  error: {
    bgColor: 'bg-red-50 dark:bg-red-950',
    borderColor: 'border-red-200 dark:border-red-800',
    textColor: 'text-red-800 dark:text-red-200',
    icon: <AlertCircle className="w-5 h-5 text-red-600 dark:text-red-400" />,
  },
  warning: {
    bgColor: 'bg-amber-50 dark:bg-amber-950',
    borderColor: 'border-amber-200 dark:border-amber-800',
    textColor: 'text-amber-800 dark:text-amber-200',
    icon: <AlertTriangle className="w-5 h-5 text-amber-600 dark:text-amber-400" />,
  },
  info: {
    bgColor: 'bg-blue-50 dark:bg-blue-950',
    borderColor: 'border-blue-200 dark:border-blue-800',
    textColor: 'text-blue-800 dark:text-blue-200',
    icon: <Info className="w-5 h-5 text-blue-600 dark:text-blue-400" />,
  },
};

interface AlertProps extends HTMLAttributes<HTMLDivElement> {
  variant?: AlertVariant;
  title?: string;
  description?: string;
  icon?: ReactNode;
  closable?: boolean;
  onClose?: () => void;
  children?: ReactNode;
}

/**
 * Alert 组件
 * 用于显示警告、错误、成功和信息消息
 */
export const Alert = forwardRef<HTMLDivElement, AlertProps>(
  (
    {
      className,
      variant = 'info',
      title,
      description,
      icon,
      closable,
      onClose,
      children,
      ...props
    },
    ref
  ) => {
    const [isOpen, setIsOpen] = React.useState(true);

    if (!isOpen) return null;

    const config = variantConfig[variant];

    const handleClose = () => {
      setIsOpen(false);
      onClose?.();
    };

    return (
      <div
        ref={ref}
        className={cn(
          'flex gap-3 p-4 rounded-lg border-2',
          config.bgColor,
          config.borderColor,
          className
        )}
        role="alert"
        {...props}
      >
        {/* Icon */}
        <div className="flex-shrink-0 mt-0.5">
          {icon || config.icon}
        </div>

        {/* Content */}
        <div className="flex-1">
          {title && <h3 className={cn('font-semibold', config.textColor)}>{title}</h3>}
          {description && (
            <p className={cn('text-sm', config.textColor)}>
              {description}
            </p>
          )}
          {children && <div className={cn('text-sm', config.textColor)}>{children}</div>}
        </div>

        {/* Close button */}
        {closable && (
          <button
            onClick={handleClose}
            className={cn(
              'flex-shrink-0 transition-opacity hover:opacity-75',
              config.textColor
            )}
            aria-label="Close alert"
          >
            <X className="w-5 h-5" />
          </button>
        )}
      </div>
    );
  }
);

Alert.displayName = 'Alert';

