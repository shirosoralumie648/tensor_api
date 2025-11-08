import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
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
import { Label } from "@/components/ui/label.tsx";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select.tsx";
import { Switch } from "@/components/ui/switch.tsx";
import {
  FeatureConfig,
  getFeatureConfig,
  setFeatureConfig,
} from "@/admin/api/feature.ts";
import { setBooleanMemory, setMemory } from "@/utils/memory.ts";
import { activeTheme } from "@/components/ThemeProvider.tsx";

const initialConfig: FeatureConfig = {
  theme: { site_theme: "system", enforce: false },
  markdown: { highlight: true, math: true, mermaid: true, chart: true },
};

export default function Feature() {
  const { t } = useTranslation();
  const { toast } = useToast();
  const [cfg, setCfg] = useState<FeatureConfig>(initialConfig);
  const [saving, setSaving] = useState(false);

  const load = async () => {
    const resp = await getFeatureConfig();
    if (!resp.status || !resp.data) {
      toastState(toast, t, resp);
      return;
    }
    const data = resp.data as FeatureConfig;
    setCfg({
      theme: {
        site_theme: data.theme?.site_theme || "system",
        enforce: !!data.theme?.enforce,
      },
      markdown: {
        highlight: data.markdown?.highlight ?? true,
        math: data.markdown?.math ?? true,
        mermaid: data.markdown?.mermaid ?? true,
        chart: data.markdown?.chart ?? true,
      },
    });
  };

  useEffect(() => {
    load();
  }, []);

  const applyMemory = (next: FeatureConfig) => {
    setBooleanMemory("feature_md_highlight", !!next.markdown.highlight);
    setBooleanMemory("feature_md_math", !!next.markdown.math);
    setBooleanMemory("feature_md_mermaid", !!next.markdown.mermaid);
    setBooleanMemory("feature_md_chart", !!next.markdown.chart);
    if (next.theme.enforce) {
      setMemory("theme", next.theme.site_theme);
      activeTheme(next.theme.site_theme as any);
    }
  };

  const save = async () => {
    setSaving(true);
    const resp = await setFeatureConfig(cfg);
    setSaving(false);
    toastState(toast, t, resp, true);
    if (resp.status) applyMemory(cfg);
  };

  return (
    <div className={`system`}>
      <Card className={`system-card admin-card`}>
        <CardHeader>
          <CardTitle>{t("admin.feature") || "功能设置"}</CardTitle>
          <CardDescription>
            {t("admin.feature-desc") || "主题与 Markdown 功能开关"}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className={`grid grid-cols-1 md:grid-cols-2 gap-4`}>
            <div className={`flex flex-col`}>
              <Label className={`mb-2`}>{t("theme") || "主题"}</Label>
              <div className={`flex items-center mb-3`}>
                <Label className={`mr-2`}>{t("site-theme") || "站点主题"}</Label>
                <Select
                  value={cfg.theme.site_theme}
                  onValueChange={(v) =>
                    setCfg((s) => ({ ...s, theme: { ...s.theme, site_theme: v as any } }))
                  }
                >
                  <SelectTrigger className={`w-[180px]`}>
                    <SelectValue placeholder={cfg.theme.site_theme} />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value={`system`}>system</SelectItem>
                    <SelectItem value={`light`}>light</SelectItem>
                    <SelectItem value={`dark`}>dark</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className={`flex items-center`}>
                <Label className={`mr-2`}>{t("enforce") || "强制覆盖用户主题"}</Label>
                <Switch
                  checked={cfg.theme.enforce}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, theme: { ...s.theme, enforce: !!v } }))
                  }
                />
              </div>
            </div>

            <div className={`flex flex-col`}>
              <Label className={`mb-2`}>{t("markdown") || "Markdown"}</Label>
              <div className={`flex items-center mb-2`}>
                <Label className={`mr-2`}>highlight</Label>
                <Switch
                  checked={cfg.markdown.highlight}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, markdown: { ...s.markdown, highlight: !!v } }))
                  }
                />
              </div>
              <div className={`flex items-center mb-2`}>
                <Label className={`mr-2`}>math</Label>
                <Switch
                  checked={cfg.markdown.math}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, markdown: { ...s.markdown, math: !!v } }))
                  }
                />
              </div>
              <div className={`flex items-center mb-2`}>
                <Label className={`mr-2`}>mermaid</Label>
                <Switch
                  checked={cfg.markdown.mermaid}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, markdown: { ...s.markdown, mermaid: !!v } }))
                  }
                />
              </div>
              <div className={`flex items-center`}>
                <Label className={`mr-2`}>chart</Label>
                <Switch
                  checked={cfg.markdown.chart}
                  onCheckedChange={(v) =>
                    setCfg((s) => ({ ...s, markdown: { ...s.markdown, chart: !!v } }))
                  }
                />
              </div>
            </div>
          </div>

          <div className={`mt-6 flex flex-row`}>
            <Button className={`mr-2`} onClick={save} loading={!!saving}>
              {t("save")}
            </Button>
            <Button variant={`outline`} onClick={load}>
              {t("refresh")}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
