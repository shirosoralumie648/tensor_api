import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Loader2 } from "lucide-react";
import { DonutChart, Legend } from "@tremor/react";
import type { RevenueGroupItem } from "@/admin/types.ts";

type Props = { data: RevenueGroupItem[] };

function RevenueGatewayChart({ data }: Props) {
  const { t } = useTranslation();

  const chart = useMemo(
    () =>
      (data || []).map((it) => ({ name: it.name, value: it.amount })),
    [data],
  );

  return (
    <div className={`chart`}>
      <div className={`chart-title mb-2`}>
        <div className={`flex items-center w-full`}>
          <div>{t("admin.revenue-gateway") || "按网关收入分布"}</div>
          {chart.length === 0 && (
            <Loader2 className={`h-4 w-4 ml-1 animate-spin`} />
          )}
        </div>
      </div>
      <div className={`flex flex-row`}>
        <DonutChart
          className={`common-chart p-4 w-[50%]`}
          variant={`donut`}
          data={chart}
          showAnimation={true}
          colors={["orange", "emerald", "cyan", "violet", "rose", "amber"]}
        />
        <Legend
          className={`common-chart p-4 w-[50%]`}
          categories={chart.map((i) => i.name)}
          colors={["orange", "emerald", "cyan", "violet", "rose", "amber"]}
        />
      </div>
    </div>
  );
}

export default RevenueGatewayChart;
