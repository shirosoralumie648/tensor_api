import { useEffect, useState } from "react";
import { useToast } from "@/components/ui/use-toast.ts";
import { toastState } from "@/api/common.ts";
import { Button } from "@/components/ui/button.tsx";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Label } from "@/components/ui/label.tsx";
import { Switch } from "@/components/ui/switch.tsx";
import {
  AggregateConfig,
  AlipayConfig,
  PaymentConfig,
  StripeConfig,
  WechatConfig,
  YiPayConfig,
  getPaymentConfig,
  setPaymentConfig,
} from "@/admin/api/payment.ts";

const initYiPay: YiPayConfig = {
  enabled: false,
  endpoint: "",
  mch_id: "",
  key: "",
  notify_url: "",
  return_url: "",
};

const initWechat: WechatConfig = {
  enabled: false,
  mchid: "",
  appid: "",
  api_v3_key: "",
  serial_no: "",
  private_key_pem: "",
  notify_url: "",
  scene: "native",
};

const initAlipay: AlipayConfig = {
  enabled: false,
  app_id: "",
  private_key: "",
  alipay_public_key: "",
  notify_url: "",
  return_url: "",
};

const initStripe: StripeConfig = {
  enabled: false,
  secret_key: "",
  webhook_secret: "",
  success_url: "",
  cancel_url: "",
  currency: "CNY",
};

const initAggregate: AggregateConfig = {
  enabled: false,
  vendor: "",
  endpoint: "",
  app_id: "",
  key: "",
  sign_type: "md5",
};

const initialConfig: PaymentConfig = {
  yipay: initYiPay,
  wechat: initWechat,
  alipay: initAlipay,
  stripe: initStripe,
  aggregate: initAggregate,
};

