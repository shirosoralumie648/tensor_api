/**
 * Oblivious Design System Tokens
 * 基于苹果设计风格的完整 Token 系统
 */

export const tokens = {
  // ============= 颜色系统 =============
  colors: {
    // 主色系统 (iOS Blue)
    primary: {
      0: '#FFFFFF',
      50: '#F0F7FF',
      100: '#E0EFFE',
      200: '#BAE7FF',
      300: '#7AC5FF',
      400: '#36A3FF',
      500: '#0A84FF', // 主色
      600: '#0066E6',
      700: '#0052CC',
      800: '#003BA3',
      900: '#00235B',
    },

    // 功能色系
    success: '#34C759',
    warning: '#FF9500',
    error: '#FF3B30',
    info: '#5AC8FA',

    // 中性色系
    neutral: {
      0: '#FFFFFF',
      50: '#F9FAFB',
      100: '#F3F4F6',
      150: '#EEEFF2',
      200: '#E5E7EB',
      300: '#D1D5DB',
      400: '#9CA3AF',
      500: '#6B7280',
      600: '#4B5563',
      700: '#374151',
      800: '#1F2937',
      900: '#111827',
    },

    // 深色模式
    dark: {
      bg: '#1C1C1E',
      surface: '#2C2C2E',
      border: '#38383A',
      text: '#FFFFFF',
      textSecondary: '#8E8E93',
    },

    // 浅色模式
    light: {
      bg: '#FFFFFF',
      surface: '#F2F2F7',
      border: '#E5E7EB',
      text: '#000000',
      textSecondary: '#8E8E93',
    },
  },

  // ============= 排版系统 =============
  typography: {
    fontFamily: {
      base: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
      mono: '"SF Mono", Monaco, "Cascadia Code", "Roboto Mono", Consolas, "Courier New", monospace',
    },

    headings: {
      h1: {
        size: '2.5rem',    // 40px
        weight: 700,
        lineHeight: 1.2,
        letterSpacing: '-0.01em',
      },
      h2: {
        size: '2rem',      // 32px
        weight: 600,
        lineHeight: 1.3,
        letterSpacing: '-0.005em',
      },
      h3: {
        size: '1.5rem',    // 24px
        weight: 600,
        lineHeight: 1.4,
      },
      h4: {
        size: '1.25rem',   // 20px
        weight: 600,
        lineHeight: 1.4,
      },
      h5: {
        size: '1.125rem',  // 18px
        weight: 600,
        lineHeight: 1.5,
      },
      h6: {
        size: '1rem',      // 16px
        weight: 600,
        lineHeight: 1.5,
      },
    },

    body: {
      large: {
        size: '1rem',      // 16px
        weight: 400,
        lineHeight: 1.6,
      },
      normal: {
        size: '0.875rem',  // 14px
        weight: 400,
        lineHeight: 1.5,
      },
      small: {
        size: '0.75rem',   // 12px
        weight: 400,
        lineHeight: 1.4,
      },
      tiny: {
        size: '0.625rem',  // 10px
        weight: 400,
        lineHeight: 1.4,
      },
    },

    label: {
      large: {
        size: '0.875rem',
        weight: 500,
        lineHeight: 1.5,
      },
      normal: {
        size: '0.75rem',
        weight: 500,
        lineHeight: 1.4,
      },
      small: {
        size: '0.625rem',
        weight: 500,
        lineHeight: 1.4,
      },
    },
  },

  // ============= 间距系统 =============
  spacing: {
    0: '0',
    px: '1px',
    0.5: '0.125rem',  // 2px
    1: '0.25rem',     // 4px
    2: '0.5rem',      // 8px
    3: '0.75rem',     // 12px
    4: '1rem',        // 16px
    5: '1.25rem',     // 20px
    6: '1.5rem',      // 24px
    8: '2rem',        // 32px
    10: '2.5rem',     // 40px
    12: '3rem',       // 48px
    16: '4rem',       // 64px
    20: '5rem',       // 80px
    24: '6rem',       // 96px
  },

  // ============= 圆角系统 =============
  radius: {
    none: '0',
    xs: '0.25rem',    // 4px
    sm: '0.375rem',   // 6px
    md: '0.5rem',     // 8px
    lg: '0.75rem',    // 12px
    xl: '1rem',       // 16px
    '2xl': '1.25rem', // 20px
    '3xl': '1.5rem',  // 24px
    full: '9999px',
  },

  // ============= 阴影系统 =============
  shadows: {
    none: 'none',
    xs: '0 1px 2px 0 rgba(0, 0, 0, 0.05)',
    sm: '0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06)',
    md: '0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)',
    lg: '0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05)',
    xl: '0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04)',
    '2xl': '0 25px 50px -12px rgba(0, 0, 0, 0.25)',
    inner: 'inset 0 2px 4px 0 rgba(0, 0, 0, 0.05)',
  },

  // ============= 动画/过渡 =============
  transitions: {
    fast: {
      duration: '150ms',
      easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
    },
    normal: {
      duration: '300ms',
      easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
    },
    slow: {
      duration: '500ms',
      easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
    },
    elastic: {
      duration: '600ms',
      easing: 'cubic-bezier(0.175, 0.885, 0.32, 1.275)',
    },
  },

  // ============= 尺寸系统 =============
  sizing: {
    xs: '20rem',   // 320px
    sm: '24rem',   // 384px
    md: '28rem',   // 448px
    lg: '32rem',   // 512px
    xl: '36rem',   // 576px
    '2xl': '42rem', // 672px
    '3xl': '48rem', // 768px
    '4xl': '56rem', // 896px
    '5xl': '64rem', // 1024px
    '6xl': '72rem', // 1152px
    full: '100%',
    screen: '100vw',
  },

  // ============= 响应式断点 =============
  breakpoints: {
    xs: '0px',      // 移动端
    sm: '640px',    // 小屏幕
    md: '768px',    // 平板
    lg: '1024px',   // 桌面
    xl: '1280px',   // 大屏幕
    '2xl': '1536px', // 超大屏幕
  },

  // ============= Z-index 系统 =============
  zIndex: {
    hide: -1,
    auto: 'auto',
    base: 0,
    dropdown: 1000,
    sticky: 1020,
    fixed: 1030,
    backdrop: 1040,
    modal: 1050,
    popover: 1060,
    tooltip: 1070,
  },
};

/**
 * Tailwind CSS 配置导出
 */
export const tailwindConfig = {
  extend: {
    colors: {
      primary: tokens.colors.primary,
      success: tokens.colors.success,
      warning: tokens.colors.warning,
      error: tokens.colors.error,
      info: tokens.colors.info,
      neutral: tokens.colors.neutral,
    },
    spacing: tokens.spacing,
    borderRadius: tokens.radius,
    boxShadow: tokens.shadows,
    fontFamily: {
      base: tokens.typography.fontFamily.base,
      mono: tokens.typography.fontFamily.mono,
    },
    transitionDuration: {
      fast: tokens.transitions.fast.duration,
      normal: tokens.transitions.normal.duration,
      slow: tokens.transitions.slow.duration,
    },
    maxWidth: tokens.sizing,
    width: tokens.sizing,
  },
};

export default tokens;

