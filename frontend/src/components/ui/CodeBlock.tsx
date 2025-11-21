'use client';

import React, { forwardRef, HTMLAttributes } from 'react';
import { cn } from '@/lib/utils';
import { Copy, Check } from 'lucide-react';
import { useState } from 'react';

interface CodeBlockProps extends HTMLAttributes<HTMLDivElement> {
  code: string;
  language?: string;
  showLineNumbers?: boolean;
  copyable?: boolean;
}

/**
 * CodeBlock 组件
 * 用于显示和高亮代码
 */
export const CodeBlock = forwardRef<HTMLDivElement, CodeBlockProps>(
  (
    {
      code,
      language = 'javascript',
      showLineNumbers = false,
      copyable = true,
      className,
      ...props
    },
    ref
  ) => {
    const [copied, setCopied] = useState(false);

    const handleCopy = () => {
      navigator.clipboard.writeText(code);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    };

    const lines = code.split('\n');

    return (
      <div
        ref={ref}
        className={cn(
          'relative rounded-lg bg-neutral-900 text-neutral-100 text-sm overflow-hidden',
          className
        )}
        {...props}
      >
        {/* Header */}
        <div className="px-4 py-3 border-b border-neutral-800 flex items-center justify-between bg-neutral-950">
          <span className="text-xs font-medium text-neutral-500 uppercase tracking-wider">
            {language}
          </span>
          {copyable && (
            <button
              onClick={handleCopy}
              className="p-2 text-neutral-400 hover:text-neutral-200 transition rounded"
              title="Copy code"
            >
              {copied ? (
                <Check size={16} className="text-green-400" />
              ) : (
                <Copy size={16} />
              )}
            </button>
          )}
        </div>

        {/* Code */}
        <div className="overflow-x-auto">
          <pre className="p-4 font-mono">
            <code>
              {lines.map((line, idx) => (
                <div key={idx} className="whitespace-pre-wrap break-words">
                  {showLineNumbers && (
                    <span className="inline-block w-8 text-right pr-4 text-neutral-600 select-none mr-2">
                      {idx + 1}
                    </span>
                  )}
                  <span className="text-neutral-100">
                    {/* Simple syntax highlighting for common patterns */}
                    {highlightLine(line)}
                  </span>
                </div>
              ))}
            </code>
          </pre>
        </div>
      </div>
    );
  }
);

CodeBlock.displayName = 'CodeBlock';

/**
 * 简单的代码高亮
 */
function highlightLine(line: string): React.ReactNode {
  // 这是一个简化的高亮实现
  // 在实际应用中可以使用 Prism.js 或 Shiki

  // 替换字符串
  let highlighted = line.replace(
    /(['"])(?:(?=(\\?))\2.)*?\1/g,
    '<span class="text-green-400">$&</span>'
  );

  // 替换数字
  highlighted = highlighted.replace(
    /\b(\d+)\b/g,
    '<span class="text-orange-400">$1</span>'
  );

  // 替换布尔值
  highlighted = highlighted.replace(
    /\b(true|false|null|undefined)\b/g,
    '<span class="text-blue-400">$1</span>'
  );

  // 替换关键字
  highlighted = highlighted.replace(
    /\b(function|const|let|var|return|if|else|for|while|class|import|export|from|as)\b/g,
    '<span class="text-purple-400">$1</span>'
  );

  // 将 HTML 字符串转换为 React 元素
  return <span dangerouslySetInnerHTML={{ __html: highlighted }} />;
}

export default CodeBlock;