export default function Payment() {
  const { toast } = useToast();
  const [cfg, setCfg] = useState<PaymentConfig>(initialConfig);
  const [loading, setLoading] = useState(false);

  const load = async () => {
    setLoading(true);
    const resp = await getPaymentConfig();
    setLoading(false);
    if (!resp.status) return toastState(toast, undefined, resp);
    setCfg({
      yipay: { ...initYiPay, ...(resp.data?.yipay || {}) },
      wechat: { ...initWechat, ...(resp.data?.wechat || {}) },
      alipay: { ...initAlipay, ...(resp.data?.alipay || {}) },
      stripe: { ...initStripe, ...(resp.data?.stripe || {}) },
      aggregate: { ...initAggregate, ...(resp.data?.aggregate || {}) },
    });
  };

  useEffect(() => {
    load();
  }, []);

  const save = async () => {
    setLoading(true);
    const resp = await setPaymentConfig(cfg);
    setLoading(false);
    toastState(toast, undefined, resp, true);
  };

  return (
    <div className={`system`}>
      <div className={`grid grid-cols-1 xl:grid-cols-2 gap-4`}>
        <Card className={`admin-card`}>
          <CardHeader>
            <CardTitle>YiPay</CardTitle>
            <CardDescription>易支付协议</CardDescription>
          </CardHeader>
          <CardContent>
            <div className={`space-y-3`}>
              <div className={`flex items-center justify-between`}>
                <Label>启用</Label>
                <Switch
                  checked={cfg.yipay.enabled}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, yipay: { ...s.yipay, enabled: !!v } }))
                  }
                />
              </div>
              <Label>endpoint</Label>
              <Input
                value={cfg.yipay.endpoint}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, yipay: { ...s.yipay, endpoint: e.target.value } }))
                }
              />
              <Label>mch_id</Label>
              <Input
                value={cfg.yipay.mch_id}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, yipay: { ...s.yipay, mch_id: e.target.value } }))
                }
              />
              <Label>key</Label>
              <Input
                value={cfg.yipay.key}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, yipay: { ...s.yipay, key: e.target.value } }))
                }
              />
              <Label>notify_url</Label>
              <Input
                value={cfg.yipay.notify_url}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, yipay: { ...s.yipay, notify_url: e.target.value } }))
                }
              />
              <Label>return_url</Label>
              <Input
                value={cfg.yipay.return_url}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, yipay: { ...s.yipay, return_url: e.target.value } }))
                }
              />
            </div>
          </CardContent>
        </Card>

        <Card className={`admin-card`}>
          <CardHeader>
            <CardTitle>Stripe</CardTitle>
            <CardDescription>Checkout/Webhook</CardDescription>
          </CardHeader>
          <CardContent>
            <div className={`space-y-3`}>
              <div className={`flex items-center justify-between`}>
                <Label>启用</Label>
                <Switch
                  checked={cfg.stripe.enabled}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, stripe: { ...s.stripe, enabled: !!v } }))
                  }
                />
              </div>
              <Label>secret_key</Label>
              <Input
                value={cfg.stripe.secret_key}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, stripe: { ...s.stripe, secret_key: e.target.value } }))
                }
              />
              <Label>webhook_secret</Label>
              <Input
                value={cfg.stripe.webhook_secret}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, stripe: { ...s.stripe, webhook_secret: e.target.value } }))
                }
              />
              <Label>success_url</Label>
              <Input
                value={cfg.stripe.success_url}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, stripe: { ...s.stripe, success_url: e.target.value } }))
                }
              />
              <Label>cancel_url</Label>
              <Input
                value={cfg.stripe.cancel_url}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, stripe: { ...s.stripe, cancel_url: e.target.value } }))
                }
              />
              <Label>currency</Label>
              <Input
                value={cfg.stripe.currency}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, stripe: { ...s.stripe, currency: e.target.value } }))
                }
              />
            </div>
          </CardContent>
        </Card>

        <Card className={`admin-card`}>
          <CardHeader>
            <CardTitle>Wechat</CardTitle>
            <CardDescription>V3</CardDescription>
          </CardHeader>
          <CardContent>
            <div className={`space-y-3`}>
              <div className={`flex items-center justify-between`}>
                <Label>启用</Label>
                <Switch
                  checked={cfg.wechat.enabled}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, wechat: { ...s.wechat, enabled: !!v } }))
                  }
                />
              </div>
              <Label>mchid</Label>
              <Input
                value={cfg.wechat.mchid}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, wechat: { ...s.wechat, mchid: e.target.value } }))
                }
              />
              <Label>appid</Label>
              <Input
                value={cfg.wechat.appid}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, wechat: { ...s.wechat, appid: e.target.value } }))
                }
              />
              <Label>api_v3_key</Label>
              <Input
                value={cfg.wechat.api_v3_key}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, wechat: { ...s.wechat, api_v3_key: e.target.value } }))
                }
              />
              <Label>serial_no</Label>
              <Input
                value={cfg.wechat.serial_no}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, wechat: { ...s.wechat, serial_no: e.target.value } }))
                }
              />
              <Label>private_key_pem</Label>
              <Input
                value={cfg.wechat.private_key_pem}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, wechat: { ...s.wechat, private_key_pem: e.target.value } }))
                }
              />
              <Label>notify_url</Label>
              <Input
                value={cfg.wechat.notify_url}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, wechat: { ...s.wechat, notify_url: e.target.value } }))
                }
              />
              <Label>scene</Label>
              <Input
                value={cfg.wechat.scene}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, wechat: { ...s.wechat, scene: e.target.value as any } }))
                }
              />
            </div>
          </CardContent>
        </Card>

        <Card className={`admin-card`}>
          <CardHeader>
            <CardTitle>Alipay</CardTitle>
            <CardDescription>RSA2</CardDescription>
          </CardHeader>
          <CardContent>
            <div className={`space-y-3`}>
              <div className={`flex items-center justify-between`}>
                <Label>启用</Label>
                <Switch
                  checked={cfg.alipay.enabled}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, alipay: { ...s.alipay, enabled: !!v } }))
                  }
                />
              </div>
              <Label>app_id</Label>
              <Input
                value={cfg.alipay.app_id}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, alipay: { ...s.alipay, app_id: e.target.value } }))
                }
              />
              <Label>private_key</Label>
              <Input
                value={cfg.alipay.private_key}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, alipay: { ...s.alipay, private_key: e.target.value } }))
                }
              />
              <Label>alipay_public_key</Label>
              <Input
                value={cfg.alipay.alipay_public_key}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, alipay: { ...s.alipay, alipay_public_key: e.target.value } }))
                }
              />
              <Label>notify_url</Label>
              <Input
                value={cfg.alipay.notify_url}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, alipay: { ...s.alipay, notify_url: e.target.value } }))
                }
              />
              <Label>return_url</Label>
              <Input
                value={cfg.alipay.return_url}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, alipay: { ...s.alipay, return_url: e.target.value } }))
                }
              />
            </div>
          </CardContent>
        </Card>

        <Card className={`admin-card xl:col-span-2`}>
          <CardHeader>
            <CardTitle>聚合支付</CardTitle>
            <CardDescription>对接第三方聚合网关</CardDescription>
          </CardHeader>
          <CardContent>
            <div className={`space-y-3`}>
              <div className={`flex items-center justify-between`}>
                <Label>启用</Label>
                <Switch
                  checked={cfg.aggregate.enabled}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, aggregate: { ...s.aggregate, enabled: !!v } }))
                  }
                />
              </div>
              <Label>vendor</Label>
              <Input
                value={cfg.aggregate.vendor}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, aggregate: { ...s.aggregate, vendor: e.target.value } }))
                }
              />
              <Label>endpoint</Label>
              <Input
                value={cfg.aggregate.endpoint}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, aggregate: { ...s.aggregate, endpoint: e.target.value } }))
                }
              />
              <Label>app_id</Label>
              <Input
                value={cfg.aggregate.app_id}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, aggregate: { ...s.aggregate, app_id: e.target.value } }))
                }
              />
              <Label>key</Label>
              <Input
                value={cfg.aggregate.key}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, aggregate: { ...s.aggregate, key: e.target.value } }))
                }
              />
              <Label>sign_type</Label>
              <Input
                value={cfg.aggregate.sign_type}
                onChange={(e) =>
                  setCfg((s) => ({ ...s, aggregate: { ...s.aggregate, sign_type: e.target.value } }))
                }
              />
            </div>
          </CardContent>
        </Card>
      </div>

      <div className={`mt-6 flex flex-row`}>
        <Button className={`mr-2`} onClick={save} loading={!!loading}>
          保存
        </Button>
        <Button variant={`outline`} onClick={load}>
          刷新
        </Button>
      </div>
    </div>
  );
}
