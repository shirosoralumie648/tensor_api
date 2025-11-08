import ModelChart from "@/components/admin/assemblies/ModelChart.tsx";
import { useState } from "react";
import {
  BillingChartResponse,
  ErrorChartResponse,
  ModelChartResponse,
  RequestChartResponse,
  UserTypeChartResponse,
} from "@/admin/types.ts";
import RequestChart from "@/components/admin/assemblies/RequestChart.tsx";
import BillingChart from "@/components/admin/assemblies/BillingChart.tsx";
import ErrorChart from "@/components/admin/assemblies/ErrorChart.tsx";
import { useEffectAsync } from "@/utils/hook.ts";
import {
  getBillingChart,
  getErrorChart,
  getModelChart,
  getRequestChart,
  getUserTypeChart,
  getRevenueByGateway,
  getRevenueByPlan,
} from "@/admin/api/chart.ts";
import ModelUsageChart from "@/components/admin/assemblies/ModelUsageChart.tsx";
import UserTypeChart from "@/components/admin/assemblies/UserTypeChart.tsx";
import RevenueGatewayChart from "@/components/admin/assemblies/RevenueGatewayChart.tsx";
import RevenuePlanChart from "@/components/admin/assemblies/RevenuePlanChart.tsx";

function ChartBox() {
  const [model, setModel] = useState<ModelChartResponse>({
    date: [],
    value: [],
  });

  const [request, setRequest] = useState<RequestChartResponse>({
    date: [],
    value: [],
  });

  const [billing, setBilling] = useState<BillingChartResponse>({
    date: [],
    value: [],
  });

  const [error, setError] = useState<ErrorChartResponse>({
    date: [],
    value: [],
  });

  const [user, setUser] = useState<UserTypeChartResponse>({
    total: 0,
    normal: 0,
    api_paid: 0,
    basic_plan: 0,
    standard_plan: 0,
    pro_plan: 0,
  });

  const [revenueGateway, setRevenueGateway] = useState<{ name: string; amount: number }[]>(
    [],
  );
  const [revenuePlan, setRevenuePlan] = useState<{ name: string; amount: number }[]>([]);

  useEffectAsync(async () => {
    setModel(await getModelChart());
    setRequest(await getRequestChart());
    setBilling(await getBillingChart());
    setError(await getErrorChart());
    setUser(await getUserTypeChart());
    setRevenueGateway((await getRevenueByGateway(30)).data || []);
    setRevenuePlan((await getRevenueByPlan(30)).data || []);
  }, []);

  return (
    <div className={`chart-boxes`}>
      <div className={`chart-box`}>
        <ModelChart labels={model.date} datasets={model.value} />
      </div>
      <div className={`chart-box`}>
        <ModelUsageChart labels={model.date} datasets={model.value} />
      </div>
      <div className={`chart-box`}>
        <BillingChart labels={billing.date} datasets={billing.value} />
      </div>
      <div className={`chart-box`}>
        <UserTypeChart data={user} />
      </div>
      <div className={`chart-box`}>
        <RequestChart labels={request.date} datasets={request.value} />
      </div>
      <div className={`chart-box`}>
        <ErrorChart labels={error.date} datasets={error.value} />
      </div>
      <div className={`chart-box`}>
        <RevenueGatewayChart data={revenueGateway} />
      </div>
      <div className={`chart-box`}>
        <RevenuePlanChart data={revenuePlan} />
      </div>
    </div>
  );
}

export default ChartBox;
