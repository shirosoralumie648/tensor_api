import axios from "axios";
import { CommonResponse } from "@/api/common.ts";

export type ThemeFeature = {
  site_theme: "system" | "light" | "dark";
  enforce: boolean;
};

export type MarkdownFeature = {
  highlight: boolean;
  math: boolean;
  mermaid: boolean;
  chart: boolean;
};

export type FeatureConfig = {
  theme: ThemeFeature;
  markdown: MarkdownFeature;
};

export type FeatureResponse = CommonResponse & {
  data?: FeatureConfig;
};

export async function getFeatureConfig(): Promise<FeatureResponse> {
  try {
    const res = await axios.get("/admin/feature/view");
    return res.data as FeatureResponse;
  } catch (e: any) {
    return { status: false, error: e?.message } as FeatureResponse;
  }
}

export async function setFeatureConfig(cfg: FeatureConfig): Promise<CommonResponse> {
  try {
    const res = await axios.post("/admin/feature/update", cfg);
    return res.data as CommonResponse;
  } catch (e: any) {
    return { status: false, error: e?.message } as CommonResponse;
  }
}

export async function getFeatureInfo(): Promise<FeatureConfig> {
  const res = await axios.get("/feature/info");
  const data = res.data as any;
  if (data && data.data) return data.data as FeatureConfig;
  return data as FeatureConfig;
}
