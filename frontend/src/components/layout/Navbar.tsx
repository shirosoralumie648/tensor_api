'use client';

import React, { forwardRef, HTMLAttributes, ReactNode } from 'react';
import { cn } from '@/lib/utils';

interface NavbarProps extends HTMLAttributes<HTMLElement> {
  sticky?: boolean;
  bordered?: boolean;
  shadow?: boolean;
  children?: ReactNode;
}

/**
 * Navbar 组件
 * 导航栏容器
 */
export const Navbar = forwardRef<HTMLElement, NavbarProps>(
  (
    {
      className,
      sticky = true,
      bordered = true,
      shadow = true,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <nav
        ref={ref}
        className={cn(
          'w-full bg-white dark:bg-neutral-800 z-40',
          sticky && 'sticky top-0',
          bordered && 'border-b border-neutral-200 dark:border-neutral-700',
          shadow && 'shadow-sm',
          className
        )}
        {...props}
      >
        {children}
      </nav>
    );
  }
);

Navbar.displayName = 'Navbar';

interface NavbarContentProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
}

/**
 * NavbarContent 组件
 * 导航栏内容容器
 */
export const NavbarContent = forwardRef<HTMLDivElement, NavbarContentProps>(
  (
    {
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          'max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between',
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);

NavbarContent.displayName = 'NavbarContent';

interface NavbarBrandProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
}

/**
 * NavbarBrand 组件
 * 导航栏品牌/Logo 区域
 */
export const NavbarBrand = forwardRef<HTMLDivElement, NavbarBrandProps>(
  (
    {
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          'flex items-center gap-2 cursor-pointer',
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);

NavbarBrand.displayName = 'NavbarBrand';

interface NavbarMenuProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
}

/**
 * NavbarMenu 组件
 * 导航菜单容器
 */
export const NavbarMenu = forwardRef<HTMLDivElement, NavbarMenuProps>(
  (
    {
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          'hidden md:flex items-center gap-8',
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);

NavbarMenu.displayName = 'NavbarMenu';

interface NavbarItemProps extends HTMLAttributes<HTMLAnchorElement> {
  href?: string;
  active?: boolean;
  children?: ReactNode;
}

/**
 * NavbarItem 组件
 * 导航菜单项
 */
export const NavbarItem = forwardRef<HTMLAnchorElement, NavbarItemProps>(
  (
    {
      className,
      href = '#',
      active,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <a
        ref={ref}
        href={href}
        className={cn(
          'text-sm font-medium transition-colors hover:text-primary-500',
          active
            ? 'text-primary-500 border-b-2 border-primary-500'
            : 'text-neutral-700 dark:text-neutral-300',
          className
        )}
        {...props}
      >
        {children}
      </a>
    );
  }
);

NavbarItem.displayName = 'NavbarItem';

interface NavbarActionsProps extends HTMLAttributes<HTMLDivElement> {
  children?: ReactNode;
}

/**
 * NavbarActions 组件
 * 导航栏操作按钮区域
 */
export const NavbarActions = forwardRef<HTMLDivElement, NavbarActionsProps>(
  (
    {
      className,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div
        ref={ref}
        className={cn(
          'flex items-center gap-4',
          className
        )}
        {...props}
      >
        {children}
      </div>
    );
  }
);

NavbarActions.displayName = 'NavbarActions';

