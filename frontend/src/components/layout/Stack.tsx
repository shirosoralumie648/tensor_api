'use client';

import React, { forwardRef, HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';

type StackDirection = 'row' | 'column';
type StackAlignment = 'start' | 'center' | 'end' | 'stretch' | 'baseline';
type StackJustification = 'start' | 'center' | 'end' | 'between' | 'around' | 'evenly';

const directionClasses: Record<StackDirection, string> = {
  row: 'flex-row',
  column: 'flex-col',
};

const alignClasses: Record<StackAlignment, string> = {
  start: 'items-start',
  center: 'items-center',
  end: 'items-end',
  stretch: 'items-stretch',
  baseline: 'items-baseline',
};

const justifyClasses: Record<StackJustification, string> = {
  start: 'justify-start',
  center: 'justify-center',
  end: 'justify-end',
  between: 'justify-between',
  around: 'justify-around',
  evenly: 'justify-evenly',
};

interface StackProps extends HTMLAttributes<HTMLDivElement> {
  direction?: StackDirection;
  align?: StackAlignment;
  justify?: StackJustification;
  spacing?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  fullWidth?: boolean;
  fullHeight?: boolean;
  wrap?: boolean;
  children?: ReactNode;
}

const spacingClasses: Record<'xs' | 'sm' | 'md' | 'lg' | 'xl', string> = {
  xs: 'gap-1',
  sm: 'gap-2',
  md: 'gap-4',
  lg: 'gap-6',
  xl: 'gap-8',
};

/**
 * Stack 组件
 * 灵活的堆栈布局组件
 */
export const Stack = forwardRef<HTMLDivElement, StackProps>(
  (
    {
      className,
      direction = 'column',
      align = 'start',
      justify = 'start',
      spacing = 'md',
      fullWidth,
      fullHeight,
      wrap,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          'flex',
          directionClasses[direction],
          alignClasses[align],
          justifyClasses[justify],
          spacingClasses[spacing],
          fullWidth && 'w-full',
          fullHeight && 'h-full',
          wrap && 'flex-wrap',
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);

Stack.displayName = 'Stack';

/**
 * VStack 组件 (Vertical Stack)
 * 竖直方向堆栈 (Column)
 */
export const VStack = forwardRef<HTMLDivElement, Omit<StackProps, 'direction'>>(
  (props, ref) => <Stack ref={ref} direction="column" {...props} />
);

VStack.displayName = 'VStack';

/**
 * HStack 组件 (Horizontal Stack)
 * 水平方向堆栈 (Row)
 */
export const HStack = forwardRef<HTMLDivElement, Omit<StackProps, 'direction'>>(
  (props, ref) => <Stack ref={ref} direction="row" {...props} />
);

HStack.displayName = 'HStack';

