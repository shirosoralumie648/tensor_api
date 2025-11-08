import React, { useMemo } from "react";
import { AreaChart, BarChart, LineChart } from "@tremor/react";
import { cn } from "@/components/ui/lib/utils.ts";
import { Check, Copy } from "lucide-react";
import { copyClipboard } from "@/utils/dom.ts";

type ChartSpec = {
  type?: string;
  data?: any[];
  index?: string;
  categories?: string[];
  colors?: string[];
  valueFormatter?: (n: number) => string;
  className?: string;
  height?: number;
};

function renderChart(spec: ChartSpec) {
  const type = (spec.type || "line").toLowerCase();
  const height = spec.height || 320;
  const commonProps: any = {
    data: spec.data || [],
    index: spec.index || "index",
    categories: spec.categories || [],
    colors: spec.colors,
    valueFormatter: spec.valueFormatter,
    className: cn("w-full", "h-full"),
  };

  if (type === "bar")
    return (
      <div style={{ height: `${height}px` }}>
        <BarChart {...commonProps} />
      </div>
    );
  if (type === "area")
    return (
      <div style={{ height: `${height}px` }}>
        <AreaChart {...commonProps} />
      </div>
    );
  return (
    <div style={{ height: `${height}px` }}>
      <LineChart {...commonProps} />
    </div>
  );
}

export function MarkdownChart({ children }: { children: React.ReactNode }) {
  const [copied, setCopied] = React.useState(false);
  const content = children?.toString() || "";

  const spec = useMemo<ChartSpec | null>(() => {
    try {
      return JSON.parse(content);
    } catch (e) {
      return null;
    }
  }, [children]);

  return (
    <div className={`markdown-syntax`}>
      <div
        className={`markdown-syntax-header`}
        onClick={async () => {
          await copyClipboard(content);
          setCopied(true);
        }}
      >
        {copied ? <Check className={`h-3 w-3`} /> : <Copy className={`h-3 w-3`} />}
        <p>chart</p>
      </div>
      {spec ? (
        <div className={`mt-2`}>{renderChart(spec)}</div>
      ) : (
        <pre className={`code-block`}>{content}</pre>
      )}
    </div>
  );
}
