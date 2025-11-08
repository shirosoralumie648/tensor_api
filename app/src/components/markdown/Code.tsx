import { MarkdownFile } from "@/components/plugins/file.tsx";
import { MarkdownProgressbar } from "@/components/plugins/progress.tsx";
import { cn } from "@/components/ui/lib/utils.ts";
import { copyClipboard } from "@/utils/dom.ts";
import { Check, Copy } from "lucide-react";
import { LightAsync as SyntaxHighlighter } from "react-syntax-highlighter";
import { atomOneDark as style } from "react-syntax-highlighter/dist/esm/styles/hljs";
import React, { useMemo } from "react";
import { MarkdownMermaid } from "@/components/plugins/mermaid.tsx";
import { MarkdownChart } from "@/components/plugins/chart.tsx";
import { getBooleanMemory } from "@/utils/memory.ts";

import js from "react-syntax-highlighter/dist/esm/languages/hljs/javascript";
import ts from "react-syntax-highlighter/dist/esm/languages/hljs/typescript";
import py from "react-syntax-highlighter/dist/esm/languages/hljs/python";
import go from "react-syntax-highlighter/dist/esm/languages/hljs/go";
import json from "react-syntax-highlighter/dist/esm/languages/hljs/json";
import bash from "react-syntax-highlighter/dist/esm/languages/hljs/bash";
import shell from "react-syntax-highlighter/dist/esm/languages/hljs/shell";
import yaml from "react-syntax-highlighter/dist/esm/languages/hljs/yaml";
import md from "react-syntax-highlighter/dist/esm/languages/hljs/markdown";
import java from "react-syntax-highlighter/dist/esm/languages/hljs/java";
import cpp from "react-syntax-highlighter/dist/esm/languages/hljs/cpp";
import c from "react-syntax-highlighter/dist/esm/languages/hljs/c";
import rust from "react-syntax-highlighter/dist/esm/languages/hljs/rust";
import php from "react-syntax-highlighter/dist/esm/languages/hljs/php";
import ruby from "react-syntax-highlighter/dist/esm/languages/hljs/ruby";
import swift from "react-syntax-highlighter/dist/esm/languages/hljs/swift";
import kotlin from "react-syntax-highlighter/dist/esm/languages/hljs/kotlin";
import sql from "react-syntax-highlighter/dist/esm/languages/hljs/sql";

(SyntaxHighlighter as any).registerLanguage("javascript", js);
(SyntaxHighlighter as any).registerLanguage("typescript", ts);
(SyntaxHighlighter as any).registerLanguage("python", py);
(SyntaxHighlighter as any).registerLanguage("go", go);
(SyntaxHighlighter as any).registerLanguage("json", json);
(SyntaxHighlighter as any).registerLanguage("bash", bash);
(SyntaxHighlighter as any).registerLanguage("shell", shell);
(SyntaxHighlighter as any).registerLanguage("yaml", yaml);
(SyntaxHighlighter as any).registerLanguage("markdown", md);
(SyntaxHighlighter as any).registerLanguage("java", java);
(SyntaxHighlighter as any).registerLanguage("cpp", cpp);
(SyntaxHighlighter as any).registerLanguage("c", c);
(SyntaxHighlighter as any).registerLanguage("rust", rust);
(SyntaxHighlighter as any).registerLanguage("php", php);
(SyntaxHighlighter as any).registerLanguage("ruby", ruby);
(SyntaxHighlighter as any).registerLanguage("swift", swift);
(SyntaxHighlighter as any).registerLanguage("kotlin", kotlin);
(SyntaxHighlighter as any).registerLanguage("sql", sql);

const LanguageMap: Record<string, string> = {
  html: "htmlbars",
  js: "javascript",
  ts: "typescript",
  jsx: "javascript",
  tsx: "typescript",
  rs: "rust",
};

export type CodeProps = {
  inline?: boolean;
  className?: string;
  children: React.ReactNode;
  codeStyle?: string;
  loading?: boolean;
};

function Code({
  inline,
  className,
  children,
  loading,
  codeStyle,
  ...props
}: CodeProps) {
  const [copied, setCopied] = React.useState(false);
  const match = /language-(\w+)/.exec(className || "");
  const language = match ? match[1].toLowerCase() : "unknown";
  const mdMermaid = getBooleanMemory("feature_md_mermaid", true);
  const mdChart = getBooleanMemory("feature_md_chart", true);
  const mdHighlight = getBooleanMemory("feature_md_highlight", true);
  if (language === "file") return <MarkdownFile children={children} />;
  if (language === "progress")
    return <MarkdownProgressbar children={children} />;
  if (language === "mermaid")
    return mdMermaid ? (
      <MarkdownMermaid children={children} />
    ) : (
      <pre className={cn("code-block", codeStyle)}>{children?.toString()}</pre>
    );
  if (language === "chart")
    return mdChart ? (
      <MarkdownChart children={children} />
    ) : (
      <pre className={cn("code-block", codeStyle)}>{children?.toString()}</pre>
    );

  if (inline)
    return (
      <code className={cn("code-inline", className)} {...props}>
        {children}
      </code>
    );

  if (!mdHighlight) {
    return (
      <pre className={cn("code-block", codeStyle)}>
        {String(children).replace(/\n$/, "")}
      </pre>
    );
  }

  return (
    <div className={`markdown-syntax`}>
      <div
        className={`markdown-syntax-header`}
        onClick={async () => {
          const text = children?.toString() || "";
          await copyClipboard(text);
          setCopied(true);
        }}
      >
        {copied ? (
          <Check className={`h-3 w-3`} />
        ) : (
          <Copy className={`h-3 w-3`} />
        )}
        <p>{language}</p>
      </div>
      <SyntaxHighlighter
        {...props}
        children={String(children).replace(/\n$/, "")}
        style={style}
        language={LanguageMap[language] || language}
        PreTag="div"
        wrapLongLines={true}
        wrapLines={true}
        className={cn("code-block", codeStyle)}
      />
    </div>
  );
}

export default function ({
  inline,
  className,
  children,
  codeStyle,
  loading,
  ...props
}: CodeProps) {
  return useMemo(() => {
    return (
      <Code
        inline={inline}
        className={className}
        children={children}
        codeStyle={codeStyle}
        loading={loading}
        {...props}
      />
    );
  }, [inline, className, children, codeStyle, loading, props]);
}
