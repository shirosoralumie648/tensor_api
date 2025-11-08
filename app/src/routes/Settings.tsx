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

const PROVIDERS = ["wechat", "qq", "google", "github"] as const;

type ProfileForm = {
  username: string;
  email: string;
};

export default function Settings() {
  const { t } = useTranslation();
  const dispatch = useDispatch();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState<ProfileForm>({ username: "", email: "" });
  const [bindings, setBindings] = useState<OAuthBinding[]>([]);

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
    const url = `${backendEndpoint}/oauth/${provider}?bind=1&token=${encodeURIComponent(token)}`;
    location.href = url;
  };

  const doUnbind = async (provider: string) => {
    await unbindOAuth(provider);
    await load();
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
    </div>
  );
}
