'use client';

import React, { forwardRef, HTMLAttributes } from 'react';
import { cn } from '@/lib/utils';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface PaginationProps extends HTMLAttributes<HTMLDivElement> {
  currentPage: number;
  totalPages: number;
  onPageChange: (page: number) => void;
  siblingCount?: number;
}

/**
 * Pagination 分页组件
 * 用于处理分页导航
 */
export const Pagination = forwardRef<HTMLDivElement, PaginationProps>(
  (
    {
      currentPage,
      totalPages,
      onPageChange,
      siblingCount = 1,
      className,
      ...props
    },
    ref
  ) => {
    // 计算显示的页码
    const getPageNumbers = () => {
      const pages: (number | string)[] = [];

      // 总是显示第一页
      pages.push(1);

      // 计算左省略号的起点
      const leftSiblingIndex = Math.max(currentPage - siblingCount, 2);
      const shouldShowLeftDots = leftSiblingIndex > 2;

      // 计算右省略号的起点
      const rightSiblingIndex = Math.min(currentPage + siblingCount, totalPages - 1);
      const shouldShowRightDots = rightSiblingIndex < totalPages - 1;

      if (shouldShowLeftDots) {
        pages.push('...');
      }

      // 添加中间页码
      for (let i = leftSiblingIndex; i <= rightSiblingIndex; i++) {
        pages.push(i);
      }

      if (shouldShowRightDots) {
        pages.push('...');
      }

      // 总是显示最后一页 (如果总页数 > 1)
      if (totalPages > 1 && !pages.includes(totalPages)) {
        pages.push(totalPages);
      }

      return pages;
    };

    const pages = getPageNumbers();

    return (
      <div
        ref={ref}
        className={cn(
          'flex items-center justify-center gap-2',
          className
        )}
        {...props}
      >
        {/* Previous Button */}
        <button
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 1}
          className={cn(
            'p-2 rounded-lg border transition',
            currentPage === 1
              ? 'border-neutral-200 dark:border-neutral-700 text-neutral-400 dark:text-neutral-600 cursor-not-allowed'
              : 'border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800'
          )}
          title="Previous page"
        >
          <ChevronLeft size={18} />
        </button>

        {/* Page Numbers */}
        <div className="flex items-center gap-1">
          {pages.map((page, idx) => (
            <button
              key={idx}
              onClick={() => typeof page === 'number' && onPageChange(page)}
              disabled={page === '...' || page === currentPage}
              className={cn(
                'w-9 h-9 rounded-lg font-medium text-sm transition',
                page === '...'
                  ? 'text-neutral-500 cursor-default'
                  : page === currentPage
                  ? 'bg-primary-500 text-white'
                  : 'border border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800'
              )}
            >
              {page}
            </button>
          ))}
        </div>

        {/* Next Button */}
        <button
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage === totalPages}
          className={cn(
            'p-2 rounded-lg border transition',
            currentPage === totalPages
              ? 'border-neutral-200 dark:border-neutral-700 text-neutral-400 dark:text-neutral-600 cursor-not-allowed'
              : 'border-neutral-300 dark:border-neutral-600 text-neutral-700 dark:text-neutral-300 hover:bg-neutral-100 dark:hover:bg-neutral-800'
          )}
          title="Next page"
        >
          <ChevronRight size={18} />
        </button>
      </div>
    );
  }
);

Pagination.displayName = 'Pagination';

export default Pagination;

