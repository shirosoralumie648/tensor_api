'use client';

import React, { forwardRef, HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';

type GridColumns = 1 | 2 | 3 | 4 | 5 | 6 | 12 | 'auto';

const gridColClasses: Record<GridColumns, string> = {
  1: 'grid-cols-1',
  2: 'grid-cols-2',
  3: 'grid-cols-3',
  4: 'grid-cols-4',
  5: 'grid-cols-5',
  6: 'grid-cols-6',
  12: 'grid-cols-12',
  auto: 'grid-cols-auto',
};

interface GridProps extends HTMLAttributes<HTMLDivElement> {
  columns?: GridColumns | Partial<Record<'xs' | 'sm' | 'md' | 'lg' | 'xl', GridColumns>>;
  gap?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  autoFlow?: 'row' | 'column' | 'dense';
  children?: ReactNode;
}

const gapClasses = {
  xs: 'gap-1',
  sm: 'gap-2',
  md: 'gap-4',
  lg: 'gap-6',
  xl: 'gap-8',
};

const autoFlowClasses = {
  row: 'auto-flow-row',
  column: 'auto-flow-col',
  dense: 'auto-flow-dense',
};

/**
 * Grid 组件
 * 响应式网格布局组件
 */
export const Grid = forwardRef<HTMLDivElement, GridProps>(
  (
    {
      className,
      columns = 3,
      gap = 'md',
      autoFlow = 'row',
      children,
      ...props
    },
    ref
  ) => {
    const getColumnClasses = () => {
      if (typeof columns === 'number') {
        return gridColClasses[columns];
      }

      const breakpoints = ['xs', 'sm', 'md', 'lg', 'xl'] as const;
      let classes = '';

      for (const breakpoint of breakpoints) {
        const cols = (columns as any)[breakpoint];
        if (cols) {
          const prefix = breakpoint === 'xs' ? '' : `${breakpoint}:`;
          classes += ` ${prefix}${gridColClasses[cols as GridColumns]}`;
        }
      }

      return classes;
    };

    return (
      <div
        ref={ref}
        className={cn(
          'grid',
          getColumnClasses(),
          gapClasses[gap],
          autoFlowClasses[autoFlow],
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);

Grid.displayName = 'Grid';

