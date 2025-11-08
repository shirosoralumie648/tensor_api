import { useEffect, useState } from "react";
import { useDispatch } from "react-redux";
import { useTranslation } from "react-i18next";
import { Card, CardContent } from "@/components/ui/card.tsx";
import { Label } from "@/components/ui/label.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Button } from "@/components/ui/button.tsx";
import { getProfile, updateProfile, getOAuthBindings, unbindOAuth, OAuthBinding } from "@/api/user.ts";
import { getErrorMessage } from "@/utils/base.ts";
import { validateToken } from "@/store/auth.ts";
import { backendEndpoint } from "@/conf/env.ts";
import { getMemory } from "@/utils/memory.ts";
import { tokenField } from "@/conf/bootstrap.ts";
import { useToast } from "@/components/ui/use-toast.ts";
import { sendCode, doReset } from "@/api/auth.ts";

const PROVIDERS = ["wechat", "qq", "google", "github"] as const;

type ProfileForm = {
  username: string;
  email: string;
};

export default function Settings() {
  const { t } = useTranslation();
  const dispatch = useDispatch();
  const { toast } = useToast();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<ProfileForm>({ username: "", email: "" });
  const [bindings, setBindings] = useState<OAuthBinding[]>([]);
  const [code, setCode] = useState("");
  const [pwd, setPwd] = useState("");
  const [repwd, setRepwd] = useState("");
  const [pwdSaving, setPwdSaving] = useState(false);

  const load = async () => {
    setLoading(true);
    try {
      const [p, b] = await Promise.all([getProfile(), getOAuthBindings()]);
      if (p.status && p.data) setForm({ username: p.data.username || "", email: p.data.email || "" });
      if (b.status && b.data) setBindings(b.data);
    } catch (e) {
      console.debug(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const onSave = async () => {
    setSaving(true);
    try {
      const resp = await updateProfile(form);
      if (!resp.status) return;
      if (resp.token) validateToken(dispatch as any, resp.token);
      await load();
    } catch (e) {
      console.debug(getErrorMessage(e));
    } finally {
      setSaving(false);
    }
  };

  const startBind = (provider: string) => {
    const token = getMemory(tokenField);
    if (!token) return;
    if (!code.trim()) {
      toast({ title: "请先输入邮箱验证码" });
      return;
    }
    const url = `${backendEndpoint}/oauth/${provider}?bind=1&token=${encodeURIComponent(token)}&code=${encodeURIComponent(code)}`;
    location.href = url;
  };

  const doUnbind = async (provider: string) => {
    if (!code.trim()) {
      toast({ title: "请先输入邮箱验证码" });
      return;
    }
    await unbindOAuth(provider, code.trim());
    await load();
  };

  const onSendCode = async () => {
    if (!form.email.trim()) {
      toast({ title: "请先填写邮箱" });
      return;
    }
    await sendCode(t as any, toast as any, form.email);
  };

  const onChangePassword = async () => {
    if (!code.trim()) {
      toast({ title: "请先输入邮箱验证码" });
      return;
    }
    if (pwd.trim().length < 6 || pwd.trim().length > 36) {
      toast({ title: "密码格式不正确(6~36位)" });
      return;
    }
    if (pwd.trim() !== repwd.trim()) {
      toast({ title: "两次输入的密码不一致" });
      return;
    }
    try {
      setPwdSaving(true);
      const res = await doReset({ email: form.email, code: code.trim(), password: pwd.trim(), repassword: repwd.trim() });
      if (!res.status) {
        toast({ title: "修改失败", description: res.error || "" });
      } else {
        toast({ title: "修改成功" });
        setPwd("");
        setRepwd("");
      }
    } finally {
      setPwdSaving(false);
    }
  };

  const boundProviders = new Set(bindings.map((b) => b.provider));

  return (
    <div className={`container mx-auto max-w-3xl px-4 py-6`}>
      <div className={`text-2xl font-semibold mb-4`}>{t("settings.title") || "用户设置"}</div>
      <Card className={`mb-6`}>
        <CardContent className={`space-y-3 pt-6`}>
          <div>
            <Label>用户名</Label>
            <Input
              value={form.username}
              onChange={(e) => setForm({ ...form, username: e.target.value })}
              placeholder={`长度 2~24`}
              disabled={loading}
            />
          </div>
          <div>
            <Label>邮箱</Label>
            <Input
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              placeholder={`example@domain.com`}
              disabled={loading}
            />
          </div>
          <div>
            <Label>邮箱验证码</Label>
            <div className={`flex gap-2`}>
              <Input value={code} onChange={(e) => setCode(e.target.value)} placeholder={`输入验证码`} />
              <Button variant={`outline`} onClick={onSendCode}>发送验证码</Button>
            </div>
          </div>
          <div>
            <Button onClick={onSave} loading={saving}>
              保存
            </Button>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardContent className={`pt-6`}>
          <div className={`text-lg mb-3`}>第三方账号绑定</div>
          <div className={`grid grid-cols-2 gap-3`}>
            {PROVIDERS.map((p) => (
              <div key={p} className={`flex items-center justify-between p-2 rounded border`}> 
                <div className={`capitalize`}>{p}</div>
                {boundProviders.has(p) ? (
                  <Button variant={`destructive`} onClick={() => doUnbind(p)}>
                    解绑
                  </Button>
                ) : (
                  <Button variant={`outline`} onClick={() => startBind(p)}>
                    绑定
                  </Button>
                )}
              </div>
            ))}
          </div>
        </CardContent>
      </Card>

      <Card className={`mt-6`}>
        <CardContent className={`pt-6 space-y-3`}>
          <div className={`text-lg`}>修改密码（需要邮箱验证码）</div>
          <div>
            <Label>新密码</Label>
            <Input type={`password`} value={pwd} onChange={(e) => setPwd(e.target.value)} placeholder={`6~36位`} />
          </div>
          <div>
            <Label>确认新密码</Label>
            <Input type={`password`} value={repwd} onChange={(e) => setRepwd(e.target.value)} placeholder={`再次输入`} />
          </div>
          <div>
            <Button onClick={onChangePassword} loading={pwdSaving}>确认修改</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
