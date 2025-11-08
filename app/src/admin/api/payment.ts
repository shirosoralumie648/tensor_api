import axios from "axios";
import { CommonResponse } from "@/api/common.ts";

export type YiPayConfig = {
  enabled: boolean;
  endpoint: string;
  mch_id: string;
  key: string;
  notify_url: string;
  return_url: string;
};

export type WechatConfig = {
  enabled: boolean;
  mchid: string;
  appid: string;
  api_v3_key: string;
  serial_no: string;
  private_key_pem: string;
  notify_url: string;
  scene: "jsapi" | "native" | "h5";
};

export type AlipayConfig = {
  enabled: boolean;
  app_id: string;
  private_key: string;
  alipay_public_key: string;
  notify_url: string;
  return_url: string;
};

export type StripeConfig = {
  enabled: boolean;
  secret_key: string;
  webhook_secret: string;
  success_url: string;
  cancel_url: string;
  currency: string;
};

export type AggregateConfig = {
  enabled: boolean;
  vendor: string;
  endpoint: string;
  app_id: string;
  key: string;
  sign_type: string;
};

export type PaymentConfig = {
  yipay: YiPayConfig;
  wechat: WechatConfig;
  alipay: AlipayConfig;
  stripe: StripeConfig;
  aggregate: AggregateConfig;
};

export type PaymentConfigResponse = CommonResponse & { data?: PaymentConfig };

export async function getPaymentConfig(): Promise<PaymentConfigResponse> {
  try {
    const res = await axios.get("/admin/payment/config/view");
    return res.data as PaymentConfigResponse;
  } catch (e: any) {
    return { status: false, error: e?.message } as PaymentConfigResponse;
  }
}

export async function setPaymentConfig(cfg: PaymentConfig): Promise<CommonResponse> {
  try {
    const res = await axios.post("/admin/payment/config/update", cfg);
    return res.data as CommonResponse;
  } catch (e: any) {
    return { status: false, error: e?.message } as CommonResponse;
  }
}

export type PaymentOrder = {
  id: number;
  order_no: string;
  trade_no: string;
  user_id: number;
  gateway: string;
  amount: number;
  currency: string;
  subject: string;
  body: string;
  status: string;
  metadata: string;
  created_at: string;
  updated_at: string;
  paid_at?: string;
  refunded_at?: string;
  expires_at?: string;
};

export type PaymentOrderListResponse = CommonResponse & {
  data?: { list: PaymentOrder[]; total: number };
};

export async function getPaymentOrders(params: {
  status?: string;
  gateway?: string;
  q?: string;
  user_id?: number;
  start?: string;
  end?: string;
  page?: number;
  size?: number;
}): Promise<PaymentOrderListResponse> {
  try {
    const res = await axios.get("/admin/payment/orders", { params });
    return res.data as PaymentOrderListResponse;
  } catch (e: any) {
    return { status: false, error: e?.message } as PaymentOrderListResponse;
  }
}

export async function getPaymentOrder(orderNo: string): Promise<CommonResponse & { data?: PaymentOrder }>{
  try {
    const res = await axios.get(`/admin/payment/order/${orderNo}`);
    return res.data as any;
  } catch (e: any) {
    return { status: false, error: e?.message } as CommonResponse;
  }
}

export async function syncPaymentOrder(orderNo: string): Promise<CommonResponse> {
  try {
    const res = await axios.get(`/admin/payment/order/sync/${orderNo}`);
    return res.data as CommonResponse;
  } catch (e: any) {
    return { status: false, error: e?.message } as CommonResponse;
  }
}

export async function refundPaymentOrder(orderNo: string): Promise<CommonResponse> {
  try {
    const res = await axios.post(`/admin/payment/order/refund/${orderNo}`);
    return res.data as CommonResponse;
  } catch (e: any) {
    return { status: false, error: e?.message } as CommonResponse;
  }
}
