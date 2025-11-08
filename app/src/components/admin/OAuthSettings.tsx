import { useMemo, useReducer, useRef, useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import Paragraph, { ParagraphDescription, ParagraphFooter, ParagraphItem, ParagraphSpace } from "@/components/Paragraph.tsx";
import { Label } from "@/components/ui/label.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Switch } from "@/components/ui/switch.tsx";
import { Button } from "@/components/ui/button.tsx";
import { toastState } from "@/api/common.ts";
import { useToast } from "@/components/ui/use-toast.ts";
import { cn } from "@/components/ui/lib/utils.ts";
import { initialOAuthConfig, getOAuthConfig, setOAuthConfig, OAuthConfig } from "@/admin/api/oauth.ts";
import { formReducer } from "@/utils/form.ts";

function validateOAuth(cfg: OAuthConfig): string[] {
  const errors: string[] = [];
  const check = (enabled: boolean, id: string, secret: string, name: string) => {
    if (!enabled) return;
    if (!id.trim()) errors.push(`${name} ID 未填写`);
    if (!secret.trim()) errors.push(`${name} Secret 未填写`);
  };
  check(cfg.github.enabled, cfg.github.client_id, cfg.github.client_secret, "GitHub");
  check(cfg.google.enabled, cfg.google.client_id, cfg.google.client_secret, "Google");
  check(cfg.wechat.enabled, cfg.wechat.app_id, cfg.wechat.app_secret, "WeChat");
  check(cfg.qq.enabled, cfg.qq.app_id, cfg.qq.app_secret, "QQ");
  return errors;
}

export default function OAuthSettings() {
  const { t } = useTranslation();
  const { toast } = useToast();

  const [data, dispatch] = useReducer(formReducer<OAuthConfig>(), initialOAuthConfig);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const initialSnapRef = useRef("");
  const errors = useMemo(() => validateOAuth(data), [data]);
  const dirty = useMemo(() => initialSnapRef.current !== JSON.stringify(data), [data]);

  const onRefresh = async () => {
    setLoading(true);
    const res = await getOAuthConfig();
    setLoading(false);
    toastState(toast, t, res);
    if (res.status && res.data) {
      dispatch({ type: "set", value: res.data });
      initialSnapRef.current = JSON.stringify(res.data);
    }
  };

  const onSave = async () => {
    if (errors.length > 0) {
      toast({ title: "配置不合法", description: errors.slice(0, 3).join("；") });
      return;
    }
    setSaving(true);
    const res = await setOAuthConfig(data);
    if (res.status) initialSnapRef.current = JSON.stringify(data);
    toastState(toast, t, res, true);
    setSaving(false);
  };

  // initial load
  useEffect(() => { void onRefresh(); }, []);

  const CallbackTip = ({ provider }: { provider: string }) => {
    const base = `${window.location.protocol}//${window.location.host}`;
    const content = `回调地址示例：${base}/oauth/${provider}/callback（如使用反向代理的 /api 前缀，则为 ${base}/api/oauth/${provider}/callback）`;
    return (
      <ParagraphDescription border>{content}</ParagraphDescription>
    );
  };

  return (
    <Paragraph title={t("admin.system.oauth") || "OAuth 设置"} configParagraph isCollapsed>
      {errors.length > 0 && (
        <div className={cn("border border-red-200 bg-red-50 text-red-700 rounded-md p-3 text-sm mb-2")}> 
          <div className={"font-medium mb-1"}>存在 {errors.length} 处配置问题：</div>
          <ul className={"list-disc pl-5"}>
            {errors.map((e, i) => (<li key={i}>{e}</li>))}
          </ul>
        </div>
      )}

      {/* GitHub */}
      <ParagraphItem>
        <Label className={"flex items-center gap-2"}>
          GitHub 登录
        </Label>
        <div className={"flex flex-col gap-2 w-full"}>
          <div className={"flex items-center gap-3"}>
            <Switch checked={data.github.enabled} onCheckedChange={(value) => dispatch({ type: "update:github.enabled", value })} />
            <span className={"text-sm text-muted-foreground"}>{data.github.enabled ? "已启用" : "未启用"}</span>
          </div>
          <div className={"grid grid-cols-1 md:grid-cols-2 gap-2"}>
            <Input placeholder="Client ID" value={data.github.client_id} onChange={(e) => dispatch({ type: "update:github.client_id", value: e.target.value })} />
            <Input placeholder="Client Secret" value={data.github.client_secret} onChange={(e) => dispatch({ type: "update:github.client_secret", value: e.target.value })} />
          </div>
          <CallbackTip provider="github" />
        </div>
      </ParagraphItem>

      {/* Google */}
      <ParagraphItem>
        <Label className={"flex items-center gap-2"}>
          Google 登录
        </Label>
        <div className={"flex flex-col gap-2 w-full"}>
          <div className={"flex items-center gap-3"}>
            <Switch checked={data.google.enabled} onCheckedChange={(value) => dispatch({ type: "update:google.enabled", value })} />
            <span className={"text-sm text-muted-foreground"}>{data.google.enabled ? "已启用" : "未启用"}</span>
          </div>
          <div className={"grid grid-cols-1 md:grid-cols-2 gap-2"}>
            <Input placeholder="Client ID" value={data.google.client_id} onChange={(e) => dispatch({ type: "update:google.client_id", value: e.target.value })} />
            <Input placeholder="Client Secret" value={data.google.client_secret} onChange={(e) => dispatch({ type: "update:google.client_secret", value: e.target.value })} />
          </div>
          <CallbackTip provider="google" />
        </div>
      </ParagraphItem>

      {/* WeChat */}
      <ParagraphItem>
        <Label className={"flex items-center gap-2"}>
          微信登录
        </Label>
        <div className={"flex flex-col gap-2 w-full"}>
          <div className={"flex items-center gap-3"}>
            <Switch checked={data.wechat.enabled} onCheckedChange={(value) => dispatch({ type: "update:wechat.enabled", value })} />
            <span className={"text-sm text-muted-foreground"}>{data.wechat.enabled ? "已启用" : "未启用"}</span>
          </div>
          <div className={"grid grid-cols-1 md:grid-cols-2 gap-2"}>
            <Input placeholder="App ID" value={data.wechat.app_id} onChange={(e) => dispatch({ type: "update:wechat.app_id", value: e.target.value })} />
            <Input placeholder="App Secret" value={data.wechat.app_secret} onChange={(e) => dispatch({ type: "update:wechat.app_secret", value: e.target.value })} />
          </div>
          <CallbackTip provider="wechat" />
        </div>
      </ParagraphItem>

      {/* QQ */}
      <ParagraphItem>
        <Label className={"flex items-center gap-2"}>
          QQ 登录
        </Label>
        <div className={"flex flex-col gap-2 w-full"}>
          <div className={"flex items-center gap-3"}>
            <Switch checked={data.qq.enabled} onCheckedChange={(value) => dispatch({ type: "update:qq.enabled", value })} />
            <span className={"text-sm text-muted-foreground"}>{data.qq.enabled ? "已启用" : "未启用"}</span>
          </div>
          <div className={"grid grid-cols-1 md:grid-cols-2 gap-2"}>
            <Input placeholder="App ID" value={data.qq.app_id} onChange={(e) => dispatch({ type: "update:qq.app_id", value: e.target.value })} />
            <Input placeholder="App Secret" value={data.qq.app_secret} onChange={(e) => dispatch({ type: "update:qq.app_secret", value: e.target.value })} />
          </div>
          <CallbackTip provider="qq" />
        </div>
      </ParagraphItem>

      <ParagraphSpace />
      <ParagraphFooter>
        <div className={"grow"} />
        <Button size="sm" variant="outline" loading={loading} onClick={onRefresh} className={"mr-2"}>
          {t("admin.refresh") || "刷新"}
        </Button>
        <Button size="sm" loading={saving} onClick={onSave} disabled={!dirty}>
          {t("admin.system.save") || "保存"}
        </Button>
      </ParagraphFooter>
    </Paragraph>
  );
}
