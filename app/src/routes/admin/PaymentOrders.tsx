import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Label } from "@/components/ui/label.tsx";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card.tsx";
import {
  PaymentOrder,
  getPaymentOrders,
  syncPaymentOrder,
  refundPaymentOrder,
} from "@/admin/api/payment.ts";
import { useToast } from "@/components/ui/use-toast.ts";
import { toastState } from "@/api/common.ts";

export default function PaymentOrders() {
  const { toast } = useToast();
  const [list, setList] = useState<PaymentOrder[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(10);
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("");
  const [gateway, setGateway] = useState("");
  const [userId, setUserId] = useState<string>("");
  const [start, setStart] = useState("");
  const [end, setEnd] = useState("");

  const canPrev = useMemo(() => page > 1, [page]);
  const canNext = useMemo(() => page * size < total, [page, size, total]);

  const load = async () => {
    setLoading(true);
    const res = await getPaymentOrders({
      status: status || undefined,
      gateway: gateway || undefined,
      q: q || undefined,
      user_id: userId ? Number(userId) : undefined,
      start: start || undefined,
      end: end || undefined,
      page,
      size,
    });
    setLoading(false);
    if (!res.status) return toastState(toast, undefined, res);
    setList(res.data?.list || []);
    setTotal(res.data?.total || 0);
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, size]);

  const onSearch = () => {
    setPage(1);
    load();
  };

  const onSync = async (orderNo: string) => {
    const res = await syncPaymentOrder(orderNo);
    toastState(toast, undefined, res, true);
    if (res.status) load();
  };

  const onRefund = async (orderNo: string) => {
    const res = await refundPaymentOrder(orderNo);
    toastState(toast, undefined, res, true);
    if (res.status) load();
  };

  return (
    <div className={`system`}>
      <Card className={`admin-card`}>
        <CardHeader>
          <CardTitle>订单管理</CardTitle>
          <CardDescription>搜索、同步状态与退款</CardDescription>
        </CardHeader>
        <CardContent>
          <div className={`grid grid-cols-1 lg:grid-cols-3 gap-3`}>
            <div>
              <Label>关键词（订单号/三方单号/主题）</Label>
              <Input
                placeholder={`支持模糊搜索`}
                value={q}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setQ(e.target.value)}
              />
            </div>
            <div>
              <Label>状态</Label>
              <select
                className={`w-full h-9 rounded border px-2 text-sm bg-background`}
                value={status}
                onChange={(e) => setStatus(e.target.value)}
              >
                <option value="">全部</option>
                <option value="pending">pending</option>
                <option value="paid">paid</option>
                <option value="success">success</option>
                <option value="failed">failed</option>
                <option value="refunding">refunding</option>
                <option value="refunded">refunded</option>
                <option value="closed">closed</option>
              </select>
            </div>
            <div>
              <Label>网关</Label>
              <select
                className={`w-full h-9 rounded border px-2 text-sm bg-background`}
                value={gateway}
                onChange={(e) => setGateway(e.target.value)}
              >
                <option value="">全部</option>
                <option value="stripe">stripe</option>
                <option value="wechat">wechat</option>
                <option value="alipay">alipay</option>
                <option value="yipay">yipay</option>
                <option value="aggregate">aggregate</option>
              </select>
            </div>
            <div>
              <Label>用户ID</Label>
              <Input
                value={userId}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setUserId(e.target.value)}
              />
            </div>
            <div>
              <Label>开始时间</Label>
              <Input
                type={`datetime-local`}
                value={start}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setStart(e.target.value)}
              />
            </div>
            <div>
              <Label>结束时间</Label>
              <Input
                type={`datetime-local`}
                value={end}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEnd(e.target.value)}
              />
            </div>
          </div>
          <div className={`mt-3`}>
            <Button onClick={onSearch} loading={loading}>
              搜索
            </Button>
          </div>

          <div className={`mt-5 overflow-auto`}>
            <table className={`w-full text-sm`}>
              <thead>
                <tr className={`text-left border-b`}> 
                  <th className={`py-2 pr-3`}>订单号</th>
                  <th className={`py-2 pr-3`}>网关</th>
                  <th className={`py-2 pr-3`}>金额</th>
                  <th className={`py-2 pr-3`}>币种</th>
                  <th className={`py-2 pr-3`}>状态</th>
                  <th className={`py-2 pr-3`}>创建时间</th>
                  <th className={`py-2 pr-3`}>支付时间</th>
                  <th className={`py-2 pr-3`}>操作</th>
                </tr>
              </thead>
              <tbody>
                {list.length === 0 && (
                  <tr>
                    <td className={`py-4 text-muted-foreground`} colSpan={8}>
                      暂无数据
                    </td>
                  </tr>
                )}
                {list.map((o) => (
                  <tr key={o.order_no} className={`border-b hover:bg-muted/30`}>
                    <td className={`py-2 pr-3`}>{o.order_no}</td>
                    <td className={`py-2 pr-3 uppercase`}>{o.gateway}</td>
                    <td className={`py-2 pr-3`}>{o.amount}</td>
                    <td className={`py-2 pr-3`}>{o.currency}</td>
                    <td className={`py-2 pr-3`}>{o.status}</td>
                    <td className={`py-2 pr-3`}>{o.created_at}</td>
                    <td className={`py-2 pr-3`}>{o.paid_at || "-"}</td>
                    <td className={`py-2 pr-3`}> 
                      <div className={`flex gap-2`}>
                        <Button size={`sm`} variant={`outline`} onClick={() => onSync(o.order_no)}>
                          同步
                        </Button>
                        {(o.status === "paid" || o.status === "success") && (
                          <Button size={`sm`} variant={`destructive`} onClick={() => onRefund(o.order_no)}>
                            退款
                          </Button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className={`mt-4 flex items-center justify-between`}>
            <div className={`text-sm text-muted-foreground`}>
              共 {total} 条
            </div>
            <div className={`flex items-center gap-2`}>
              <Button disabled={!canPrev} variant={`outline`} onClick={() => setPage((p) => Math.max(1, p - 1))}>
                上一页
              </Button>
              <div className={`text-sm`}>第 {page} 页</div>
              <Button disabled={!canNext} variant={`outline`} onClick={() => setPage((p) => p + 1)}>
                下一页
              </Button>
              <select
                className={`h-9 rounded border px-2 text-sm bg-background`}
                value={size}
                onChange={(e) => setSize(Number(e.target.value))}
              >
                <option value={10}>10</option>
                <option value={20}>20</option>
                <option value={50}>50</option>
              </select>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
