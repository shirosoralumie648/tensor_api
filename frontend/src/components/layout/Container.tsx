'use client';

import React, { forwardRef, HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';

type ContainerSize = 'sm' | 'md' | 'lg' | 'xl' | '2xl' | 'full';

const sizeClasses: Record<ContainerSize, string> = {
  sm: 'max-w-sm',
  md: 'max-w-md',
  lg: 'max-w-lg',
  xl: 'max-w-xl',
  '2xl': 'max-w-2xl',
  full: 'w-full',
};

interface ContainerProps extends HTMLAttributes<HTMLDivElement> {
  size?: ContainerSize;
  centered?: boolean;
  withPadding?: boolean;
  children?: ReactNode;
}

/**
 * Container 组件
 * 用于限制内容宽度和居中
 */
export const Container = forwardRef<HTMLDivElement, ContainerProps>(
  (
    {
      className,
      size = 'xl',
      centered = true,
      withPadding = true,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          'mx-auto',
          sizeClasses[size],
          centered && 'mx-auto',
          withPadding && 'px-4 sm:px-6 lg:px-8',
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);

Container.displayName = 'Container';

