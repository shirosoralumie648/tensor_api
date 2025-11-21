'use client';

import React, { forwardRef, HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface Column<T> {
  key: keyof T;
  label: string;
  render?: (value: any, row: T) => ReactNode;
  width?: string;
  align?: 'left' | 'center' | 'right';
}

interface TableProps<T> extends HTMLAttributes<HTMLTableElement> {
  columns: Column<T>[];
  data: T[];
  striped?: boolean;
  hoverable?: boolean;
  compact?: boolean;
}

/**
 * Table 表格组件
 * 用于展示结构化数据
 */
export const Table = forwardRef<
  HTMLTableElement,
  TableProps<any>
>(
  (
    {
      columns,
      data,
      striped = true,
      hoverable = true,
      compact = false,
      className,
      ...props
    },
    ref
  ) => {
    const alignClass = (align?: string) => {
      switch (align) {
        case 'center':
          return 'text-center';
        case 'right':
          return 'text-right';
        default:
          return 'text-left';
      }
    };

    return (
      <div className="overflow-x-auto rounded-lg border border-neutral-200 dark:border-neutral-700">
        <table
          ref={ref}
          className={cn(
            'w-full border-collapse',
            className
          )}
          {...props}
        >
          {/* Header */}
          <thead className="bg-neutral-100 dark:bg-neutral-800 border-b border-neutral-200 dark:border-neutral-700">
            <tr>
              {columns.map(column => (
                <th
                  key={String(column.key)}
                  className={cn(
                    'font-semibold text-neutral-900 dark:text-white',
                    'border-b border-neutral-200 dark:border-neutral-700',
                    alignClass(column.align),
                    compact ? 'px-3 py-2 text-xs' : 'px-4 py-3 text-sm'
                  )}
                  style={{ width: column.width }}
                >
                  {column.label}
                </th>
              ))}
            </tr>
          </thead>

          {/* Body */}
          <tbody>
            {data.length === 0 ? (
              <tr>
                <td
                  colSpan={columns.length}
                  className="px-4 py-8 text-center text-neutral-500 dark:text-neutral-400"
                >
                  No data available
                </td>
              </tr>
            ) : (
              data.map((row, rowIdx) => (
                <tr
                  key={rowIdx}
                  className={cn(
                    'border-b border-neutral-200 dark:border-neutral-700 last:border-b-0',
                    striped && rowIdx % 2 === 1
                      ? 'bg-neutral-50 dark:bg-neutral-900/50'
                      : 'bg-white dark:bg-neutral-900',
                    hoverable &&
                    'hover:bg-neutral-100 dark:hover:bg-neutral-800 transition'
                  )}
                >
                  {columns.map(column => (
                    <td
                      key={String(column.key)}
                      className={cn(
                        'text-neutral-900 dark:text-neutral-100',
                        alignClass(column.align),
                        compact ? 'px-3 py-2 text-xs' : 'px-4 py-3 text-sm'
                      )}
                    >
                      {column.render
                        ? column.render(row[column.key], row)
                        : String(row[column.key] ?? '-')}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    );
  }
);

Table.displayName = 'Table';

export default Table;

